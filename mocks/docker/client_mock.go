package docker

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	myContainer "github.com/strabox/caravela/docker/container"
	"sync"
)

const containerIDSize = 64

type ClientMock struct {
	numOfContainers   int
	containersRunning sync.Map
	idGenMutex        sync.Mutex
}

func NewClientMock() *ClientMock {
	return &ClientMock{
		containersRunning: sync.Map{},
		idGenMutex:        sync.Mutex{},
	}
}

func (cliMock *ClientMock) ContainersRunning() int {
	return cliMock.numOfContainers
}

/*
===============================================================================
							  Docker Client Interface
===============================================================================
*/

func (cliMock *ClientMock) GetDockerCPUAndRAM() (int, int) {
	// TODO: Instead of hardcode values put a function that
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
	cliMock.numOfContainers++

	// Generate a random ID for the container and store it in an HashMap
	randomContainerID := util.RandomString(containerIDSize)
	cliMock.containersRunning.Store(randomContainerID, nil)
	return randomContainerID, nil
}

func (cliMock *ClientMock) RemoveContainer(containerID string) error {
	cliMock.numOfContainers--
	cliMock.containersRunning.Delete(containerID)
	return nil
}
