/*
Copyright 2021 The Crossplane Authors.
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

package kubernetes

import (
	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	"github.com/digitalocean/godo"
)

// GenerateKubernetes generates *godo.RegistryCreateRequest instance from DOContainerRegistryParameters.
func GenerateContainerRegistry(name string, in v1alpha1.DOContainerRegistryParameters, create *godo.RegistryCreateRequest) {
	create.Name = name
	create.SubscriptionTierSlug = in.SubscriptionTier
	create.Region = do.StringValue(in.Region)
}

// RegistryLateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied DOContainerRegistryParameters that are set (i.e. non-zero) on the supplied
// Kubernetes Cluster.
func RegistryLateInitializeSpec(p *v1alpha1.DOContainerRegistryParameters, observed godo.Registry) {
	p.Region = do.LateInitializeString(p.Region, observed.Region)
}
