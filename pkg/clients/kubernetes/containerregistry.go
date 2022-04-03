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
	"context"

	"github.com/digitalocean/godo"

	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
)

// RegistryClient is the external client used for DOContainerRegistry Custom Resource
type RegistryClient interface {
	Get(context.Context) (*godo.Registry, *godo.Response, error)
	GetSubscription(context.Context) (*godo.RegistrySubscription, *godo.Response, error)
	Create(context.Context, *godo.RegistryCreateRequest) (*godo.Registry, *godo.Response, error)
	UpdateSubscription(context.Context, *godo.RegistrySubscriptionUpdateRequest) (*godo.RegistrySubscription, *godo.Response, error)
	Delete(context.Context) (*godo.Response, error)
}

// GenerateContainerRegistry generates *godo.RegistryCreateRequest instance from DOContainerRegistryParameters.
func GenerateContainerRegistry(name string, in v1alpha1.DOContainerRegistryParameters, create *godo.RegistryCreateRequest) {
	create.Name = name
	create.SubscriptionTierSlug = in.SubscriptionTier
	create.Region = do.StringValue(in.Region)
}

// RegistryLateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied DOContainerRegistryParameters that are set (i.e. non-zero) on the supplied
// Container Registry.
func RegistryLateInitializeSpec(p *v1alpha1.DOContainerRegistryParameters, observed godo.Registry) {
	p.Region = do.LateInitializeString(p.Region, observed.Region)
}

// GenerateContainerRegistryObservation generates DOContainerRegistryObservation instance
// from godo.Registry and godo.RegistrySubscription
func GenerateContainerRegistryObservation(registry *godo.Registry, subscription *godo.RegistrySubscription) v1alpha1.DOContainerRegistryObservation {
	return v1alpha1.DOContainerRegistryObservation{
		Name:                       registry.Name,
		Region:                     registry.Region,
		CreatedAt:                  registry.CreatedAt.String(),
		StorageUsageBytes:          registry.StorageUsageBytes,
		StorageUsageBytesUpdatedAt: registry.StorageUsageBytesUpdatedAt.String(),
		Subscription: v1alpha1.Subscription{
			Tier: v1alpha1.Tier{
				Name:                   subscription.Tier.Name,
				Slug:                   subscription.Tier.Slug,
				IncludedRepositories:   subscription.Tier.IncludedRepositories,
				IncludedStorageBytes:   subscription.Tier.IncludedStorageBytes,
				AllowStorageOverage:    subscription.Tier.AllowStorageOverage,
				IncludedBandwidthBytes: subscription.Tier.IncludedBandwidthBytes,
				MonthlyPriceInCents:    subscription.Tier.MonthlyPriceInCents,
			},
			CreatedAt: subscription.CreatedAt.String(),
			UpdatedAt: subscription.UpdatedAt.String(),
		},
	}
}
