package chord

type speedupNodeMock struct {
	*NodeMock
	index int
}

func newSpeedupNodeMock(index int, nodeMock *NodeMock) *speedupNodeMock {
	return &speedupNodeMock{
		NodeMock: nodeMock,
		index:    index,
	}
}
