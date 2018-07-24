package chord

import "github.com/strabox/caravela-sim/mocks/overlay"

type speedupNodeMock struct {
	*overlay.NodeMock
	index int
}

func newSpeedupNodeMock(index int, nodeMock *overlay.NodeMock) *speedupNodeMock {
	return &speedupNodeMock{
		NodeMock: nodeMock,
		index:    index,
	}
}
