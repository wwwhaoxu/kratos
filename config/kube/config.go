package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kratos/kratos/v2/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Option is kube option.
type Option func(*options)

type options struct {
	// kube namespace
	Namespace string
	// kube labelSelector example `app=test`
	LabelSelector string
	// kube fieldSelector example `app=test`
	FieldSelector string
	// set KubeConfig out-of-cluster Use outside cluster
	KubeConfig string
	// set master url
	Master string
}

// Namespace with kube namespace.
func Namespace(ns string) Option {
	return func(o *options) {
		o.Namespace = ns
	}
}

// LabelSelector with kube label selector.
func LabelSelector(label string) Option {
	return func(o *options) {
		o.LabelSelector = label
	}
}

// FieldSelector with kube field selector.
func FieldSelector(field string) Option {
	return func(o *options) {
		o.FieldSelector = field
	}
}

// KubeConfig with kube config.
func KubeConfig(config string) Option {
	return func(o *options) {
		o.KubeConfig = config
	}
}

// Master with kube master.
func Master(master string) Option {
	return func(o *options) {
		o.Master = master
	}
}

type kube struct {
	opts   options
	client *kubernetes.Clientset
}

// NewSource new a kube config source.
func NewSource(opts ...Option) config.Source {
	options := options{}
	for _, o := range opts {
		o(&options)
	}
	return &kube{
		opts: options,
	}
}

func (k *kube) init() (err error) {
	var config *rest.Config
	if k.opts.KubeConfig != "" {
		if config, err = clientcmd.BuildConfigFromFlags(k.opts.Master, k.opts.KubeConfig); err != nil {
			return err
		}
	} else {
		if config, err = rest.InClusterConfig(); err != nil {
			return err
		}
	}
	if k.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	return nil
}

func (k *kube) load() (kvs []*config.KeyValue, err error) {
	cmList, err := k.client.
		CoreV1().
		ConfigMaps(k.opts.Namespace).
		List(context.Background(), metav1.ListOptions{
			LabelSelector: k.opts.LabelSelector,
			FieldSelector: k.opts.FieldSelector,
		})
	if err != nil {
		return nil, err
	}
	for _, cm := range cmList.Items {
		kvs = append(kvs, k.configMap(cm)...)
	}
	return kvs, nil
}

func (k *kube) configMap(cm v1.ConfigMap) (kvs []*config.KeyValue) {
	for name, val := range cm.Data {
		kvs = append(kvs, &config.KeyValue{
			Key:   fmt.Sprintf("%s/%s/%s", k.opts.Namespace, cm.Name, name),
			Value: []byte(val),
		})
	}
	return kvs
}

func (k *kube) Load() ([]*config.KeyValue, error) {
	if k.opts.Namespace == "" {
		return nil, errors.New("options namespace not full")
	}
	if err := k.init(); err != nil {
		return nil, err
	}
	return k.load()
}

func (k *kube) Watch() (config.Watcher, error) {
	return newWatcher(k)
}
