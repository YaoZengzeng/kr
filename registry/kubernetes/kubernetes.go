package kubernetes

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"time"

	// "k8s.io/apimachinery/pkg/api/errors"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisterv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/YaoZengzeng/kr/types"
)

var (
	nameprefix = "service"

	labelKey   = "registered-service-filter"
	labelValue = "true"

	annotationKey = "service-content"
)

type Registry struct {
	client corev1.EndpointsInterface
	lister corelisterv1.EndpointsNamespaceLister

	// Time To Live for a service, default is 60 * time.Second.
	ttl time.Duration
}

func NewRegistry() (*Registry, error) {
	// Only support in-cluster config.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return newRegistry(clientset, 60 * time.Second)
}

// Easy for test: take fake.NewSimpleClientset() as input.
func newRegistry(clientset kubernetes.Interface, ttl time.Duration) (*Registry, error) {
	informers := informers.NewSharedInformerFactory(clientset, 0)

	endpointInformer := informers.Core().V1().Endpoints().Informer()

	stop := make(chan struct{})
	informers.Start(stop)

	if !cache.WaitForCacheSync(stop, endpointInformer.HasSynced) {
		return nil, fmt.Errorf("failed to wait endpoint informer synced")
	}

	return &Registry{
		client: clientset.CoreV1().Endpoints(apiv1.NamespaceDefault),
		// Only operate endpoints in default namespace.
		lister: informers.Core().V1().Endpoints().Lister().Endpoints(apiv1.NamespaceDefault),

		ttl: ttl,
	}, nil
}

type Item struct {
	Service *types.Service `json:"service"`
	Update  time.Time      `json:"update"`
}

func (r *Registry) Register(service *types.Service) error {
	b, err := json.Marshal(service)
	if err != nil {
		return err
	}

	exist := true
	name := fmt.Sprintf("%s-%x", nameprefix, md5.Sum(b))
	result, err := r.lister.Get(name)
	if err != nil {
		// If we failed to get endpoint from cache, just assume it doesn't exist.
		exist = false
	}

	now := time.Now()

	if exist {
		value := result.Annotations[annotationKey]
		oldItem := &Item{}
		// TODO: handle this error properly.
		if err := json.Unmarshal([]byte(value), oldItem); err != nil {
			return err
		}
		// If we have multiple instances of registry, it's possible that have items newer than now.
		// Only update the endpoint if we could set the Update field newer.
		if now.After(oldItem.Update) {
			oldItem.Update = now
			value, err := json.Marshal(oldItem)
			if err != nil {
				return err
			}

			result.Annotations[annotationKey] = string(value)
			// TODO: handle this error properly.
			if _, err := r.client.Update(result); err != nil {
				return err
			}
		}
	} else {
		// For simplicity, don't consider the disorder of network packets.
		item := &Item{
			Service: service,
			Update:  now,
		}
		value, err := json.Marshal(item)
		if err != nil {
			return err
		}

		endpoint := &apiv1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					labelKey: labelValue,
				},
				Annotations: map[string]string{
					annotationKey: string(value),
				},
			},
		}
		// TODO: handle this error properly.
		if _, err := r.client.Create(endpoint); err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) ListServices() ([]*types.Service, error) {
	endpoints, err := r.lister.List(labels.SelectorFromSet(labels.Set{labelKey: labelValue}))
	if err != nil {
		return nil, err
	}

	now := time.Now()

	res := make([]*types.Service, 0, len(endpoints))
	for _, endpoint := range endpoints {
		value := endpoint.Annotations[annotationKey]
		item := &Item{}
		err := json.Unmarshal([]byte(value), item)
		if err != nil {
			log.Printf("failed to unmarshal registered service from %v\n", endpoint.Name)
			continue
		}

		if item.Update.Add(r.ttl).Before(now) {
			// The registered service has expired, skip.
			continue
		}

		res = append(res, item.Service)
	}

	return res, nil
}
