package caravela

import (
	"github.com/strabox/caravela-sim/simulation"
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
)

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
                      Caravela Remote Client Interface
===============================================================================
*/

func (mock *RemoteClientMock) CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64, amount int, cpus int, ram int) error {

	mock.sim.NodeByGUID(toTraderGUID).CreateOffer(fromSupplierGUID, fromSupplierIP, toTraderGUID,
		offerID, amount, cpus, ram)
	return nil
}

func (mock *RemoteClientMock) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (bool, error) {
	return mock.sim.NodeByIP(toSupplierIP).RefreshOffer(offerID, fromTraderGUID), nil
}

func (mock *RemoteClientMock) RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string,
	toTraderGUID string, offerID int64) error {

	mock.sim.NodeByGUID(toTraderGUID).RemoveOffer(fromSupplierIP, fromSupplierGUID, toTraderGUID, offerID)
	return nil
}

func (mock *RemoteClientMock) GetOffers(toTraderIP string, toTraderGUID string, relay bool,
	fromNodeGUID string) ([]nodeAPI.Offer, error) {

	offers := mock.sim.NodeByGUID(toTraderGUID).GetOffers(toTraderGUID, relay, fromNodeGUID)
	return offers, nil
}

func (mock *RemoteClientMock) AdvertiseOffersNeighbor(toNeighborTraderIP string, toNeighborTraderGUID string,
	fromTraderGUID string, traderOfferingGUID string, traderOfferingIP string) error {

	mock.sim.NodeByGUID(toNeighborTraderGUID).AdvertiseNeighborOffers(toNeighborTraderGUID, fromTraderGUID,
		traderOfferingIP, traderOfferingGUID)
	return nil
}

func (mock *RemoteClientMock) LaunchContainer(toSupplierIP string, fromBuyerIP string, offerID int64, containerImageKey string,
	portMappings []rest.PortMapping, containerArgs []string, cpus int, ram int) (*rest.ContainerStatus, error) {

	containerID, err := mock.sim.NodeByIP(toSupplierIP).LaunchContainers(fromBuyerIP, offerID, containerImageKey, portMappings, containerArgs,
		cpus, ram)

	return &rest.ContainerStatus{
		ImageKey:     containerImageKey,
		ID:           containerID,
		PortMappings: portMappings,
	}, err
}

func (mock *RemoteClientMock) StopLocalContainer(toSupplierIP string, containerID string) error {
	return mock.sim.NodeByIP(toSupplierIP).StopLocalContainer(containerID)
}

func (mock *RemoteClientMock) ObtainConfiguration(systemsNodeIP string) (*configuration.Configuration, error) {
	// Do Nothing (For now not necessary for the simulation)
	return nil, nil
}
