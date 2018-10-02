package caravela

import "github.com/strabox/caravela/node"

// simNodeService provides an interface to obtain nodes from its IPs or GUIDs.
type simNodeService interface {
	NodeByIP(ip string) (*node.Node, int)
	NodeByGUID(guid string) (*node.Node, int)
}
