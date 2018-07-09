package overlay

import (
	"math/big"
)

var IdSizeBytes = 16

type NodeMock struct {
	ip string
	id *big.Int
}

func NewNode(id []byte) *NodeMock {
	temp := big.NewInt(0)
	temp.SetBytes(id)
	return &NodeMock{
		id: temp,
		ip: "",
	}
}

func NewRandomNode() *NodeMock {
	id := make([]byte, IdSizeBytes)
	generateRandomHash(id)
	temp := big.NewInt(0)
	temp.SetBytes(id)
	return &NodeMock{
		id: temp,
		ip: generateRandomIP(),
	}
}

func (node *NodeMock) Bytes() []byte {
	res := make([]byte, IdSizeBytes)
	idBytes := node.id.Bytes()
	index := 0
	for ; index < IdSizeBytes-len(idBytes); index++ { // Padding the higher bytes with 0
		res[index] = 0
	}
	for k := 0; index < IdSizeBytes; k++ {
		res[index] = idBytes[k]
		index++
	}
	return res
}

func (node *NodeMock) IP() string {
	return node.ip
}

func (node *NodeMock) String() string {
	return node.id.String()
}

func (node *NodeMock) Equals(nodeArg *NodeMock) bool {
	return node.id.Cmp(nodeArg.id) == 0
}

func (node *NodeMock) Smaller(nodeArg *NodeMock) bool {
	return node.id.Cmp(nodeArg.id) < 0
}
