package overlay

import (
	"github.com/strabox/caravela-sim/util"
	"math/big"
	"math/bits"
)

var idSizeBytes = 16

func Init(guidSizeBytes int) {
	idSizeBytes = guidSizeBytes
}

type NodeMock struct {
	ip string
	id *big.Int
}

func NewNodeString(idString string) *NodeMock {
	temp := big.NewInt(0)
	temp.SetString(idString, 10)
	return &NodeMock{
		id: temp,
		ip: "",
	}
}

func NewNodeBytes(id []byte) *NodeMock {
	temp := big.NewInt(0)
	temp.SetBytes(id)
	return &NodeMock{
		id: temp,
		ip: "",
	}
}

func NewRandomNode() *NodeMock {
	id := make([]byte, idSizeBytes)
	util.RandomHash(id)
	temp := big.NewInt(0)
	temp.SetBytes(id)
	return &NodeMock{
		id: temp,
		ip: util.RandomIP(),
	}
}

func (node *NodeMock) Bytes() []byte {
	res := make([]byte, idSizeBytes)
	idBytes := node.id.Bytes()
	index := 0
	for ; index < idSizeBytes-len(idBytes); index++ { // Padding the higher bytes with 0
		res[index] = 0
	}
	for k := 0; index < idSizeBytes; k++ {
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

func FingersToFollow(nodeID, key []byte) (acc int) {
	acc = 0
	tempKey := big.NewInt(0)
	tempKey.SetBytes(key)
	tempNodeID := big.NewInt(0)
	tempNodeID.SetBytes(nodeID)

	diff := tempKey.Sub(tempKey, tempNodeID)
	diffBytes := diff.Bytes()
	for i := 0; i < len(diffBytes); i++ {
		acc += bits.OnesCount8(diffBytes[i])
	}
	return
}
