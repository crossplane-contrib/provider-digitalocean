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

// KubernetesClusterParameters define the desired state of a DigitalOcean Kubernetes Cluster
// Most fields map directly to a KubernetesCluster.
// See docs https://docs.digitalocean.com/reference/api/api-reference/#operation/create_kubernetes_cluster
type KubernetesClusterParameters struct {
	// A human-readable name for a Kubernetes cluster.
	Name string `json:"name"`

	// The slug identifier for the region where the Kubernetes cluster is located.
	Region string `json:"region"`

	// The slug identifier for the version of Kubernetes used for the cluster.
	// If set to a minor version (e.g. "1.14"), the latest version within it will be used (e.g. "1.14.6-do.1");
	// if set to "latest", the latest published version will be used. See the /v2/kubernetes/options endpoint
	// to find all currently available versions.
	Version string `json:"version"`

	// A string specifying the UUID of the VPC to which the Kubernetes cluster is assigned.
	// +kubebuilder:validation:Optional
	VPCUUID string `json:"vpcuui,omitempty"`

	// An array of tags applied to the Kubernetes cluster. All clusters are automatically tagged k8s and k8s:$K8S_CLUSTER_ID.
	// +kubebuilder:validation:Optional
	Tags []string `json:"tags,omitempty"`

	// An array of objects specifying the details of the worker nodes available to the Kubernetes cluster.
	NodePools []KubernetesNodePool `json:"nodePools"`

	// An object specifying the maintenance window policy for the Kubernetes cluster.
	// +kubebuilder:validation:Optional
	MaintenancePolicy KubernetesClusterMaintenancePolicy `json:"maintenancePolicy,omitempty"`

	// A boolean value indicating whether the cluster will be automatically upgraded to new patch releases during its maintenance window.
	// +kubebuilder:validation:Optional
	AutoUpgrade bool `json:"autoUpgrade,omitempty"`

	// A boolean value indicating whether surge upgrade is enabled/disabled for the cluster. Surge upgrade makes cluster upgrades fast and reliable by bringing up new nodes before destroying the outdated nodes.
	// +kubebuilder:validation:Optional
	SurgeUpgrade bool `json:"surgeUpgrade,omitempty"`

	// A boolean value indicating whether the control plane is run in a highly available configuration in the cluster. Highly available control planes incur less downtime.
	// +kubebuilder:validation:Optional
	HighlyAvailable bool `json:"highlyAvailable,omitempty"`
}

// A KubernetesClusterObservation reflects the observed state of a KubernetesCluster on DigitalOcean.
// See docs https://docs.digitalocean.com/reference/api/api-reference/#operation/create_kubernetes_cluster
type KubernetesClusterObservation struct {
	// ID for the resource. This identifier is defined by the server.
	ID string `json:"id,omitempty"`

	// A human-readable name for a Kubernetes cluster.
	Name string `json:"name"`

	// The slug identifier for the region where the Kubernetes cluster is located.
	Region string `json:"region"`

	// The slug identifier for the version of Kubernetes used for the cluster.
	// If set to a minor version (e.g. "1.14"), the latest version within it will be used (e.g. "1.14.6-do.1");
	// if set to "latest", the latest published version will be used. See the /v2/kubernetes/options endpoint
	// to find all currently available versions.
	Version string `json:"version"`

	// The range of IP addresses in the overlay network of the Kubernetes cluster in CIDR notation.
	ClusterSubnet string `json:"clusterSubnet,omitempty"`

	// The range of assignable IP addresses for services running in the Kubernetes cluster in CIDR notation.
	ServiceSubnet string `json:"serviceSubnet,omitempty"`

	// A string specifying the UUID of the VPC to which the Kubernetes cluster is assigned.
	VPCUUID string `json:"vpcuuid,omitempty"`

	// The public IPv4 address of the Kubernetes master node.
	IPV4 string `json:"ipv4,omitempty"`

	// The base URL of the API server on the Kubernetes master node.
	Endpoint string `json:"endpoint,omitempty"`

	// An array of tags applied to the Kubernetes cluster. All clusters are automatically tagged k8s and k8s:$K8S_CLUSTER_ID.
	Tags []string `json:"tags,omitempty"`

	// An array of objects specifying the details of the worker nodes available to the Kubernetes cluster.
	NodePools []KubernetesNodePoolObservation `json:"nodePools"`

	// An object specifying the maintenance window policy for the Kubernetes cluster.
	MaintenancePolicy KubernetesClusterMaintenancePolicyObservation `json:"maintenancePolicy,omitempty"`

	// A boolean value indicating whether the cluster will be automatically upgraded to new patch releases during its maintenance window.
	AutoUpgrade bool `json:"autoUpgrade,omitempty"`

	// An object containing a state attribute whose value is set to a string indicating the current status of the cluster.
	Status KubernetesStatus `json:"status,omitempty"`

	// A time value given in ISO8601 combined date and time format that represents when the Kubernetes cluster was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// A time value given in ISO8601 combined date and time format that represents when the Kubernetes cluster was last updated.
	UpdatedAt string `json:"updatedAt,omitempty"`

	// A boolean value indicating whether surge upgrade is enabled/disabled for the cluster. Surge upgrade makes cluster upgrades fast and reliable by bringing up new nodes before destroying the outdated nodes.
	SurgeUpgrade bool `json:"surgeUpgrade,omitempty"`

	// A boolean value indicating whether the control plane is run in a highly available configuration in the cluster. Highly available control planes incur less downtime.
	HighlyAvailable bool `json:"highlyAvailable,omitempty"`

	// A read-only boolean value indicating if a container registry is integrated with the cluster.
	RegistryEnabled bool `json:"registryEnabled,omitempty"`
}

type KubernetesNodePool struct {
	// The slug identifier for the type of Droplet used as workers in the node pool.
	Size string `json:"size"`

	// A human-readable name for the node pool.
	Name string `json:"name"`

	// The number of Droplet instances in the node pool.
	Count int `json:"count"`

	// An array containing the tags applied to the node pool. All node pools are automatically tagged k8s, k8s-worker, and k8s:$K8S_CLUSTER_ID.
	// +kubebuilder:validation:Optional
	Tags []string `json:"tags,omitempty"`

	// An object containing a set of Kubernetes labels. The keys and are values are both user-defined.
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	// An array of taints to apply to all nodes in a pool.
	// +kubebuilder:validation:Optional
	Taints []KubernetesNodePoolTaint `json:"taints,omitempty"`

	// A boolean value indicating whether auto-scaling is enabled for this node pool.
	// +kubebuilder:validation:Optional
	AutoScale bool `json:"autoScale,omitempty"`

	// The minimum number of nodes that this node pool can be auto-scaled to. The value will be 0 if auto_scale is set to false.
	// +kubebuilder:validation:Optional
	MinNodes int `json:"minNodes,omitempty"`

	// The maximum number of nodes that this node pool can be auto-scaled to. The value will be 0 if auto_scale is set to false.
	// +kubebuilder:validation:Optional
	MaxNodes int `json:"maxNodes,omitempty"`
}

type KubernetesNodePoolObservation struct {
	// A unique ID that can be used to identify and reference a specific node pool.
	Id string `json:"id,omitempty"`

	NodePool KubernetesNodePool `json:",inline"`

	// An object specifying the details of a specific worker node in a node pool.
	Nodes []KubernetesNode `json:"nodes,omitempty"`
}

// KubernetesNodePoolTaint represents a Kubernetes Node Pool Taint.
// Taints will automatically be applied to all existing nodes and any subsequent nodes added to the pool. When a taint is removed, it is removed from all nodes in the pool
type KubernetesNodePoolTaint struct {
	// An arbitrary string. The key and value fields of the taint object form a key-value pair.
	// For example, if the value of the key field is "special" and the value of the value field is "gpu",
	// the key value pair would be special=gpu.
	// +kubebuilder:validation:Optional
	Key string `json:"key,omitempty"`

	// An arbitrary string. The key and value fields of the taint object form a key-value pair.
	// For example, if the value of the key field is "special" and the value of the value field is "gpu",
	// the key value pair would be special=gpu.
	// +kubebuilder:validation:Optional
	Value string `json:"value,omitempty"`

	//How the node reacts to pods that it won't tolerate. Available effect values are NoSchedule, PreferNoSchedule, and NoExecute.
	// +kubebuilder:validation:Optional
	Effect string `json:"effect,omitempty"`
}

type KubernetesNode struct {
	// A unique ID that can be used to identify and reference the node.
	// +kubebuilder:validation:Optional
	Id string `json:"id,omitempty"`

	// An automatically generated, human-readable name for the node.
	// +kubebuilder:validation:Optional
	Name string `json:"name,omitempty"`

	// An object containing a state attribute whose value is set to a string indicating the current status of the node.
	// +kubebuilder:validation:Optional
	Status KubernetesStatus `json:"status,omitempty"`

	// The ID of the Droplet used for the worker node.
	// +kubebuilder:validation:Optional
	DropletId string `json:"dropletID,omitempty"`

	// A time value given in ISO8601 combined date and time format that represents when the node was created.
	// +kubebuilder:validation:Optional
	CreatedAt string `json:"createdAt,omitempty"`

	// A time value given in ISO8601 combined date and time format that represents when the node was last updated.
	// +kubebuilder:validation:Optional
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type KubernetesClusterMaintenancePolicy struct {
	// The start time in UTC of the maintenance window policy in 24-hour clock format / HH:MM notation (e.g., 15:00).
	// +kubebuilder:validation:Optional
	StartTime string `json:"startTime,omitempty"`

	// The day of the maintenance window policy. May be one of monday through sunday, or any to indicate an arbitrary week day.
	// +kubebuilder:validation:Optional
	Day string `json:"day,omitempty"`
}

type KubernetesClusterMaintenancePolicyObservation struct {
	Policy KubernetesClusterMaintenancePolicy `json:",inline"`

	// The duration of the maintenance window policy in human-readable format.
	// +kubebuilder:validation:Optional
	Duration string `json:"duration,omitempty"`
}

type KubernetesStatus struct {
	// A string indicating the current status of the node.
	// +kubebuilder:validation:Optional
	State string `json:"state,omitempty"`

	// A message relating to the current state
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`
}

// A KubernetesClusterSpec defines the desired state of a KubernetesCluster.
type KubernetesClusterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       KubernetesClusterParameters `json:"forProvider"`
}

// A KubernetesClusterStatus represents the observed state of a KubernetesCluster.
type KubernetesClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          KubernetesClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A KubernetesCluster is a managed resource that represents a DigitalOcean Kubernetes Cluster.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,do}
type KubernetesCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesClusterSpec   `json:"spec"`
	Status KubernetesClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubernetesClusterList contains a list of KubernetesClusters.
type KubernetesClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesCluster `json:"items"`
}
