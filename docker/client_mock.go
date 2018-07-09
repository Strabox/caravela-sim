package docker

import (
	"github.com/strabox/caravela/api/rest"
	dockerAPI "github.com/strabox/caravela/docker"
	"sync"
)

type ClientMock struct {
	containersRunning sync.Map
	containersIdGen   int
	idGenMutex        sync.Mutex
}

func NewClientMock() *ClientMock {
	return &ClientMock{
		containersRunning: sync.Map{},
		containersIdGen:   0,
		idGenMutex:        sync.Mutex{},
	}
}

func (cliMock *ClientMock) generateContainerID() string {
	cliMock.idGenMutex.Lock()
	defer cliMock.idGenMutex.Unlock()

	id := cliMock.containersIdGen
	cliMock.containersIdGen++
	return string(id)
}

/*
===============================================================================
							  Docker Client Interface
===============================================================================
*/

func (cliMock *ClientMock) GetDockerCPUAndRAM() (int, int) {
	// TODO: Instead of hardcode values put a function that distributes the resources of each node realistically
	return 4, 2048
}

func (cliMock *ClientMock) CheckContainerStatus(containerID string) (dockerAPI.ContainerStatus, error) {
	_, exist := cliMock.containersRunning.Load(containerID)
	if exist {
		return dockerAPI.NewContainerStatus(dockerAPI.Running), nil
	} else {
		return dockerAPI.NewContainerStatus(dockerAPI.Finished), nil
	}
}

func (cliMock *ClientMock) RunContainer(imageKey string, portMappings []rest.PortMapping, args []string, cpus int64,
	ram int) (string, error) {

	// Generate a random ID for the container and store it in an HashMap
	randomContainerID := cliMock.generateContainerID()
	cliMock.containersRunning.Store(randomContainerID, nil)
	return randomContainerID, nil
}

func (cliMock *ClientMock) RemoveContainer(containerID string) {
	cliMock.containersRunning.Delete(containerID)
}
