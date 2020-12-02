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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// Known Droplet statuses.
const (
	StatusNew     = "new"
	StatusActive  = "active"
	StatusOff     = "off"
	StatusArchive = "archive"
)

// DropletParameters define the desired state of a DigitalOcean Droplet.
// Most fields map directly to a Droplet:
// https://developers.digitalocean.com/documentation/v2/#droplets
type DropletParameters struct {
	// Region: The unique slug identifier for the region that you wish to
	// deploy in.
	// +immutable
	Region string `json:"region"`

	// Size: The unique slug identifier for the size that you wish to select
	// for this Droplet.
	// +immutable
	Size string `json:"size"`

	// Image: The image ID of a public or private image, or the unique slug
	// identifier for a public image. This image will be the base image for
	// your Droplet.
	// +immutable
	Image string `json:"image"`

	// SSHKeys: An array containing the IDs or fingerprints of the SSH keys
	// that you wish to embed in the Droplet's root account upon creation.
	// +optional
	// +immutable
	SSHKeys []string `json:"ssh_keys,omitempty"`

	// Backups: A boolean indicating whether automated backups should be enabled
	// for the Droplet. Automated backups can only be enabled when the Droplet is
	// created.
	// +optional
	// +immutable
	Backups *bool `json:"backups,omitempty"`

	// IPv6: A boolean indicating whether IPv6 is enabled on the Droplet.
	// +optional
	// +immutable
	IPv6 *bool `json:"ipv6,omitempty"`

	// PrivateNetworking: This parameter has been deprecated. Use 'vpc_uuid'
	// instead to specify a VPC network for the Droplet. If no `vpc_uuid` is
	// provided, the Droplet will be placed in the default VPC.
	// +optional
	// +immutable
	PrivateNetworking *bool `json:"private_networking,omitempty"`

	// Monitoring: A boolean indicating whether to install the DigitalOcean
	// agent for monitoring.
	// +optional
	// +immutable
	Monitoring *bool `json:"monitoring,omitempty"`

	// Volumes: A flat array including the unique string identifier for each block
	// storage volume to be attached to the Droplet. At the moment a volume can only
	// be attached to a single Droplet.
	// +optional
	// +immutable
	Volumes []string `json:"volumes,omitempty"`

	// Tags: A flat array of tag names as strings to apply to the Droplet after it
	// is created. Tag names can either be existing or new tags.
	// +optional
	// +immutable
	Tags []string `json:"tags,omitempty"`

	// VPCUUID: A string specifying the UUID of the VPC to which the Droplet
	// will be assigned. If excluded, beginning on April 7th, 2020, the Droplet
	// will be assigned to your account's default VPC for the region.
	// +optional
	// +immutable
	VPCUUID *string `json:"vpc_uuid,omitempty"`
}

// A DropletObservation reflects the observed state of a Droplet on DigitalOcean.
type DropletObservation struct {
	// CreationTimestamp in RFC3339 text format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// ID for the resource. This identifier is defined by the server.
	ID int `json:"id,omitempty"`

	// A Status string indicating the state of the Droplet instance.
	//
	// Possible values:
	//   "new"
	//   "active"
	//   "off"
	//   "archive"
	Status string `json:"status,omitempty"`
}

// A DropletSpec defines the desired state of a Droplet.
type DropletSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DropletParameters `json:"forProvider"`
}

// A DropletStatus represents the observed state of a Droplet.
type DropletStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DropletObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Droplet is a managed resource that represents a DigitalOcean Droplet.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,do}
type Droplet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DropletSpec   `json:"spec"`
	Status DropletStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DropletList contains a list of Droplet.
type DropletList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Droplet `json:"items"`
}
