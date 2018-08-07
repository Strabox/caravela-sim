package docker

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	myContainer "github.com/strabox/caravela/docker/container"
	"sync"
	"sync/atomic"
)

// Size of the container's IDs.
const containerIDSize = 64

// ClientMock mocks the interactions with the docker daemon.
// It implements the github.com/strabox/caravela/node/external DockerClient interface.
type ClientMock struct {
	numOfContainers    int64
	containersRunning  sync.Map
	resourcesGenerator ResourcesGenerator
}

// NewClientMock creates a new docker client mock to be used.
func NewClientMock(resourcesGenerator ResourcesGenerator) *ClientMock {
	return &ClientMock{
		numOfContainers:    0,
		containersRunning:  sync.Map{},
		resourcesGenerator: resourcesGenerator,
	}
}

func (cliMock *ClientMock) ContainersRunning() int64 {
	return cliMock.numOfContainers
}

// ===============================================================================
// =						   DockerClient Interface                            =
// ===============================================================================

func (cliMock *ClientMock) GetDockerCPUAndRAM() (int, int) {
	return cliMock.resourcesGenerator.Generate()
}

func (cliMock *ClientMock) CheckContainerStatus(containerID string) (myContainer.ContainerStatus, error) {
	_, exist := cliMock.containersRunning.Load(containerID)
	if exist {
		return myContainer.NewContainerStatus(myContainer.Running), nil
	} else {
		return myContainer.NewContainerStatus(myContainer.Finished), nil
	}
}

func (cliMock *ClientMock) RunContainer(contConfig types.ContainerConfig) (*types.ContainerStatus, error) {

	// Generate a random ID for the container and store it in an HashMap
	randomContainerID := util.RandomString(containerIDSize)
	cliMock.containersRunning.Store(randomContainerID, nil)
	atomic.AddInt64(&cliMock.numOfContainers, 1)

	return &types.ContainerStatus{
		ContainerConfig: contConfig,
		ContainerID:     randomContainerID,
		Status:          "Running",
	}, nil
}

func (cliMock *ClientMock) RemoveContainer(containerID string) error {
	if _, exist := cliMock.containersRunning.Load(containerID); exist {
		cliMock.containersRunning.Delete(containerID)
		cliMock.numOfContainers--
	}
	return nil
}
