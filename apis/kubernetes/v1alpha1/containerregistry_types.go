/*
Copyright 2020 The Crossplane Authors.

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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DOContainerRegistryParameters define the desired state of a DigitalOcean Container Registry.
// Most fields map directly to a Containe rRegistry:
// https://docs.digitalocean.com/reference/api/api-reference/#tag/Container-Registry
type DOContainerRegistryParameters struct {
	// The slug of the subscription tier to sign up for.
	// Valid values can be retrieved using the options endpoint.
	SubscriptionTier string `json:"subscription_tier"`

	// Slug of the region where registry data is stored.
	// When not provided, a region will be selected.
	// +immutable
	// +kubebuilder:validation:Optional
	Region *string `json:"region,omitempty"`
}

// The Tier defines a subscription tier for a Container Registry.
type Tier struct {
	// The name of the subscription tier.
	Name string `json:"name"`

	// The slug identifier of the subscription tier.
	Slug string `json:"slug"`

	// The number of repositories included in the subscription tier.
	// 0 indicates that the subscription tier includes unlimited repositories.
	IncludedRepositories uint64 `json:"included_repositories"`

	// The amount of storage included in the subscription tier in bytes.
	IncludedStorageBytes uint64 `json:"included_storage_bytes"`

	// A boolean indicating whether the subscription tier supports
	// additional storage above what is included in the base plan at an additional cost per GiB used.
	AllowStorageOverage bool `json:"allow_storage_overrage"`

	// The amount of outbound data transfer included in the subscription tier in bytes.
	IncludedBandwidthBytes uint64 `json:"included_bandwidth_bytes"`

	// The monthly cost of the subscription tier in cents.
	MonthlyPriceInCents uint64 `json:"monthly_price_in_cents"`
}

// The Subscription defines a subscription for a Container Registry.
type Subscription struct {
	// An object specifying the subscription tier for a Container Registry.
	Tier Tier `json:"tier"`

	// The time at which the subscription was created.
	CreatedAt string `json:"created_at"`

	// The time at which the subscription was last updated.
	UpdatedAt string `json:"updated_at"`
}

// A DOContainerRegistryObservation reflects the observed state of a Container Registry on DigitalOcean.
type DOContainerRegistryObservation struct {
	// A globally unique name for the container registry.
	// Must be lowercase and be composed only of numbers, letters and -, up to a limit of 63 characters.
	Name string `json:"name"`

	// A time value given in ISO8601 combined date and time format that represents when the registry was created.
	CreatedAt string `json:"created_at"`

	// Slug of the region where registry data is stored.
	Region string `json:"region"`

	// The amount of storage used in the registry in bytes.
	StorageUsageBytes uint64 `json:"storage_usage_bytes"`

	// The time at which the storage usage was updated.
	// Storage usage is calculated asynchronously, and may not immediately reflect pushes to the registry.
	StorageUsageBytesUpdatedAt string `json:"storage_usage_bytes_updated_at"`

	// An object specifying the subscription for a Container Registry.
	Subscription Subscription `json:"subscription"`
}

// A DOContainerRegistrySpec defines the desired state of a ContainerRegistry.
type DOContainerRegistrySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DOContainerRegistryParameters `json:"forProvider"`
}

// A DOContainerRegistryStatus represents the observed state of a ContainerRegistry.
type DOContainerRegistryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DOContainerRegistryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DOContainerRegistry is a managed resource that represents a DigitalOcean Container Registry.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,do}
type DOContainerRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DOContainerRegistrySpec   `json:"spec"`
	Status DOContainerRegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DOContainerRegistryList contains a list of ContainerRegistrys.
type DOContainerRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DOContainerRegistry `json:"items"`
}
