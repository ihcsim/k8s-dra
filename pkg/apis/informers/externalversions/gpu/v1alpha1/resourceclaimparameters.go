/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	versioned "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpuv1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/gpu/v1alpha1"
	internalinterfaces "github.com/ihcsim/k8s-dra/pkg/apis/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/ihcsim/k8s-dra/pkg/apis/listers/gpu/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ResourceClaimParametersInformer provides access to a shared informer and lister for
// ResourceClaimParameters.
type ResourceClaimParametersInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ResourceClaimParametersLister
}

type resourceClaimParametersInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewResourceClaimParametersInformer constructs a new informer for ResourceClaimParameters type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewResourceClaimParametersInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredResourceClaimParametersInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredResourceClaimParametersInformer constructs a new informer for ResourceClaimParameters type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredResourceClaimParametersInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.GpuV1alpha1().ResourceClaimParameters(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.GpuV1alpha1().ResourceClaimParameters(namespace).Watch(context.TODO(), options)
			},
		},
		&gpuv1alpha1.ResourceClaimParameters{},
		resyncPeriod,
		indexers,
	)
}

func (f *resourceClaimParametersInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredResourceClaimParametersInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *resourceClaimParametersInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&gpuv1alpha1.ResourceClaimParameters{}, f.defaultInformer)
}

func (f *resourceClaimParametersInformer) Lister() v1alpha1.ResourceClaimParametersLister {
	return v1alpha1.NewResourceClaimParametersLister(f.Informer().GetIndexer())
}
