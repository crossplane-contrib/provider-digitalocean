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
	"github.com/digitalocean/godo"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

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

// GenerateObservation generates a DOKubernetesClusterObservation from a given observed state from godo
func GenerateObservation(observed *godo.KubernetesCluster) v1alpha1.DOKubernetesClusterObservation {
	observation := v1alpha1.DOKubernetesClusterObservation{
		ID:            observed.ID,
		Name:          observed.Name,
		Region:        observed.RegionSlug,
		Version:       observed.VersionSlug,
		ClusterSubnet: observed.ClusterSubnet,
		ServiceSubnet: observed.ServiceSubnet,
		VPCUUID:       observed.VPCUUID,
		IPV4:          observed.IPv4,
		Endpoint:      observed.Endpoint,
		Tags:          observed.Tags,
		MaintenancePolicy: v1alpha1.KubernetesClusterMaintenancePolicyObservation{
			Policy: v1alpha1.KubernetesClusterMaintenancePolicy{
				StartTime: observed.MaintenancePolicy.StartTime,
				Day:       observed.MaintenancePolicy.Day.String(),
			},
			Duration: observed.MaintenancePolicy.Duration,
		},
		AutoUpgrade: observed.AutoUpgrade,
		Status: v1alpha1.KubernetesStatus{
			State:   getStateFromGodoState(observed.Status.State),
			Message: observed.Status.Message,
		},
		CreatedAt:       observed.CreatedAt.String(),
		UpdatedAt:       observed.UpdatedAt.String(),
		SurgeUpgrade:    observed.SurgeUpgrade,
		HighlyAvailable: observed.HA,
		RegistryEnabled: observed.RegistryEnabled,
	}

	observation.NodePools = make([]v1alpha1.KubernetesNodePoolObservation, len(observed.NodePools))
	for i, nodePool := range observed.NodePools {
		observation.NodePools[i] = v1alpha1.KubernetesNodePoolObservation{
			ID:        nodePool.ID,
			Size:      nodePool.Size,
			Name:      nodePool.Name,
			Count:     nodePool.Count,
			Tags:      nodePool.Tags,
			Labels:    nodePool.Labels,
			AutoScale: nodePool.AutoScale,
			MinNodes:  nodePool.MinNodes,
			MaxNodes:  nodePool.MaxNodes,
		}

		observation.NodePools[i].Taints = make([]v1alpha1.KubernetesNodePoolTaint, len(nodePool.Taints))
		for taintIndex, taint := range nodePool.Taints {
			observation.NodePools[i].Taints[taintIndex] = v1alpha1.KubernetesNodePoolTaint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			}
		}

		observation.NodePools[i].Nodes = make([]v1alpha1.KubernetesNode, len(nodePool.Nodes))
		for nodeIndex, node := range nodePool.Nodes {
			observation.NodePools[i].Nodes[nodeIndex] = v1alpha1.KubernetesNode{
				ID:   node.ID,
				Name: node.Name,
				Status: v1alpha1.KubernetesStatus{
					State:   getStateFromString(node.Status.State),
					Message: node.Status.Message,
				},
				DropletID: node.DropletID,
				CreatedAt: node.CreatedAt.String(),
				UpdatedAt: node.UpdatedAt.String(),
			}
		}
	}

	return observation
}

func getStateFromGodoState(state godo.KubernetesClusterStatusState) v1alpha1.KubernetesState {
	return getStateFromString(string(state))
}

func getStateFromString(state string) v1alpha1.KubernetesState {
	switch state {
	case "running":
		return v1alpha1.KubernetesStateRunning
	case "provisioning":
		return v1alpha1.KubernetesStateProvisioning
	case "degraded":
		return v1alpha1.KubernetesStateDegraded
	case "error":
		return v1alpha1.KubernetesStateError
	case "deleted":
		return v1alpha1.KubernetesStateDeleted
	case "upgrading":
		return v1alpha1.KubernetesStateUpgrading
	case "deleting":
		return v1alpha1.KubernetesStateDeleting
	default:
		return v1alpha1.KubernetesStateError // Just return an error if we can't find the state
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

// SetCondition sets the condition for a DOKubernetesCluster resource from its state
func SetCondition(cr *v1alpha1.DOKubernetesCluster) {
	switch cr.Status.AtProvider.Status.State {
	case v1alpha1.KubernetesStateProvisioning:
		cr.Status.SetConditions(xpv1.Creating())
	case v1alpha1.KubernetesStateRunning:
		fallthrough
	case v1alpha1.KubernetesStateDegraded: // Still available just in a poor state
		cr.Status.SetConditions(xpv1.Available())
	case v1alpha1.KubernetesStateDeleting:
		fallthrough
	case v1alpha1.KubernetesStateDeleted:
		cr.Status.SetConditions(xpv1.Deleting())
	case v1alpha1.KubernetesStateError:
		fallthrough
	case v1alpha1.KubernetesStateUpgrading:
		cr.Status.SetConditions(xpv1.Unavailable())
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
