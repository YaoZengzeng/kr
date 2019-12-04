package client

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/YaoZengzeng/kr/types"
)

type Client struct {
	c         *http.Client
	registry  string
	heartbeat time.Duration

	mtx      sync.Mutex
	services map[string]chan struct{}
}

type Option func(*Client) error

func WithRegistry(registry string) Option {
	return func(c *Client) error {
		c.registry = registry
		return nil
	}
}

func WithHeartbeat(heartbeat time.Duration) Option {
	return func(c *Client) error {
		c.heartbeat = heartbeat
		return nil
	}
}

func New(opts ...Option) (*Client, error) {
	c := &Client{
		services: make(map[string]chan struct{}),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.c = &http.Client{
		// Set timeout of http client to heartbeat period.
		Timeout: c.heartbeat,
	}

	return c, nil
}

func (c *Client) Register(service *types.Service) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	b, err := json.Marshal(service)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%x", md5.Sum(b))

	if _, ok := c.services[key]; ok {
		// The service has registered, return.
		return nil
	}

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(c.heartbeat)
		for {
			select {
			case <-ticker.C:
				resp, err := c.c.PostForm(c.registry, url.Values{
					"address":  {service.Address},
					"port":     {strconv.Itoa(service.Port)},
					"endpoint": {service.Endpoint},
				})
				if err != nil {
					log.Printf("register service to register failed: %v\n", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Printf("read body of http response failed: %v", err)
					}
					log.Printf("the status code of register is not 200, body: %v", string(body))
				}
			case <-stop:
				return
			}
		}
	}()

	c.services[key] = stop
	return nil
}

func (c *Client) Deregister(service *types.Service) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	b, err := json.Marshal(service)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%x", md5.Sum(b))

	ch, ok := c.services[key]
	if !ok {
		// If the service dosn't exist, do nothing and return.
		return nil
	}

	close(ch)
	delete(c.services, key)

	return nil
}
