package registry

import (
	"github.com/YaoZengzeng/kr/types"
)

type Registry interface {
	Register(*types.Service) error
	ListServices() ([]*types.Service, error)
}
