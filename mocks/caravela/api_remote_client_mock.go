package caravela

import (
	"context"
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

func (mock *RemoteClientMock) CreateOffer(ctx context.Context, fromSupp *types.Node, toTrader *types.Node,
	offer *types.Offer) error {

	node, nodeIndex := mock.nodeService.NodeByIP(toTrader.IP)

	// Collect Metrics
	mock.collector.APIRequestReceived(nodeIndex)

	node.CreateOffer(ctx, fromSupp, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) RefreshOffer(ctx context.Context, fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	node, nodeIndex := mock.nodeService.NodeByIP(toSupp.IP)

	// Collect Metrics
	mock.collector.APIRequestReceived(nodeIndex)

	return node.RefreshOffer(ctx, fromTrader, offer), nil
}

func (mock *RemoteClientMock) UpdateOffer(ctx context.Context, fromSupplier, toTrader *types.Node, offer *types.Offer) error {
	node, nodeIndex := mock.nodeService.NodeByIP(toTrader.IP)

	// Collect Metrics
	mock.collector.APIRequestReceived(nodeIndex)

	node.UpdateOffer(ctx, fromSupplier, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) RemoveOffer(ctx context.Context, fromSupp, toTrader *types.Node, offer *types.Offer) error {
	node, nodeIndex := mock.nodeService.NodeByIP(toTrader.IP)

	// Collect Metrics
	mock.collector.APIRequestReceived(nodeIndex)

	node.RemoveOffer(ctx, fromSupp, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) GetOffers(ctx context.Context, fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	node, nodeIndex := mock.nodeService.NodeByIP(toTrader.IP)

	offers := node.GetOffers(ctx, fromNode, toTrader, relay)

	// Collect Metrics
	mock.collector.IncrMessagesTradedRequest(types.RequestID(ctx), 1)
	mock.collector.APIRequestReceived(nodeIndex)
	if !relay {
		mock.collector.GetOfferRelayed(1)
	}
	if len(offers) == 0 {
		mock.collector.EmptyGetOfferMessage(1)
	}
	return offers, nil
}

func (mock *RemoteClientMock) AdvertiseOffersNeighbor(ctx context.Context, fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	node, nodeIndex := mock.nodeService.NodeByIP(toNeighborTrader.IP)

	// Collect Metrics
	mock.collector.APIRequestReceived(nodeIndex)

	node.AdvertiseOffersNeighbor(ctx, fromTrader, toNeighborTrader, traderOffering)
	return nil
}

func (mock *RemoteClientMock) LaunchContainer(ctx context.Context, fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containersConfigs []types.ContainerConfig) ([]types.ContainerStatus, error) {
	node, nodeIndex := mock.nodeService.NodeByIP(toSupplier.IP)

	// Collect Metrics
	mock.collector.IncrMessagesTradedRequest(types.RequestID(ctx), 1)
	mock.collector.APIRequestReceived(nodeIndex)

	return node.LaunchContainers(ctx, fromBuyer, offer, containersConfigs)
}

func (mock *RemoteClientMock) StopLocalContainer(ctx context.Context, toSupplier *types.Node, containerID string) error {
	node, nodeIndex := mock.nodeService.NodeByIP(toSupplier.IP)

	// Collect Metrics
	mock.collector.APIRequestReceived(nodeIndex)

	return node.StopLocalContainer(ctx, containerID)
}

func (mock *RemoteClientMock) ObtainConfiguration(_ context.Context, _ *types.Node) (*configuration.Configuration, error) {
	// Do Nothing (Not necessary for the simulation)
	return nil, nil
}
