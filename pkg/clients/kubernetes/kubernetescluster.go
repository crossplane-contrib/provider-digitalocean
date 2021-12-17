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
	"net/http"

	"github.com/digitalocean/godo"

	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
)

// GenerateKubernetes generates *godo.KubernetesRequest instance from DOKubernetesClusterParameters.
func GenerateKubernetes(name string, in v1alpha1.DOKubernetesClusterParameters, create *godo.KubernetesClusterCreateRequest) {
	create.Name = name
	create.VersionSlug = in.Version
	create.RegionSlug = in.Region
	create.VPCUUID = do.StringValue(in.VPCUUID)
	create.Tags = in.Tags
	create.MaintenancePolicy = &godo.KubernetesMaintenancePolicy{
		StartTime: in.MaintenancePolicy.StartTime,
		Day:       getDayFromParam(in.MaintenancePolicy.Day),
	}
	create.AutoUpgrade = do.BoolValue(in.AutoUpgrade)
	create.SurgeUpgrade = do.BoolValue(in.SurgeUpgrade)
	create.HA = do.BoolValue(in.HighlyAvailable)

	create.NodePools = make([]*godo.KubernetesNodePoolCreateRequest, len(in.NodePools))
	for i, nodePool := range in.NodePools {
		create.NodePools[i] = &godo.KubernetesNodePoolCreateRequest{
			Size:      nodePool.Size,
			Name:      nodePool.Name,
			Count:     nodePool.Count,
			Tags:      nodePool.Tags,
			Labels:    nodePool.Labels,
			AutoScale: nodePool.AutoScale,
			MinNodes:  nodePool.MinNodes,
			MaxNodes:  nodePool.MaxNodes,
		}

		create.NodePools[i].Taints = make([]godo.Taint, len(nodePool.Taints))
		for taintIndex, taint := range nodePool.Taints {
			create.NodePools[i].Taints[taintIndex] = godo.Taint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			}
		}
	}
}

func getDayFromParam(day string) godo.KubernetesMaintenancePolicyDay {
	switch day {
	case "monday":
		return godo.KubernetesMaintenanceDayMonday
	case "tuesday":
		return godo.KubernetesMaintenanceDayTuesday
	case "wednesday":
		return godo.KubernetesMaintenanceDayWednesday
	case "thursday":
		return godo.KubernetesMaintenanceDayThursday
	case "friday":
		return godo.KubernetesMaintenanceDayFriday
	case "saturday":
		return godo.KubernetesMaintenanceDaySaturday
	case "sunday":
		return godo.KubernetesMaintenanceDaySunday
	default:
		return godo.KubernetesMaintenanceDayAny
	}
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied DOKubernetesClusterParameters that are set (i.e. non-zero) on the supplied
// Kubernetes Cluster.
func LateInitializeSpec(p *v1alpha1.DOKubernetesClusterParameters, observed godo.KubernetesCluster) {
	p.VPCUUID = do.LateInitializeString(p.VPCUUID, observed.VPCUUID)
	p.Tags = do.LateInitializeStringSlice(p.Tags, observed.Tags)
	p.AutoUpgrade = do.LateInitializeBool(p.AutoUpgrade, observed.AutoUpgrade)
	p.SurgeUpgrade = do.LateInitializeBool(p.SurgeUpgrade, observed.SurgeUpgrade)
	p.HighlyAvailable = do.LateInitializeBool(p.HighlyAvailable, observed.HA)
}

// IgnoreNotFound checks for response of DigitalOcean GET API call
// and the content of returned error to ignore it if the response
// is a '404 not found' error otherwise bubble up the error.
func IgnoreNotFound(err error, response *godo.Response) error {
	if err != nil && err.Error() == "databaseID is invalid because cannot be less than 1" {
		return nil
	}
	if response != nil && response.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}
