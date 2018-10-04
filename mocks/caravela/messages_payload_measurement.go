package caravela

import (
	"encoding/json"
	"github.com/strabox/caravela/api/rest/util"
	"github.com/strabox/caravela/api/types"
)

func sizeofCreateOfferMessage(msg *util.CreateOfferMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofRefreshOfferMessage(msg *util.RefreshOfferMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofRefreshOfferMessageResponse(msg *util.RefreshOfferResponseMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofUpdateOfferMessage(msg *util.UpdateOfferMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofRemoveOfferMessage(msg *util.OfferRemoveMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofGetOffersMessage(msg *util.GetOffersMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofLaunchContainerMessage(msg *util.LaunchContainerMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofStopLocalContainerMessage(msg *util.StopLocalContainerMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofNeighborOfferMessage(msg *util.NeighborOffersMsg) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofAvailableOffersMessage(msg []types.AvailableOffer) int {
	string, _ := json.Marshal(msg)
	return len(string)
}

func sizeofContainersStatusMessage(msg []types.ContainerStatus) int {
	string, _ := json.Marshal(msg)
	return len(string)
}
