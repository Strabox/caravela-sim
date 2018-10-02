package chord

import (
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/node/common/guid"
	"math/big"
)

var (
	preCalculatedFingerOffset []*big.Int = nil
	maxNumIDs                 *big.Int   = nil
)

type fingerMock struct {
	fingerID *guid.GUID
	simIndex int
}

type NodeMock struct {
	guid        *guid.GUID
	ip          string
	fingerTable []fingerMock
}

func NewNodeBytes(idBytes []byte) *NodeMock {
	return &NodeMock{
		guid: guid.NewGUIDBytes(idBytes),
		ip:   "",
	}
}

func Init(guidSizeBits int) {
	preCalculatedFingerOffset = make([]*big.Int, guidSizeBits)
	for i := range preCalculatedFingerOffset {
		fingerOffset := big.NewInt(0)
		fingerOffset.Exp(big.NewInt(2), big.NewInt(int64(i)), nil)
		preCalculatedFingerOffset[i] = fingerOffset
	}

	maxNumIDs = big.NewInt(0)
	maxNumIDs.Exp(big.NewInt(2), big.NewInt(int64(guidSizeBits)), nil)
}

func maxNumGUIDs() *big.Int {
	maxNumGUIDs := big.NewInt(0)
	maxNumGUIDs.Set(maxNumIDs)
	return maxNumGUIDs
}

func NewNodeRandomIP(guidBigInt *big.Int) *NodeMock {
	res := &NodeMock{
		guid:        guid.NewGUIDBigInt(guidBigInt),
		ip:          util.RandomIP(),
		fingerTable: make([]fingerMock, guid.SizeBits()),
	}

	for i := range res.fingerTable {
		newFinger := big.NewInt(0)
		tempBigInt := big.NewInt(0)
		tempBigInt.Add(res.guid.BigInt(), preCalculatedFingerOffset[i])
		newFinger.Mod(tempBigInt, maxNumIDs)
		res.fingerTable[i].fingerID = guid.NewGUIDBigInt(newFinger)
	}
	return res
}

func (n *NodeMock) SetFingersNodeIndexes(fingersIndexes []int, fingersGUIDs []*guid.GUID) {
	for i := range n.fingerTable {
		n.fingerTable[i].simIndex = fingersIndexes[i]
		n.fingerTable[i].fingerID = fingersGUIDs[i]
	}
}

func (n *NodeMock) FingerTable() []*guid.GUID {
	res := make([]*guid.GUID, len(n.fingerTable))
	for i := range res {
		res[i] = n.fingerTable[i].fingerID
	}
	return res
}

func (n *NodeMock) belongToExcludedLimits(key, bottom, top *big.Int) bool {
	if bottom.Cmp(top) > 0 {
		return (key.Cmp(top) < 0) || (key.Cmp(bottom) > 0)
	} else {
		return (key.Cmp(top) < 0) && (key.Cmp(bottom) > 0)
	}
}

func (n *NodeMock) belongToIncludedTop(key, bottom, top *big.Int) bool {
	if bottom.Cmp(top) > 0 {
		return (key.Cmp(top) <= 0) || (key.Cmp(bottom) > 0)
	} else {
		return (key.Cmp(top) <= 0) && (key.Cmp(bottom) > 0)
	}
}

func (n *NodeMock) Lookup(fromNodeIndex int, key []byte) (int, bool) {
	keyBigInt := big.NewInt(0)
	keyBigInt.SetBytes(key)
	if n.belongToIncludedTop(keyBigInt, n.guid.BigInt(), n.fingerTable[0].fingerID.BigInt()) {
		return n.fingerTable[0].simIndex, true
	} else {
		for i := len(n.fingerTable) - 1; i >= 0; i-- {
			if n.belongToExcludedLimits(n.fingerTable[i].fingerID.BigInt(), n.guid.BigInt(), keyBigInt) {
				return n.fingerTable[i].simIndex, false
			}
		}
	}
	return fromNodeIndex, false
}

func (n *NodeMock) Bytes() []byte {
	return n.guid.Bytes()
}

func (n *NodeMock) IP() string {
	return n.ip
}

func (n *NodeMock) String() string {
	return n.guid.String()
}

func (n *NodeMock) Equals(nodeArg *NodeMock) bool {
	return n.guid.Equals(*nodeArg.guid)
}

func (n *NodeMock) Smaller(nodeArg *NodeMock) bool {
	return n.guid.Cmp(*nodeArg.guid) < 0
}

func (n *NodeMock) HigherOrEqual(nodeArg *NodeMock) bool {
	return n.guid.Cmp(*nodeArg.guid) >= 0
}

func (n *NodeMock) SetZeroGUID() {
	n.guid = guid.NewGUIDInteger(0)
}
