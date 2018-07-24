package caravela

import (
	"github.com/strabox/caravela/api/types"
)

const FakePort = 8000

func EmptyPortMappings() []types.PortMapping {
	return make([]types.PortMapping, 0)
}

func EmptyContainerArgs() []string {
	return make([]string, 0)
}
