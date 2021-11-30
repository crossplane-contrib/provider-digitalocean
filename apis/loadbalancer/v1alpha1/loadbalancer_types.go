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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// Known LB statuses.
const (
	StatusNew    = "new"
	StatusActive = "active"
	StatusOff    = "off"
)

// LBParameters define the desired state of a DigitalOcean LoadBalancer.
// Most fields map directly to a LoadBalancer:
// https://developers.digitalocean.com/documentation/v2/#load-balancers
type LBParameters struct {
	// Region: The unique slug identifier for the region that you wish to
	// deploy in.
	// +immutable
	Region string `json:"region"`

	// Algorithm: The load balancing algorithm used to determine which backend
	// Droplet will be selected by a client.
	// It must be either "round_robin" or "least_connections".
	// +kubebuilder:validation:Enum=round_robin;least_connections
	Algorithm string `json:"algorithm"`

	// API Server port. It must be valid ports range (1-65535). If omitted, default value is 6443.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`

	// An object specifying health check settings for the Load Balancer. If omitted, default values will be provided.
	// +optional
	HealthCheck DOLoadBalancerHealthCheck `json:"healthCheck,omitempty"`

	// Tags: A flat array of tag names as strings to apply to the LB after it
	// is created. Tag names can either be existing or new tags.
	// +optional
	// +immutable
	Tags []string `json:"tags,omitempty"`

	// VPCUUID: A string specifying the UUID of the VPC to which the LB
	// will be assigned. If excluded, beginning on April 7th, 2020, the LB
	// will be assigned to your account's default VPC for the region.
	// +optional
	// +immutable
	VPCUUID *string `json:"vpc_uuid,omitempty"`
}

// DOLoadBalancerHealthCheck define the DigitalOcean loadbalancers health check configurations.
type DOLoadBalancerHealthCheck struct {
	// The number of seconds between between two consecutive health checks. The value must be between 3 and 300.
	// If not specified, the default value is 10.
	// +optional
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Maximum=300
	Interval int `json:"interval,omitempty"`
	// The number of seconds the Load Balancer instance will wait for a response until marking a health check as failed.
	// The value must be between 3 and 300. If not specified, the default value is 5.
	// +optional
	// +kubebuilder:validation:Minimum=3
	// +kubebuilder:validation:Maximum=300
	Timeout int `json:"timeout,omitempty"`
	// The number of times a health check must fail for a backend Droplet to be marked "unhealthy" and be removed from the pool.
	// The vaule must be between 2 and 10. If not specified, the default value is 3.
	// +optional
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:validation:Maximum=10
	UnhealthyThreshold int `json:"unhealthyThreshold,omitempty"`
	// The number of times a health check must pass for a backend Droplet to be marked "healthy" and be re-added to the pool.
	// The vaule must be between 2 and 10. If not specified, the default value is 5.
	// +optional
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:validation:Maximum=10
	HealthyThreshold int `json:"healthyThreshold,omitempty"`
}

// A LBObservation reflects the observed state of a LB on DigitalOcean.
type LBObservation struct {
	// CreationTimestamp in RFC3339 text format.
	CreationTimestamp string `json:"creationTimestamp,omitempty"`

	// ID for the resource. This identifier is defined by the server.
	ID string `json:"id,omitempty"`

	// IP for the resource.
	IP int `json:"ip,omitempty"`

	// A Status string indicating the state of the LB instance.
	//
	// Possible values:
	//   "new"
	//   "active"
	//   "off"
	Status string `json:"status,omitempty"`
}

// A LBSpec defines the desired state of a LB.
type LBSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       LBParameters `json:"forProvider"`
}

// A LBStatus represents the observed state of a LB.
type LBStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LBObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A LB is a managed resource that represents a DigitalOcean LB.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,do}
type LB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LBSpec   `json:"spec"`
	Status LBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LBList contains a list of LBs.
type LBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LB `json:"items"`
}
