package caravela

import (
	"encoding/json"
	"github.com/strabox/caravela/api/types"
)

func sizeofNode(node *types.Node) int {
	/*
		nodeSizeBytes := uintptr(0)
		nodeSizeBytes += sizeofString(node.IP)
		nodeSizeBytes += sizeofString(node.GUID)
	*/
	string, _ := json.Marshal(node)
	return len(string)
}

func sizeofOffer(offer *types.Offer) int {
	/*
		offerSizeBytes := uintptr(0)
		offerSizeBytes += unsafe.Sizeof(offer.ID)
		offerSizeBytes += unsafe.Sizeof(offer.Amount)
		offerSizeBytes += unsafe.Sizeof(offer.ContainersRunning)
		offerSizeBytes += unsafe.Offsetof(offer.FreeResources)
		offerSizeBytes += unsafe.Offsetof(offer.UsedResources)
	*/
	string, _ := json.Marshal(offer)
	return len(string)
}

func sizeofAvailableOffer(offer *types.AvailableOffer) int {
	/*
		offerSizeBytes := uintptr(0)
		offerSizeBytes += unsafe.Sizeof(offer.ID)
		offerSizeBytes += unsafe.Sizeof(offer.Amount)
		offerSizeBytes += unsafe.Sizeof(offer.ContainersRunning)
		offerSizeBytes += sizeofResources(&offer.FreeResources)
		offerSizeBytes += sizeofResources(&offer.UsedResources)
		offerSizeBytes += sizeofString(offer.SupplierIP)
		offerSizeBytes += unsafe.Sizeof(offer.Weight)
	*/
	string, _ := json.Marshal(offer)
	return len(string)
}

func sizeofContainerConfigSlice(containerConfigs []types.ContainerConfig) int {
	/*
		containerConfigSliceSizeBytes := uintptr(0)
		for i := range containerConfigs {
			containerConfigSliceSizeBytes += sizeofContainerConfig(&containerConfigs[i])
		}
	*/
	containerConfigSliceSizeBytes := 0
	for i := range containerConfigs {
		containerConfigSliceSizeBytes += sizeofContainerConfig(&containerConfigs[i])
	}
	return containerConfigSliceSizeBytes
}

func sizeofContainerConfig(containerConfig *types.ContainerConfig) int {
	/*
		containerConfigSizeBytes := uintptr(0)
		containerConfigSizeBytes += sizeofString(containerConfig.Name)
		containerConfigSizeBytes += sizeofResources(&containerConfig.Resources)
		containerConfigSizeBytes += unsafe.Sizeof(containerConfig.GroupPolicy)
		containerConfigSizeBytes += sizeofStringSlice(containerConfig.Args)
		for i := range containerConfig.PortMappings {
			containerConfigSizeBytes += sizeofPortMapping(&containerConfig.PortMappings[i])
		}
	*/
	string, _ := json.Marshal(containerConfig)
	return len(string)
}

/*
func sizeofResources(res *types.Resources) uintptr {
	resourcesSizeBytes := uintptr(0)
	resourcesSizeBytes += unsafe.Sizeof(res.CPUClass)
	resourcesSizeBytes += unsafe.Sizeof(res.CPUs)
	resourcesSizeBytes += unsafe.Sizeof(res.Memory)
	return resourcesSizeBytes
}
*/
/*
func sizeofPortMapping(portMapping *types.PortMapping) uintptr {
	portMappingSizeBytes := uintptr(0)
	portMappingSizeBytes += unsafe.Sizeof(portMapping.HostPort)
	portMappingSizeBytes += unsafe.Sizeof(portMapping.ContainerPort)
	portMappingSizeBytes += sizeofString(portMapping.Protocol)
	return portMappingSizeBytes
}
*/

/* ==================================== Go's builtin types =============================== */

// sizeofStringSlice returns the size of the string slice in bytes.
func sizeofStringSlice(slice []string) int {
	/*
		sizeAccBytes := uintptr(0)
		for _, s := range slice {
			sizeAccBytes += sizeofString(s)
		}*/
	sizeAccBytes := 0
	for _, s := range slice {
		sizeAccBytes += sizeofString(s)
	}
	return sizeAccBytes
}

// sizeofString returns the size of the string in bytes.
func sizeofString(s string) int {
	if s == "" {
		return 0
	}
	return len(s)
}
