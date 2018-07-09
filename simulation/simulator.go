package simulation

import caravelaNode "github.com/strabox/caravela/node"

type Simulator interface {
	Init()
	Start()
	NodeByIP(ip string) *caravelaNode.Node
	NodeByGUID(guid string) *caravelaNode.Node
}
