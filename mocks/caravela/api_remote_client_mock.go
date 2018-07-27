package caravela

import (
	"github.com/strabox/caravela-sim/simulation/metrics"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
)

// RemoteClientMock mocks the remote calls from a node to another via the simulator.
type RemoteClientMock struct {
	nodeService NodeService        // Obtains nodes to send messages
	collector   *metrics.Collector // Collects metrics
}

// NewRemoteClientMock creates a new mock for the inter-node interactions.
// It implements the github.com/strabox/caravela/node/external Caravela interface.
func NewRemoteClientMock(nodeService NodeService, metricsCollector *metrics.Collector) *RemoteClientMock {
	return &RemoteClientMock{
		nodeService: nodeService,
		collector:   metricsCollector,
	}
}

// ===============================================================================
// =                      CARAVELA's Remote Client Interface                     =
// ===============================================================================

func (mock *RemoteClientMock) CreateOffer(fromSupp *types.Node, toTrader *types.Node,
	offer *types.Offer) error {

	node, nodeIndex := mock.nodeService.NodeByGUID(toTrader.GUID)
	mock.collector.APIRequestReceived(nodeIndex)

	node.CreateOffer(fromSupp, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) RefreshOffer(fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	node, nodeIndex := mock.nodeService.NodeByIP(toSupp.IP)
	mock.collector.APIRequestReceived(nodeIndex)

	return node.RefreshOffer(fromTrader, offer), nil
}

func (mock *RemoteClientMock) RemoveOffer(fromSupp, toTrader *types.Node, offer *types.Offer) error {
	node, nodeIndex := mock.nodeService.NodeByGUID(toTrader.GUID)
	mock.collector.APIRequestReceived(nodeIndex)

	node.RemoveOffer(fromSupp, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) GetOffers(fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	node, nodeIndex := mock.nodeService.NodeByGUID(toTrader.GUID)
	mock.collector.MsgsTradedActiveRequest(1)
	mock.collector.APIRequestReceived(nodeIndex)

	offers := node.GetOffers(fromNode, toTrader, relay)
	return offers, nil
}

func (mock *RemoteClientMock) AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	node, nodeIndex := mock.nodeService.NodeByGUID(toNeighborTrader.GUID)
	mock.collector.APIRequestReceived(nodeIndex)

	node.AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering)
	return nil
}

func (mock *RemoteClientMock) LaunchContainer(fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containerConfig *types.ContainerConfig) (*types.ContainerStatus, error) {
	node, nodeIndex := mock.nodeService.NodeByIP(toSupplier.IP)
	mock.collector.MsgsTradedActiveRequest(1)
	mock.collector.APIRequestReceived(nodeIndex)

	return node.LaunchContainers(fromBuyer, offer, containerConfig)
}

func (mock *RemoteClientMock) StopLocalContainer(toSupplier *types.Node, containerID string) error {
	node, nodeIndex := mock.nodeService.NodeByIP(toSupplier.IP)
	mock.collector.APIRequestReceived(nodeIndex)

	return node.StopLocalContainer(containerID)
}

func (mock *RemoteClientMock) ObtainConfiguration(systemsNode *types.Node) (*configuration.Configuration, error) {
	// Do Nothing (Not necessary for the simulation)
	return nil, nil
}
