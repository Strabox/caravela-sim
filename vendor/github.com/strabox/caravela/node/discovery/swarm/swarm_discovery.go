package swarm

import (
	"context"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common"
	"github.com/strabox/caravela/node/common/guid"
	"github.com/strabox/caravela/node/common/resources"
	"github.com/strabox/caravela/node/discovery/backend"
	"github.com/strabox/caravela/node/external"
	"sync"
	"time"
)

// Zero GUID is the master's GUID (used in simulation only).
const mastersNodeGUID = 0

// Discovery backend is based on a master-slave cluster architecture (Centralized) that simulates the Docker Swarm.
// It is implemented on top of a Chord overlay because it suits better our prototype framework.
// It is NOT DESIGNED to be used in REAL DEPLOYMENT, we only use it in Simulation to compare with our discovery backends.
type Discovery struct {
	common.NodeComponent // Base component.

	// Common fields
	config       *configuration.Configuration // System's configurations.
	overlay      external.Overlay             // Overlay component.
	client       external.Caravela            // Remote caravela's client.
	nodeGUID     *guid.GUID                   // Node's own GUID.
	isMasterNode bool                         // True: if the node is the master, False: if it is a regular peer.

	// Master node fields
	clusterNodesByGUID sync.Map // Map the node's IP with the node's structure.
	clusterNodes       []*node  // One node's structure per node in the cluster.

	// Peer node fields
	refreshTicker      <-chan time.Time     //
	containersRunning  int                  //
	maximumResources   *resources.Resources //
	availableResources *resources.Resources //
	resourcesMutex     sync.Mutex           //
}

// NewSwarmResourcesDiscovery creates a resource discovery backend based on the Docker Swarm.
func NewSwarmResourcesDiscovery(config *configuration.Configuration, overlay external.Overlay,
	client external.Caravela, _ *resources.Mapping, maxResources resources.Resources) (backend.Discovery, error) {

	return &Discovery{
		config:       config,
		overlay:      overlay,
		client:       client,
		nodeGUID:     nil,
		isMasterNode: false,

		clusterNodesByGUID: sync.Map{},
		clusterNodes:       make([]*node, 0),

		refreshTicker:      time.NewTicker(config.RefreshingInterval()).C,
		containersRunning:  0,
		maximumResources:   maxResources.Copy(),
		availableResources: maxResources.Copy(),
		resourcesMutex:     sync.Mutex{},
	}, nil
}

// start starts the discovery backend in the node.
func (d *Discovery) start() {
	if !d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()
		usedResources := d.usedResources()
		d.client.CreateOffer(
			context.Background(),
			&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
			&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
			&types.Offer{
				FreeResources: types.Resources{
					CPUClass: types.CPUClass(d.availableResources.CPUClass()),
					CPUs:     d.availableResources.CPUs(),
					RAM:      d.availableResources.RAM(),
				},
				UsedResources: types.Resources{
					CPUClass: types.CPUClass(usedResources.CPUClass()),
					CPUs:     usedResources.CPUs(),
					RAM:      usedResources.RAM(),
				},
			},
		)
	}

	if !d.config.Simulation() && !d.isMasterNode {
		go func() {
			for {
				select {
				case <-d.refreshTicker:
					masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()

					d.client.RefreshOffer(
						context.Background(),
						&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
						&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
						&types.Offer{ /* Nothing (only used to simulate real world refreshes in swarm) */ },
					)
				}
			}
		}()
	}
}

// usedResources returns the amount of used resources in this node (if it is not the master).
func (d *Discovery) usedResources() *resources.Resources {
	usedResources := d.maximumResources.Copy()
	usedResources.Sub(*d.availableResources)
	return usedResources
}

// getMasterNodeIDs returns the IP and GUID of the master node.
func (d *Discovery) getMasterNodeIDs() (string, string) {
	nodes, _ := d.overlay.Lookup(
		context.Background(),
		guid.NewGUIDInteger(mastersNodeGUID).Bytes(), // Master's node has GUID 0 (in simulator).
	)

	masterNode := nodes[0]
	return masterNode.IP(), guid.NewGUIDBytes(masterNode.GUID()).String()
}

// ====================== Local Services (Consumed by other Components) ============================

func (d *Discovery) AddTrader(traderGUID guid.GUID) {
	d.nodeGUID = guid.NewGUIDBytes(traderGUID.Bytes())
	d.isMasterNode = d.nodeGUID.Equals(*guid.NewGUIDInteger(mastersNodeGUID))
}

func (d *Discovery) FindOffers(_ context.Context, targetResources resources.Resources) []types.AvailableOffer {
	if d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		resultOffers := make([]types.AvailableOffer, 0)

		// If the resource combination is not valid we will refuse the request.
		if !targetResources.IsValid() {
			return resultOffers
		}

		for _, clusterNode := range d.clusterNodes {
			// Skip nodes that are smaller than the requested resources.
			if !clusterNode.freeResources.Contains(targetResources) {
				continue
			}

			resultOffers = append(resultOffers, types.AvailableOffer{
				SupplierIP: clusterNode.ip(),
				Offer: types.Offer{
					FreeResources: types.Resources{
						CPUClass: types.CPUClass(clusterNode.freeResources.CPUClass()),
						CPUs:     clusterNode.freeResources.CPUs(),
						RAM:      clusterNode.freeResources.RAM(),
					},
					UsedResources: types.Resources{
						CPUClass: types.CPUClass(clusterNode.usedResources.CPUClass()),
						CPUs:     clusterNode.usedResources.CPUs(),
						RAM:      clusterNode.usedResources.RAM(),
					},
				},
			})
		}

		return resultOffers
	}
	return make([]types.AvailableOffer, 0)
}

func (d *Discovery) ObtainResources(_ int64, resourcesNecessary resources.Resources) bool {
	if !d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		if d.availableResources.Contains(resourcesNecessary) {
			d.availableResources.Sub(resourcesNecessary)
			d.containersRunning++

			masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()
			usedResources := d.usedResources()
			// Update the resources offered in the master.
			d.client.UpdateOffer(
				context.Background(),
				&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
				&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
				&types.Offer{
					Amount: d.containersRunning,
					FreeResources: types.Resources{
						CPUClass: types.CPUClass(d.availableResources.CPUClass()),
						CPUs:     d.availableResources.CPUs(),
						RAM:      d.availableResources.RAM(),
					},
					UsedResources: types.Resources{
						CPUClass: types.CPUClass(usedResources.CPUClass()),
						CPUs:     usedResources.CPUs(),
						RAM:      usedResources.RAM(),
					},
				},
			)
			return true
		}
		return false
	}
	return false
}

func (d *Discovery) ReturnResources(releasedResources resources.Resources) {
	if !d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		d.availableResources.Add(releasedResources)
		d.containersRunning--

		masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()
		usedResources := d.usedResources()
		d.client.UpdateOffer( // Update the resources offered in the master.
			context.Background(),
			&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
			&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
			&types.Offer{
				Amount: d.containersRunning,
				FreeResources: types.Resources{
					CPUClass: types.CPUClass(d.availableResources.CPUClass()),
					CPUs:     d.availableResources.CPUs(),
					RAM:      d.availableResources.RAM(),
				},
				UsedResources: types.Resources{
					CPUClass: types.CPUClass(usedResources.CPUClass()),
					CPUs:     usedResources.CPUs(),
					RAM:      usedResources.RAM(),
				},
			},
		)
	}
}

// ======================= External Services (Consumed by other Nodes) ==============================

func (d *Discovery) CreateOffer(fromSupp *types.Node, _ *types.Node, offer *types.Offer) {
	if d.isMasterNode {
		d.resourcesMutex.Lock()
		defer d.resourcesMutex.Unlock()

		availableResources := *resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.RAM)
		usedResources := *resources.NewResourcesCPUClass(int(offer.UsedResources.CPUClass), offer.UsedResources.CPUs, offer.UsedResources.RAM)
		clusterNode := newNode(fromSupp.IP, availableResources, usedResources)

		d.clusterNodes = append(d.clusterNodes, clusterNode)
		d.clusterNodesByGUID.Store(fromSupp.GUID, clusterNode)
	}
}

func (d *Discovery) RefreshOffer(_ *types.Node, _ *types.Offer) bool {
	return true
}

func (d *Discovery) UpdateOffer(fromSupp, _ *types.Node, offer *types.Offer) {
	if d.isMasterNode {
		if nodeStored, exist := d.clusterNodesByGUID.Load(fromSupp.GUID); exist {
			if nodePtr, ok := nodeStored.(*node); ok {
				nodeFreeUpdatedRes := *resources.NewResourcesCPUClass(int(offer.FreeResources.CPUClass), offer.FreeResources.CPUs, offer.FreeResources.RAM)
				nodeUsedUpdatedRes := *resources.NewResourcesCPUClass(int(offer.UsedResources.CPUClass), offer.UsedResources.CPUs, offer.UsedResources.RAM)

				nodePtr.setFreeResources(nodeFreeUpdatedRes)
				nodePtr.setUsedResources(nodeUsedUpdatedRes)
				nodePtr.setContainerRunning(offer.Amount) // HACK: Careful if we use stack deployments!
			}
		}
	}
}

func (d *Discovery) RemoveOffer(_ *types.Node, _ *types.Node, _ *types.Offer) {
	// Do Nothing - Not necessary for this backend.
}

func (d *Discovery) GetOffers(_ context.Context, _, _ *types.Node, _ bool) []types.AvailableOffer {
	// Do Nothing - Not necessary for this backend.
	return nil
}

func (d *Discovery) AdvertiseNeighborOffers(_, _, _ *types.Node) {
	// Do Nothing - Not necessary for this backend.
}

// ======================= External Services (Consumed during simulation ONLY) =========================

// Simulation
func (d *Discovery) AvailableResourcesSim() types.Resources {
	d.resourcesMutex.Lock()
	defer d.resourcesMutex.Unlock()

	return types.Resources{
		CPUClass: types.CPUClass(d.availableResources.CPUClass()),
		CPUs:     d.availableResources.CPUs(),
		RAM:      d.availableResources.RAM(),
	}
}

// Simulation
func (d *Discovery) MaximumResourcesSim() types.Resources {
	return types.Resources{
		CPUClass: types.CPUClass(d.maximumResources.CPUClass()),
		CPUs:     d.maximumResources.CPUs(),
		RAM:      d.maximumResources.RAM(),
	}
}

// Simulation
func (d *Discovery) RefreshOffersSim() {
	if !d.isMasterNode {
		masterNodeIP, masterNodeGUID := d.getMasterNodeIDs()

		d.client.RefreshOffer(
			context.Background(),
			&types.Node{IP: d.config.HostIP(), GUID: d.nodeGUID.String()},
			&types.Node{IP: masterNodeIP, GUID: masterNodeGUID},
			&types.Offer{ /* Nothing (only used to simulate real world refreshes in swarm) */ },
		)
	}
}

// Simulation
func (d *Discovery) SpreadOffersSim() {
	// Do Nothing - Not necessary for this backend.
}

// ===============================================================================
// =							SubComponent Interface                           =
// ===============================================================================

func (d *Discovery) Start() {
	d.Started(d.config.Simulation(), func() {
		d.start()
	})
}

func (d *Discovery) Stop() {
	d.Stopped(func() {
		// Do Nothing
	})
}

func (d *Discovery) IsWorking() bool {
	return d.Working()
}
