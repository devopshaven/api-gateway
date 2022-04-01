package gateway

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

type ConfigClient struct {
	mutex   *sync.Mutex
	cmName  string
	ctx     context.Context
	cancel  func()
	client  *kubernetes.Clientset
	ns      string
	watcher watch.Interface

	config *GatewayConfig
}

func (cc *ConfigClient) Config() *GatewayConfig {
	return cc.config
}

func readNamespace() string {
	file, err := ioutil.ReadFile(namespaceFile)
	if err != nil {
		return "gateway"
	}

	return string(file)
}

func getClient(pathToCfg string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if pathToCfg == "" {
		// in cluster access
		log.Info().Msg("Using in cluster config")
		config, err = rest.InClusterConfig()
	} else {
		log.Info().Msg("Using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", pathToCfg)
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func NewConfigClient() *ConfigClient {
	ctx, cancel := context.WithCancel(context.Background())

	var clientset *kubernetes.Clientset

	_, err := os.Open(namespaceFile)
	if err != nil {
		home, _ := os.UserHomeDir()
		clientset, err = getClient(path.Join(home, ".kube/ace-platform.yaml"))
		if err != nil {
			panic(fmt.Errorf("cannot create kubernetes client: %w", err))
		}
	} else { // In cluster
		clientset, err = getClient("")
		if err != nil {
			panic(fmt.Errorf("cannot create kubernetes client: %w", err))
		}
	}

	ns := readNamespace()

	return &ConfigClient{
		mutex:  new(sync.Mutex),
		cmName: "gateway-config",
		ctx:    ctx,
		cancel: cancel,
		client: clientset,
		ns:     ns,
	}
}

func (cc *ConfigClient) StartWatcher() {
	ch, err := cc.createWatcher()
	if err != nil {
		log.Fatal().Err(err).Msgf("cannot start: %s", err)
	}

	go cc.handleConfigChanges(ch)
}

func (cc *ConfigClient) createWatcher() (<-chan watch.Event, error) {
	watcher, err := cc.client.CoreV1().ConfigMaps(cc.ns).Watch(
		context.TODO(),
		metav1.SingleObject(metav1.ObjectMeta{
			Name:      cc.cmName,
			Namespace: cc.ns,
		}))
	if err != nil {
		return nil, fmt.Errorf("unable to create watcher: %w", err)
	}

	cc.watcher = watcher

	return watcher.ResultChan(), err
}

func (cc *ConfigClient) handleConfigChanges(eventChannel <-chan watch.Event) {
	defer cc.cancel()

	for {
		select {
		case event, open := <-eventChannel:
			if open {
				switch event.Type {
				case watch.Added:
					fallthrough
				case watch.Modified:
					cc.mutex.Lock()
					log.Debug().Msg("configmap modified")
					// Update our endpoint
					if updatedMap, ok := event.Object.(*corev1.ConfigMap); ok {
						cc.updateConfig(updatedMap.Data["services.yaml"])
					}
					cc.mutex.Unlock()
				case watch.Deleted:
					cc.mutex.Lock()
					// Fall back to the default value
					fmt.Println("configmap deleted...")
					cc.mutex.Unlock()
				default:
					// Do nothing
				}
			} else {
				// If eventChannel is closed, it means the server has closed the connection
				return
			}
		case <-cc.ctx.Done():
			cc.watcher.Stop()
		}
	}
}

func (cc *ConfigClient) updateConfig(config string) error {
	var conf GatewayConfig

	// fmt.Println(config)

	if err := yaml.Unmarshal([]byte(config), &conf); err != nil {
		return fmt.Errorf("cannot decode config: %w", err)
	}

	// fmt.Printf("Config: %+v\n", conf)

	cc.config = &conf

	return nil
}
