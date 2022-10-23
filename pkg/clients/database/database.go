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

package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/digitalocean/godo"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane-contrib/provider-digitalocean/apis/database/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
)

// GenerateDatabase generates *godo.DatabaseRequest instance from LBParameters.
func GenerateDatabase(name string, in v1alpha1.DODatabaseClusterParameters, create *godo.DatabaseCreateRequest) {
	create.Name = name
	create.EngineSlug = do.StringValue(in.Engine)
	create.Version = do.StringValue(in.Version)
	create.NumNodes = in.NumNodes
	create.SizeSlug = in.Size
	create.Region = in.Region
	create.PrivateNetworkUUID = do.StringValue(in.PrivateNetworkUUID)
	create.Tags = in.Tags
}

// GenerateObservation creates an observation from a *godo.Database.
func GenerateObservation(observed *godo.Database) v1alpha1.DODatabaseClusterObservation {
	obs := v1alpha1.DODatabaseClusterObservation{
		ID:                 &observed.ID,
		Name:               observed.Name,
		Engine:             observed.EngineSlug,
		Version:            observed.VersionSlug,
		NumNodes:           observed.NumNodes,
		Size:               observed.SizeSlug,
		Region:             observed.RegionSlug,
		Status:             observed.Status,
		CreatedAt:          observed.CreatedAt.String(),
		PrivateNetworkUUID: observed.PrivateNetworkUUID,
		Tags:               observed.Tags,
		DbNames:            observed.DBNames,
		MaintenanceWindow: v1alpha1.DODatabaseClusterMaintenanceWindow{
			Day:         observed.MaintenanceWindow.Day,
			Hour:        observed.MaintenanceWindow.Hour,
			Pending:     observed.MaintenanceWindow.Pending,
			Description: observed.MaintenanceWindow.Description,
		},
	}

	obs.Users = make([]v1alpha1.DODatabaseClusterUser, len(observed.Users))
	for i, user := range observed.Users {
		obs.Users[i] = v1alpha1.DODatabaseClusterUser{
			Name: user.Name,
			Role: user.Role,
		}

		if user.MySQLSettings != nil {
			obs.Users[i].MySQLSettings = v1alpha1.DODatabaseUserMySQLSettings{
				AuthPlugin: user.MySQLSettings.AuthPlugin,
			}
		}
	}

	return obs
}

// GenerateConnectionDetails generates connection details for a DODatabaseCluster
func GenerateConnectionDetails(ctx context.Context, observation *godo.Database, c *godo.Client) managed.ConnectionDetails {
	connDetails := managed.ConnectionDetails{
		fmt.Sprintf("public-%s", xpv1.ResourceCredentialsSecretEndpointKey): []byte(observation.Connection.URI),
		fmt.Sprintf("public-%s", xpv1.ResourceCredentialsSecretPortKey):     []byte(strconv.Itoa(observation.Connection.Port)),
		fmt.Sprintf("public-%s", xpv1.ResourceCredentialsSecretUserKey):     []byte(observation.Connection.User),
		fmt.Sprintf("public-%s", xpv1.ResourceCredentialsSecretPasswordKey): []byte(observation.Connection.Password),
		"public-database": []byte(observation.Connection.Database),
		"public-host":     []byte(observation.Connection.Host),
		fmt.Sprintf("private-%s", xpv1.ResourceCredentialsSecretEndpointKey): []byte(observation.PrivateConnection.URI),
		fmt.Sprintf("private-%s", xpv1.ResourceCredentialsSecretPortKey):     []byte(strconv.Itoa(observation.PrivateConnection.Port)),
		fmt.Sprintf("private-%s", xpv1.ResourceCredentialsSecretUserKey):     []byte(observation.PrivateConnection.User),
		fmt.Sprintf("private-%s", xpv1.ResourceCredentialsSecretPasswordKey): []byte(observation.PrivateConnection.Password),
		"private-database": []byte(observation.PrivateConnection.Database),
		"private-host":     []byte(observation.PrivateConnection.Host),
	}

	if observation.Connection.SSL || observation.PrivateConnection.SSL {
		ca, _, err := c.Databases.GetCA(ctx, observation.ID)

		if err == nil && observation.Connection.SSL {
			connDetails[xpv1.ResourceCredentialsSecretCAKey] = ca.Certificate
		}
	}

	return connDetails
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied LBParameters that are set (i.e. non-zero) on the supplied
// LB.
func LateInitializeSpec(p *v1alpha1.DODatabaseClusterParameters, observed godo.Database) {
	p.Engine = do.LateInitializeString(p.Engine, observed.EngineSlug)
	p.Version = do.LateInitializeString(p.Version, observed.EngineSlug)
	p.NumNodes = observed.NumNodes
	p.Size = observed.SizeSlug
	p.Region = observed.RegionSlug
	p.PrivateNetworkUUID = do.LateInitializeString(p.PrivateNetworkUUID, observed.PrivateNetworkUUID)
	if len(p.Tags) == 0 && len(observed.Tags) != 0 {
		p.Tags = make([]string, len(observed.Tags))
		copy(p.Tags, observed.Tags)
	}
}
