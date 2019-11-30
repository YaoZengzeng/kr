// Memory based registry is just used for test.
package memory

import (
	"crypto/md5"
	"encoding/json"
	"time"
    "sync"

	"github.com/YaoZengzeng/kr/types"
)

type Registry struct {
	mtx		sync.RWMutex
	// key is the hash of service.
	store	map[[md5.Size]byte][]byte
}

type Item struct {
	Service *types.Service 	`json:"service"`
	Update	time.Time 		`json:"update"`
}

func NewRegistry() (*Registry, error) {
	return &Registry{
		store:	make(map[[md5.Size]byte][]byte),
	}, nil
}

func (r *Registry) Register(service *types.Service) error {
	b, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// For simplicity, don't consider the disorder of network packets.
	i := &Item{
		Service:	service,
		Update:		time.Now(),
	}
	value, err := json.Marshal(i)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	r.store[md5.Sum(b)] = value
	r.mtx.Unlock()

	return nil
}

func (r *Registry) ListServices() ([]*types.Service, error) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	res := make([]*types.Service, 0, len(r.store))
	for _, value := range r.store {
		item := &Item{}
		err := json.Unmarshal(value, item)
		if err != nil {
			return nil, err
		}

		res = append(res, item.Service)
	}

	return res, nil
}
