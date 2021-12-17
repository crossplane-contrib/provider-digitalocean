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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	dok8s "github.com/crossplane-contrib/provider-digitalocean/pkg/clients/kubernetes"
)

const (
	// Error strings.
	errNotK8s          = "managed resource is not a DOKubernetesCluster resource"
	errGetK8s          = "cannot get a DOKubernetesCluster"
	errK8sNameRequired = "name of DOKubernetesCluster is required"

	errK8sCreateFailed = "creation of DOKubernetesCluster resource has failed"
	errK8sDeleteFailed = "deletion of DOKubernetesCluster resource has failed"
	errK8sUpdate       = "cannot update managed DOKubernetesCluster resource"
)

// SetupKubernetesCluster adds a controller that reconciles DOKubernetesCluster managed
// resources.
func SetupKubernetesCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DOKubernetesClusterKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.DOKubernetesCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DOKubernetesClusterGroupVersionKind),
			managed.WithExternalConnecter(&k8sConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type k8sConnector struct {
	kube client.Client
}

func (c *k8sConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	token, err := do.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	client := godo.NewFromToken(token)
	return &k8sExternal{Client: client, kube: c.kube}, nil
}

type k8sExternal struct {
	kube client.Client
	*godo.Client
}

func (c *k8sExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DOKubernetesCluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotK8s)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observed, response, err := c.Kubernetes.Get(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(dok8s.IgnoreNotFound(err, response), errGetK8s)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	dok8s.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errK8sUpdate)
		}
	}

	cr.Status.AtProvider = v1alpha1.DOKubernetesClusterObservation{
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
			State:   string(observed.Status.State),
			Message: observed.Status.Message,
		},
		CreatedAt:       observed.CreatedAt.String(),
		UpdatedAt:       observed.UpdatedAt.String(),
		SurgeUpgrade:    observed.SurgeUpgrade,
		HighlyAvailable: observed.HA,
		RegistryEnabled: observed.RegistryEnabled,
	}

	cr.Status.AtProvider.NodePools = make([]v1alpha1.KubernetesNodePoolObservation, len(observed.NodePools))
	for i, nodePool := range observed.NodePools {
		cr.Status.AtProvider.NodePools[i] = v1alpha1.KubernetesNodePoolObservation{
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

		cr.Status.AtProvider.NodePools[i].Taints = make([]v1alpha1.KubernetesNodePoolTaint, len(nodePool.Taints))
		for taintIndex, taint := range nodePool.Taints {
			cr.Status.AtProvider.NodePools[i].Taints[taintIndex] = v1alpha1.KubernetesNodePoolTaint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			}
		}

		cr.Status.AtProvider.NodePools[i].Nodes = make([]v1alpha1.KubernetesNode, len(nodePool.Nodes))
		for nodeIndex, node := range nodePool.Nodes {
			cr.Status.AtProvider.NodePools[i].Nodes[nodeIndex] = v1alpha1.KubernetesNode{
				ID:   node.ID,
				Name: node.Name,
				Status: v1alpha1.KubernetesStatus{
					State:   node.Status.State,
					Message: node.Status.Message,
				},
				DropletID: node.DropletID,
				CreatedAt: node.CreatedAt.String(),
				UpdatedAt: node.UpdatedAt.String(),
			}
		}
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *k8sExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DOKubernetesCluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotK8s)
	}

	cr.Status.SetConditions(xpv1.Creating())

	create := &godo.KubernetesClusterCreateRequest{}
	name := ""
	if meta.GetExternalName(cr) != "" {
		name = meta.GetExternalName(cr)
	} else {
		name = cr.GetName()
	}

	if name == "" {
		return managed.ExternalCreation{}, errors.New(errK8sNameRequired)
	}

	dok8s.GenerateKubernetes(name, cr.Spec.ForProvider, create)

	k8s, _, err := c.Kubernetes.Create(ctx, create)
	if err != nil || k8s == nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errK8sCreateFailed)
	}

	meta.SetExternalName(cr, k8s.ID)
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *k8sExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Droplets cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (c *k8sExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DOKubernetesCluster)
	if !ok {
		return errors.New(errNotK8s)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := c.Databases.Delete(ctx, cr.Status.AtProvider.ID)
	return errors.Wrap(dok8s.IgnoreNotFound(err, response), errK8sDeleteFailed)
}
