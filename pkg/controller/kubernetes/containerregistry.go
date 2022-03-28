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
	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	dok8s "github.com/crossplane-contrib/provider-digitalocean/pkg/clients/kubernetes"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/digitalocean/godo"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Error strings.
	errNotContianerRegistiry             = "managed resource is not a DOContainerRegistery resource"
	errGetContianerRegistiry             = "cannot get DOContainerRegistery"
	errGetContianerRegistirySubscription = "cannot get DOContainerRegistery subscrioption"

	errContianerRegistiryCreateFailed = "creation of DOContainerRegistery resource has failed"
	errContianerRegistiryDeleteFailed = "deletion of DOContainerRegistery resource has failed"
	errContianerRegistiryUpdate       = "cannot update managed DOContainerRegistery resource"

	subscriptionOutDated = "subscription is not up to date"
)

// SetupDOContainerRegistry adds a controller that reconciles DOContainerRegistry managed
// resources.
func SetupDOContainerRegistry(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DOContainerRegistryGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.DOContainerRegistry{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DOContainerRegistryGroupVersionKind),
			managed.WithExternalConnecter(&containerRegistryConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type containerRegistryConnector struct {
	kube client.Client
}

func (c *containerRegistryConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	token, err := do.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	client := godo.NewFromToken(token)
	return &containerRegistryExternal{Client: client, kube: c.kube}, nil
}

type containerRegistryExternal struct {
	kube client.Client
	*godo.Client
}

func (c *containerRegistryExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotContianerRegistiry)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observed, response, err := c.Registry.Get(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetContianerRegistiry)
	}

	subscription, response, err := c.Registry.GetSubscription(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetContianerRegistirySubscription)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	dok8s.RegistryLateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errContianerRegistiryUpdate)
		}
	}

	cr.Status.AtProvider = v1alpha1.DOContainerRegistryObservation{
		Name:                       observed.Name,
		Region:                     observed.Region,
		CreatedAt:                  observed.CreatedAt.String(),
		StorageUsageBytes:          observed.StorageUsageBytes,
		StorageUsageBytesUpdatedAt: observed.StorageUsageBytesUpdatedAt.String(),
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

	if cr.Spec.ForProvider.SubscriptionTier != cr.Status.AtProvider.Subscription.Tier.Slug {
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: false,
			Diff:             subscriptionOutDated,
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *containerRegistryExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotContianerRegistiry)
	}

	cr.Status.SetConditions(xpv1.Creating())

	name := meta.GetExternalName(cr)
	if meta.GetExternalName(cr) == "" {
		name = cr.GetName()
	}

	create := &godo.RegistryCreateRequest{}
	dok8s.GenerateContainerRegistry(name, cr.Spec.ForProvider, create)

	containerRegistry, _, err := c.Registry.Create(ctx, create)
	if err != nil || containerRegistry == nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errContianerRegistiryCreateFailed)
	}

	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, containerRegistry.Name)
	}

	return managed.ExternalCreation{}, nil
}

func (c *containerRegistryExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotContianerRegistiry)
	}
	update := &godo.RegistrySubscriptionUpdateRequest{TierSlug: cr.Spec.ForProvider.SubscriptionTier}

	containerRegistry, _, err := c.Registry.UpdateSubscription(ctx, update)
	if err != nil || containerRegistry == nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errContianerRegistiryUpdate)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *containerRegistryExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return errors.New(errNotContianerRegistiry)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := c.Registry.Delete(ctx)
	return errors.Wrap(do.IgnoreNotFound(err, response), errContianerRegistiryDeleteFailed)
}
