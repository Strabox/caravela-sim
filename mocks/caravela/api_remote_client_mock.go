package caravela

import (
	"github.com/strabox/caravela-sim/simulation"
	"github.com/strabox/caravela/api/types"
	"github.com/strabox/caravela/configuration"
)

// RemoteClientMock mocks the RPC from a node to another via the simulator.
type RemoteClientMock struct {
	sim simulation.Simulator
}

func NewRemoteClientMock(sim simulation.Simulator) *RemoteClientMock {
	return &RemoteClientMock{
		sim: sim,
	}
}

/*
===============================================================================
=                      Caravela Remote Client Interface                       =
===============================================================================
*/

func (mock *RemoteClientMock) CreateOffer(fromSupp *types.Node, toTrader *types.Node,
	offer *types.Offer) error {

	node, _ := mock.sim.NodeByGUID(toTrader.GUID)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	node.CreateOffer(fromSupp, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) RefreshOffer(fromTrader, toSupp *types.Node, offer *types.Offer) (bool, error) {
	node, _ := mock.sim.NodeByIP(toSupp.IP)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	return node.RefreshOffer(fromTrader, offer), nil
}

func (mock *RemoteClientMock) RemoveOffer(fromSupp, toTrader *types.Node, offer *types.Offer) error {
	node, _ := mock.sim.NodeByGUID(toTrader.GUID)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	node.RemoveOffer(fromSupp, toTrader, offer)
	return nil
}

func (mock *RemoteClientMock) GetOffers(fromNode, toTrader *types.Node, relay bool) ([]types.AvailableOffer, error) {
	node, _ := mock.sim.NodeByGUID(toTrader.GUID)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	offers := node.GetOffers(fromNode, toTrader, relay)
	return offers, nil
}

func (mock *RemoteClientMock) AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering *types.Node) error {
	node, _ := mock.sim.NodeByGUID(toNeighborTrader.GUID)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	node.AdvertiseOffersNeighbor(fromTrader, toNeighborTrader, traderOffering)
	return nil
}

func (mock *RemoteClientMock) LaunchContainer(fromBuyer, toSupplier *types.Node, offer *types.Offer,
	containerConfig *types.ContainerConfig) (*types.ContainerStatus, error) {
	node, _ := mock.sim.NodeByIP(toSupplier.IP)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	return node.LaunchContainers(fromBuyer, offer, containerConfig)
}

func (mock *RemoteClientMock) StopLocalContainer(toSupplier *types.Node, containerID string) error {
	node, _ := mock.sim.NodeByIP(toSupplier.IP)
	mock.sim.Metrics().MsgsTradedActiveRequest(1)

	return node.StopLocalContainer(containerID)
}

func (mock *RemoteClientMock) ObtainConfiguration(systemsNode *types.Node) (*configuration.Configuration, error) {
	// Do Nothing (For now not necessary for the simulation)
	return nil, nil
}
