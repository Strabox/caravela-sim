package caravela

import (
	"context"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela/api/rest/util"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
)

// RemoteClientMock mocks the remote calls from a node to another via the simulator.
type RemoteClientMock struct {
	nodeService simNodeService     // Obtains nodes to send messages
	collector   *metrics.Collector // Collects metrics
}

// NewRemoteClientMock creates a new mock for the inter-node interactions.
// It implements the github.com/strabox/caravela/node/external Caravela interface.
func NewRemoteClientMock(nodeService simNodeService, metricsCollector *metrics.Collector) *RemoteClientMock {
	return &RemoteClientMock{
		nodeService: nodeService,
		collector:   metricsCollector,
	}
}

// ===============================================================================
// =                      CARAVELA's Remote Client Interface                     =
// ===============================================================================

func (r *RemoteClientMock) CreateOffer(ctx context.Context, fromSupp *types.Node, toTrader *types.Node, offer *types.Offer) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	messageSize := sizeofCreateOfferMessage(&util.CreateOfferMsg{ToNode: *toTrader, FromNode: *fromSupp, Offer: *offer})
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.CreateOffer(ctx, fromSupp, toTrader, offer)

	// Collect Metrics (fromNode)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(8))

	return nil
}

func (r *RemoteClientMock) RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toSupp.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	toMessageSize := sizeofRefreshOfferMessage(&util.RefreshOfferMsg{FromTrader: *fromTrader, Offer: *offer})
	r.collector.MessageReceived(toNodeIndex, 1, int64(toMessageSize))

	response := toNode.RefreshOffer(ctx, fromTrader, offer)

	// Collect Metrics (fromNode)
	fromMessageSize := sizeofRefreshOfferMessageResponse(&util.RefreshOfferResponseMsg{Refreshed: response})
	r.collector.MessageReceived(fromNodeIndex, 1, int64(fromMessageSize))

	return response, nil
}

func (r *RemoteClientMock) UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	messageSize := sizeofUpdateOfferMessage(&util.UpdateOfferMsg{FromSupplier: *fromSupplier, ToTrader: *toTrader, Offer: *offer})
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.UpdateOffer(ctx, fromSupplier, toTrader, offer)

	// Collect Metrics (fromNode)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(8))

	return nil
}

func (r *RemoteClientMock) RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	messageSize := sizeofRemoveOfferMessage(&util.OfferRemoveMsg{FromSupplier: *fromSupp, ToTrader: *toTrader, Offer: *offer})
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.RemoveOffer(ctx, fromSupp, toTrader, offer)

	// Collect Metrics (fromNode)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(8))

	return nil
}

func (r *RemoteClientMock) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	toMessageSize := sizeofGetOffersMessage(&util.GetOffersMsg{FromNode: *fromNode, ToTrader: *toTrader, Relay: relay})
	r.collector.MessageReceived(toNodeIndex, 1, int64(toMessageSize))
	r.collector.IncrMessagesTradedRequest(types.RequestID(ctx), 1)

	offers := toNode.GetOffers(ctx, fromNode, toTrader, relay)

	// Collect Metrics (fromNode)
	fromMessageSize := sizeofAvailableOffersMessage(offers)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(fromMessageSize))

	return offers, nil
}

func (r *RemoteClientMock) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toNeighborTrader.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	messageSize := sizeofNeighborOfferMessage(&util.NeighborOffersMsg{FromNeighbor: *fromTrader, ToNeighbor: *toNeighborTrader, NeighborOffering: *traderOffering})
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.AdvertiseOffersNeighbor(ctx, fromTrader, toNeighborTrader, traderOffering)

	// Collect Metrics (fromNode)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(8))

	return nil
}

func (r *RemoteClientMock) LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toSupplier.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	toMessageSize := sizeofLaunchContainerMessage(&util.LaunchContainerMsg{FromBuyer: *fromBuyer, Offer: *offer, ContainersConfigs: containersConfigs})
	r.collector.MessageReceived(toNodeIndex, 1, int64(toMessageSize))
	r.collector.IncrMessagesTradedRequest(types.RequestID(ctx), 1)

	containersStatus, requestErr := toNode.LaunchContainers(ctx, fromBuyer, offer, containersConfigs)

	// Collect Metrics (fromNode)
	fromMessageSize := sizeofContainersStatusMessage(containersStatus)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(fromMessageSize))

	return containersStatus, requestErr
}

func (r *RemoteClientMock) StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error {
	node, nodeIndex := r.nodeService.NodeByIP(toSupplier.IP)
	_, fromNodeIndex := r.nodeService.NodeByGUID(types.NodeGUID(ctx))

	// Collect Metrics (toNode)
	messageSize := sizeofStopLocalContainerMessage(&util.StopLocalContainerMsg{ContainerID: containerID})
	r.collector.MessageReceived(nodeIndex, 1, int64(messageSize))

	requestErr := node.StopLocalContainer(ctx, containerID)

	// Collect Metrics (fromNode)
	r.collector.MessageReceived(fromNodeIndex, 1, int64(8))

	return requestErr
}

func (r *RemoteClientMock) ObtainConfiguration(_ context.Context, _ *types.Node) (*configuration.Configuration, error) {
	// Do Nothing (Not necessary for the engine)
	return nil, nil
}
