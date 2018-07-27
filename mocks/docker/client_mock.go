package docker

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	myContainer "github.com/strabox/caravela/docker/container"
	"sync"
)

// Size of the container's IDs.
const containerIDSize = 64

// ClientMock mocks the interactions with the docker daemon.
// It implements the github.com/strabox/caravela/node/external DockerClient interface.
type ClientMock struct {
	numOfContainers   int
	containersRunning sync.Map
}

// NewClientMock creates a new docker client mock.
func NewClientMock() *ClientMock {
	return &ClientMock{
		numOfContainers:   0,
		containersRunning: sync.Map{},
	}
}

func (cliMock *ClientMock) ContainersRunning() int {
	return cliMock.numOfContainers
}

// ===============================================================================
// =						   DockerClient Interface                            =
// ===============================================================================

func (cliMock *ClientMock) GetDockerCPUAndRAM() (int, int) {
	// TODO: Instead of hardcoded values put a function that
	// distributes the resources of each node realistically
	return 2, 512
}

func (cliMock *ClientMock) CheckContainerStatus(containerID string) (myContainer.ContainerStatus, error) {
	_, exist := cliMock.containersRunning.Load(containerID)
	if exist {
		return myContainer.NewContainerStatus(myContainer.Running), nil
	} else {
		return myContainer.NewContainerStatus(myContainer.Finished), nil
	}
}

func (cliMock *ClientMock) RunContainer(imageKey string, portMappings []types.PortMapping, args []string,
	cpus int64, ram int) (string, error) {

	// Generate a random ID for the container and store it in an HashMap
	randomContainerID := util.RandomString(containerIDSize)
	cliMock.containersRunning.Store(randomContainerID, nil)
	cliMock.numOfContainers++
	return randomContainerID, nil
}

func (cliMock *ClientMock) RemoveContainer(containerID string) error {
	if _, exist := cliMock.containersRunning.Load(containerID); exist {
		cliMock.containersRunning.Delete(containerID)
		cliMock.numOfContainers--
	}
	return nil
}
