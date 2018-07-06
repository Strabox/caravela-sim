package docker

import (
	"github.com/strabox/caravela/api/rest"
	dockerAPI "github.com/strabox/caravela/docker"
)

type ClientMock struct {
	// TODO
}

func NewClientMock() *ClientMock {
	return &ClientMock{}
}

func (cli *ClientMock) GetDockerCPUAndRAM() (int, int) {
	// TODO
	return 0, 0
}

func (cli *ClientMock) CheckContainerStatus(containerID string) (dockerAPI.ContainerStatus, error) {
	// TODO
	return dockerAPI.NewContainerStatus(0), nil
}

func (cli *ClientMock) RunContainer(imageKey string, portMappings []rest.PortMapping, args []string, cpus int64,
	ram int) (string, error) {
	// TODO
	return "", nil
}

func (cli *ClientMock) RemoveContainer(containerID string) {
	// TODO
	return
}
