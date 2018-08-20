package docker

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	myContainer "github.com/strabox/caravela/docker/container"
	"github.com/strabox/caravela/docker/events"
	"sync"
	"sync/atomic"
)

// Size of the container's IDs.
const containerIDSize = 64

// ClientMock mocks the interactions with the docker daemon.
// It implements the github.com/strabox/caravela/node/external DockerClient interface.
type ClientMock struct {
	maxCPUS            int
	maxRAM             int
	numOfContainers    int64
	containersRunning  sync.Map
	resourcesGenerator ResourcesGenerator
}

// NewClientMock creates a new docker client mock to be used.
func NewClientMock(resourcesGenerator ResourcesGenerator) *ClientMock {
	return &ClientMock{
		maxCPUS:            0,
		maxRAM:             0,
		numOfContainers:    0,
		containersRunning:  sync.Map{},
		resourcesGenerator: resourcesGenerator,
	}
}

func (cliMock *ClientMock) ContainersRunning() int64 {
	return cliMock.numOfContainers
}

func (cliMock *ClientMock) MaxResourcesAvailable() (int, int) {
	return cliMock.maxCPUS, cliMock.maxRAM
}

// ===============================================================================
// =						   DockerClient Interface                            =
// ===============================================================================

func (cliMock *ClientMock) Start() <-chan *events.Event {
	// Do Nothing (Not necessary for the simulation)
	return nil
}

func (cliMock *ClientMock) GetDockerCPUAndRAM() (int, int) {
	cpus, ram := cliMock.resourcesGenerator.Generate()
	cliMock.maxCPUS += cpus
	cliMock.maxRAM += ram
	return cpus, ram
}

func (cliMock *ClientMock) CheckContainerStatus(containerID string) (myContainer.Status, error) {
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
		atomic.AddInt64(&cliMock.numOfContainers, -1)
	}
	return nil
}
