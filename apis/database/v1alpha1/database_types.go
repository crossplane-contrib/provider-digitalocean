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

// Known Database Cluster statuses
const (
	StatusCreating  = "creating"
	StatusOnline    = "online"
	StatusResizing  = "resizing"
	StatusMigrating = "migrating"
	StatusForking   = "forking"
)

// A DODatabaseClusterParameters defines the desired state of a DigitalOcean Database Cluster.
// All fields map directly to a Database Cluster
// https://docs.digitalocean.com/reference/api/api-reference/#operation/create_database_cluster
type DODatabaseClusterParameters struct {
	// Engine: A slug representing the database engine used for the cluster. The possible values are: "pg" for PostgreSQL, "mysql" for MySQL, "redis" for Redis, and "mongodb" for MongoDB.
	// +kubebuilder:validation:Enum="pg";"mysql";"redis";"mongodb"
	// +immutable
	Engine *string `json:"engine"`

	// Version: A string representing the version of the database engine in use for the cluster (Optional).
	// +optional
	// +immutable
	Version *string `json:"version,omitempty"`

	// NumNodes: The number of nodes in the database cluster.
	// +immutable
	NumNodes int `json:"numNodes"`

	// Size: The slug identifier representing the size of the nodes in the database cluster.
	// +immutable
	Size string `json:"size"`

	// Region: The slug identifier for the region where the database cluster is located.
	// +immutable
	Region string `json:"region"`

	// PrivateNetworkUUID: A string specifying the UUID of the VPC to which the database cluster will be assigned. If excluded, the cluster when creating a new database cluster, it will be assigned to your account's default VPC for the region (Optional).
	// +optional
	// +immutable
	PrivateNetworkUUID *string `json:"privateNetworkUUID,omitempty"`

	// Tags: An array of tags that have been applied to the database cluster (Optional).
	// +optional
	// +immutable
	Tags []string `json:"tags,omitempty"`
}

// A DODatabaseClusterObservation reflects the observed state of a Database Cluster on DigitalOcean.
// https://docs.digitalocean.com/reference/api/api-reference/#operation/create_database_cluster
type DODatabaseClusterObservation struct {
	// A unique ID that can be used to identify and reference a database cluster.
	// +kubebuilder:validation:Optional
	ID *string `json:"id,omitempty"`

	// A unique, human-readable name referring to a database cluster.
	Name string `json:"name"`

	// A slug representing the database engine used for the cluster. The possible values are: "pg" for PostgreSQL, "mysql" for MySQL, "redis" for Redis, and "mongodb" for MongoDB
	Engine string `json:"engine"`

	// A string representing the version of the database engine in use for the cluster.
	Version string `json:"version,omitempty"`

	// The number of nodes in the database cluster.
	NumNodes int `json:"numNodes"`

	// The slug identifier representing the size of the nodes in the database cluster.
	Size string `json:"size"`

	// The slug identifier for the region where the database cluster is located.
	Region string `json:"region"`

	// A string representing the current status of the database cluster.
	//
	// Possible values:
	//	"creating"
	//	"online"
	//	"resizing"
	//	"migrating"
	//	"forking"
	Status string `json:"status,omitempty"`

	// A time value given in ISO8601 combined date and time format that represents when the database cluster was created.
	CreatedAt string `json:"createdAt,omitempty"`

	// A string specifying the UUID of the VPC to which the database cluster will be assigned. If excluded, the cluster when creating a new database cluster, it will be assigned to your account's default VPC for the region.
	PrivateNetworkUUID string `json:"privateNetworkUUID,omitempty"`

	// An array of tags that have been applied to the database cluster.
	Tags []string `json:"tags,omitempty"`

	// An array of strings containing the names of databases created in the database cluster.
	DbNames []string `json:"dbNames,omitempty"`

	Connection DODatabaseClusterConnection `json:"connection,omitempty"`

	PrivateConnection DODatabaseClusterConnection `json:"private_connection"`

	Users []DODatabaseClusterUser `json:"users,omitempty"`

	// +kubebuilder:validation:Optional
	MaintenanceWindow DODatabaseClusterMaintenanceWindow `json:"maintenanceWindow,omitempty"`
}

// A DODatabaseClusterConnection defines the connection information for a Database Cluster.
type DODatabaseClusterConnection struct {
	// A connection string in the format accepted by the psql command. This is provided as a convenience and should be able to be constructed by the other attributes.
	URI *string `json:"uri,omitempty"`

	// The name of the default database.
	Database *string `json:"database,omitempty"`

	// The FQDN pointing to the database cluster's current primary node.
	Host *string `json:"host,omitempty"`

	// The port on which the database cluster is listening.
	Port *int `json:"port,omitempty"`

	// The default user for the database.
	User *string `json:"user,omitempty"`

	// The randomly generated password for the default user.
	Password *string `json:"password,omitempty"`

	// A boolean value indicating if the connection should be made over SSL.
	SSL *bool `json:"ssl,omitempty"`
}

// The DODatabaseClusterUser defines a Database Cluster User.
type DODatabaseClusterUser struct {
	Name string `json:"name"`

	// A string representing the database user's role. The value will be either "primary" or "normal".
	Role string `json:"role,omitempty"`

	// A randomly generated password for the database user.
	Password string `json:"password,omitempty"`

	// +kubebuilder:validation:Optional
	MySQLSettings DODatabaseUserMySQLSettings `json:"mySQLSettings,omitempty"`
}

// DODatabaseUserMySQLSettings Represents the MySQL Settings of a user for a DigitalOcean Database Cluster
type DODatabaseUserMySQLSettings struct {
	// A string specifying the authentication method to be used for connections to the MySQL user account.
	// The valid values are mysql_native_password or caching_sha2_password. If excluded when creating a new user,
	// the default for the version of MySQL in use will be used. As of MySQL 8.0, the default is caching_sha2_password.
	AuthPlugin string `json:"authPlugin"`
}

// A DODatabaseClusterMaintenanceWindow defines a Database Cluster Maintenance Window.
type DODatabaseClusterMaintenanceWindow struct {
	// The day of the week on which to apply maintenance updates.
	Day string `json:"day"`

	// The hour in UTC at which maintenance updates will be applied in 24 hour format.
	Hour string `json:"hour"`

	// A boolean value indicating whether any maintenance is scheduled to be performed in the next window.
	Pending bool `json:"pending,omitempty"`

	// A list of strings, each containing information about a pending maintenance update.
	Description []string `json:"description,omitempty"`
}

// A DODatabaseClusterSpec defines the desired state of a Database Cluster
type DODatabaseClusterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DODatabaseClusterParameters `json:"forProvider"`
}

// A DODatabaseClusterStatus represents the observed state of a Database Cluster
type DODatabaseClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DODatabaseClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DODatabaseCluster is a managed resource that represents a DigitalOcean Database Cluster.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,do}
type DODatabaseCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DODatabaseClusterSpec   `json:"spec"`
	Status DODatabaseClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DODatabaseClusterList contains a list of Database Clusters.
type DODatabaseClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DODatabaseCluster `json:"items"`
}
