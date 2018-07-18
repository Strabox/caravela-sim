package discovery

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/strabox/caravela/api/rest"
	"net/http"
)

var nodeDiscoveryAPI Discovery = nil

func Init(router *mux.Router, nodeDiscovery Discovery) {
	nodeDiscoveryAPI = nodeDiscovery
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(createOffer)).Methods(http.MethodPost)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(refreshOffer)).Methods(http.MethodPatch)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(removeOffer)).Methods(http.MethodDelete)
	router.Handle(rest.DiscoveryOfferBaseEndpoint, rest.AppHandler(getOffers)).Methods(http.MethodGet)
	router.Handle(rest.DiscoveryNeighborOfferBaseEndpoint, rest.AppHandler(neighborOffers)).Methods(http.MethodPatch)
}

func createOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var createOfferMsg rest.CreateOfferMessage

	err := rest.ReceiveJSONFromHttp(w, r, &createOfferMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- CREATE OFFER To: %s, ID: %d, Amt: %d, Res: <%d,%d>, From: %s",
		createOfferMsg.ToTraderGUID, createOfferMsg.OfferID, createOfferMsg.Amount, createOfferMsg.CPUs,
		createOfferMsg.RAM, createOfferMsg.FromSupplierIP)

	nodeDiscoveryAPI.CreateOffer(createOfferMsg.FromSupplierGUID, createOfferMsg.FromSupplierIP, createOfferMsg.ToTraderGUID,
		createOfferMsg.OfferID, createOfferMsg.Amount, createOfferMsg.CPUs, createOfferMsg.RAM)
	return nil, nil
}

func refreshOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var offerRefreshMsg rest.RefreshOfferMessage

	err := rest.ReceiveJSONFromHttp(w, r, &offerRefreshMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REFRESH OFFER ID: %d, From: %s", offerRefreshMsg.OfferID,
		offerRefreshMsg.FromTraderGUID)

	res := nodeDiscoveryAPI.RefreshOffer(offerRefreshMsg.OfferID, offerRefreshMsg.FromTraderGUID)
	refreshOfferResp := rest.RefreshOfferResponseMessage{Refreshed: res}
	return refreshOfferResp, nil
}

func removeOffer(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var offerRemoveMsg rest.OfferRemoveMessage

	err := rest.ReceiveJSONFromHttp(w, r, &offerRemoveMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("<-- REMOVE OFFER To: %s, ID: %d, From: %s", offerRemoveMsg.ToTraderGUID,
		offerRemoveMsg.OfferID, offerRemoveMsg.FromSupplierIP)

	nodeDiscoveryAPI.RemoveOffer(offerRemoveMsg.FromSupplierIP, offerRemoveMsg.FromSupplierGUID,
		offerRemoveMsg.ToTraderGUID, offerRemoveMsg.OfferID)
	return nil, nil
}

func getOffers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var getOffersMsg rest.GetOffersMessage

	err := rest.ReceiveJSONFromHttp(w, r, &getOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- GET OFFERS To: %s", getOffersMsg.ToTraderGUID)

	offers := nodeDiscoveryAPI.GetOffers(getOffersMsg.ToTraderGUID, getOffersMsg.Relay, getOffersMsg.FromNodeGUID)

	var offersResp []rest.OfferJSON = nil
	offersResp = make([]rest.OfferJSON, len(offers))
	for index, offer := range offers {
		offersResp[index].ID = offer.ID
		offersResp[index].SupplierIP = offer.SupplierIP
	}

	return rest.OffersListMessage{Offers: offersResp}, nil
}

func neighborOffers(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var neighborOffersMsg rest.NeighborOffersMessage

	err := rest.ReceiveJSONFromHttp(w, r, &neighborOffersMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("<-- NEIGHBOR OFFERS To: %s, TraderOffering: <%s;%s>",
		neighborOffersMsg.ToNeighborGUID, neighborOffersMsg.NeighborOfferingIP, neighborOffersMsg.NeighborOfferingGUID)

	nodeDiscoveryAPI.AdvertiseNeighborOffers(neighborOffersMsg.ToNeighborGUID, neighborOffersMsg.FromNeighborGUID,
		neighborOffersMsg.NeighborOfferingIP, neighborOffersMsg.NeighborOfferingGUID)

	return nil, nil
}
