package overlay

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/node/common/guid"
)

type NodeMock struct {
	guid *guid.GUID
	ip   string
}

func NewNodeBytes(idBytes []byte) *NodeMock {
	return &NodeMock{
		guid: guid.NewGUIDBytes(idBytes),
		ip:   "",
	}
}

func NewRandomNode() *NodeMock {
	return &NodeMock{
		guid: guid.NewGUIDRandom(),
		ip:   util.RandomIP(),
	}
}

func (node *NodeMock) Bytes() []byte {
	return node.guid.Bytes()
}

func (node *NodeMock) IP() string {
	return node.ip
}

func (node *NodeMock) String() string {
	return node.guid.String()
}

func (node *NodeMock) Equals(nodeArg *NodeMock) bool {
	return node.guid.Equals(*nodeArg.guid)
}

func (node *NodeMock) Smaller(nodeArg *NodeMock) bool {
	return node.guid.Cmp(*nodeArg.guid) < 0
}
