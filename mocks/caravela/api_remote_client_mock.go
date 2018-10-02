package caravela

import (
	"context"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
	"unsafe"
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

	// Collect Metrics
	messageSize := sizeofNode(fromSupp) + sizeofNode(toTrader) + sizeofOffer(offer)
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.CreateOffer(ctx, fromSupp, toTrader, offer)
	return nil
}

func (r *RemoteClientMock) RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toSupp.IP)

	// Collect Metrics
	messageSize := sizeofNode(fromTrader) + sizeofNode(toSupp) + sizeofOffer(offer)
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	return toNode.RefreshOffer(ctx, fromTrader, offer), nil
}

func (r *RemoteClientMock) UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)

	// Collect Metrics
	messageSize := sizeofNode(fromSupplier) + sizeofNode(toTrader) + sizeofOffer(offer)
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.UpdateOffer(ctx, fromSupplier, toTrader, offer)
	return nil
}

func (r *RemoteClientMock) RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)

	// Collect Metrics
	messageSize := sizeofNode(fromSupp) + sizeofNode(toTrader) + sizeofOffer(offer)
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.RemoveOffer(ctx, fromSupp, toTrader, offer)
	return nil
}

func (r *RemoteClientMock) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toTrader.IP)

	offers := toNode.GetOffers(ctx, fromNode, toTrader, relay)

	// Collect Metrics
	messageSize := sizeofNode(fromNode) + sizeofNode(toTrader) + int(unsafe.Sizeof(relay))
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))
	r.collector.IncrMessagesTradedRequest(types.RequestID(ctx), 1)
	if !relay {
		r.collector.GetOfferRelayed(1)
	}
	if len(offers) == 0 {
		r.collector.EmptyGetOfferMessage(1)
	}
	return offers, nil
}

func (r *RemoteClientMock) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toNeighborTrader.IP)

	// Collect Metrics
	messageSize := sizeofNode(fromTrader) + sizeofNode(toNeighborTrader) + sizeofNode(traderOffering)
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))

	toNode.AdvertiseOffersNeighbor(ctx, fromTrader, toNeighborTrader, traderOffering)
	return nil
}

func (r *RemoteClientMock) LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	toNode, toNodeIndex := r.nodeService.NodeByIP(toSupplier.IP)

	// Collect Metrics
	messageSize := sizeofNode(fromBuyer) + sizeofNode(toSupplier) + sizeofOffer(offer) + sizeofContainerConfigSlice(containersConfigs)
	r.collector.MessageReceived(toNodeIndex, 1, int64(messageSize))
	r.collector.IncrMessagesTradedRequest(types.RequestID(ctx), 1)

	return toNode.LaunchContainers(ctx, fromBuyer, offer, containersConfigs)
}

func (r *RemoteClientMock) StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error {
	node, nodeIndex := r.nodeService.NodeByIP(toSupplier.IP)

	// Collect Metrics
	messageSize := sizeofNode(toSupplier) + sizeofString(containerID)
	r.collector.MessageReceived(nodeIndex, 1, int64(messageSize))

	return node.StopLocalContainer(ctx, containerID)
}

func (r *RemoteClientMock) ObtainConfiguration(_ context.Context, _ *types.Node) (*configuration.Configuration, error) {
	// Do Nothing (Not necessary for the engine)
	return nil, nil
}
