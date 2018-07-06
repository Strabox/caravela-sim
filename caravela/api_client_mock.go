package caravela

import (
	"github.com/strabox/caravela/api/rest"
	"github.com/strabox/caravela/configuration"
	nodeAPI "github.com/strabox/caravela/node/api"
)

type RemoteClientMock struct {
	// TODO
}

func NewRemoteClientMock() *RemoteClientMock {
	return &RemoteClientMock{}
}

func (mock *RemoteClientMock) CreateOffer(fromSupplierIP string, fromSupplierGUID string, toTraderIP string,
	toTraderGUID string, offerID int64, amount int, cpus int, ram int) error {
	// Do Nothing
	return nil
}

func (mock *RemoteClientMock) RefreshOffer(toSupplierIP string, fromTraderGUID string, offerID int64) (bool, error) {
	// Do Nothing
	return false, nil
}

func (mock *RemoteClientMock) RemoveOffer(fromSupplierIP string, fromSupplierGUID, toTraderIP string,
	toTraderGUID string, offerID int64) error {
	// Do Nothing
	return nil
}

func (mock *RemoteClientMock) GetOffers(toTraderIP string, toTraderGUID string, relay bool,
	fromNodeGUID string) ([]nodeAPI.Offer, error) {
	// Do Nothing
	return nil, nil
}

func (mock *RemoteClientMock) AdvertiseOffersNeighbor(toNeighborTraderIP string, toNeighborTraderGUID string,
	fromTraderGUID string, traderOfferingGUID string, traderOfferingIP string) error {
	// Do Nothing
	return nil
}

func (mock *RemoteClientMock) LaunchContainer(toSupplierIP string, fromBuyerIP string, offerID int64, containerImageKey string,
	portMappings []rest.PortMapping, containerArgs []string, cpus int, ram int) error {
	// Do Nothing
	return nil
}

func (mock *RemoteClientMock) ObtainConfiguration(systemsNodeIP string) (*configuration.Configuration, error) {
	// Do Nothing
	return nil, nil
}
