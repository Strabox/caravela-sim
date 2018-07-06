package overlay

import (
	nodeAPI "github.com/strabox/caravela/node/api"
	"github.com/strabox/caravela/overlay"
)

type ChordMock struct {
	// TODO
}

func NewChordMock() *ChordMock {
	return &ChordMock{}
}

func (chord *ChordMock) Create(thisNode nodeAPI.OverlayMembership) error {
	// Do Nothing
	return nil
}

func (chord *ChordMock) Join(overlayNodeIP string, overlayNodePort int, thisNode nodeAPI.OverlayMembership) error {
	// Do Nothing
	return nil
}

func (chord *ChordMock) Lookup(key []byte) ([]*overlay.Node, error) {
	// Do Nothing
	return nil, nil
}

func (chord *ChordMock) Neighbors(nodeID []byte) ([]*overlay.Node, error) {
	// Do Nothing
	return nil, nil
}

func (chord *ChordMock) NodeID() ([]byte, error) {
	// Do Nothing
	return nil, nil
}

func (chord *ChordMock) Leave() error {
	// Do Nothing
	return nil
}
