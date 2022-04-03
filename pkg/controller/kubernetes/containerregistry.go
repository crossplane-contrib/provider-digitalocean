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

	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	dok8s "github.com/crossplane-contrib/provider-digitalocean/pkg/clients/kubernetes"
)

const (
	// Error strings.
	errNotContainerRegistry             = "managed resource is not a DOContainerRegistry resource"
	errGetContainerRegistry             = "cannot get DOContainerRegistry"
	errGetContainerRegistrySubscription = "cannot get DOContainerRegistry subscription"

	errContainerRegistryCreateFailed = "creation of DOContainerRegistry resource has failed"
	errContainerRegistryDeleteFailed = "deletion of DOContainerRegistry resource has failed"
	errContainerRegistryUpdate       = "cannot update managed DOContainerRegistry resource"

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
	return &containerRegistryExternal{client: client.Registry, kube: c.kube}, nil
}

type containerRegistryExternal struct {
	kube   client.Client
	client dok8s.RegistryClient
}

func (c *containerRegistryExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotContainerRegistry)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observed, response, err := c.client.Get(ctx)
	if err != nil {
		cr.Status.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetContainerRegistry)
	}

	cr.Status.SetConditions(xpv1.Available())

	subscription, response, err := c.client.GetSubscription(ctx)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetContainerRegistrySubscription)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	dok8s.RegistryLateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errContainerRegistryUpdate)
		}
	}

	cr.Status.AtProvider = dok8s.GenerateContainerRegistryObservation(observed, subscription)

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
		return managed.ExternalCreation{}, errors.New(errNotContainerRegistry)
	}

	cr.Status.SetConditions(xpv1.Creating())

	name := meta.GetExternalName(cr)
	if name == "" {
		name = cr.GetName()
	}

	create := &godo.RegistryCreateRequest{}
	dok8s.GenerateContainerRegistry(name, cr.Spec.ForProvider, create)

	containerRegistry, _, err := c.client.Create(ctx, create)
	if err != nil || containerRegistry == nil {
		err = errors.Wrap(err, errContainerRegistryCreateFailed)
		cr.Status.SetConditions(xpv1.ReconcileError(err))
		return managed.ExternalCreation{}, err
	}

	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, containerRegistry.Name)
	}

	return managed.ExternalCreation{}, nil
}

func (c *containerRegistryExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotContainerRegistry)
	}
	update := &godo.RegistrySubscriptionUpdateRequest{TierSlug: cr.Spec.ForProvider.SubscriptionTier}

	containerRegistry, _, err := c.client.UpdateSubscription(ctx, update)
	if err != nil || containerRegistry == nil {
		err = errors.Wrap(err, errContainerRegistryUpdate)
		cr.Status.SetConditions(xpv1.ReconcileError(err))
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

func (c *containerRegistryExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DOContainerRegistry)
	if !ok {
		return errors.New(errNotContainerRegistry)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := c.client.Delete(ctx)
	if err != nil {
		err = errors.Wrap(do.IgnoreNotFound(err, response), errContainerRegistryDeleteFailed)
		cr.Status.SetConditions(xpv1.ReconcileError(err))
		return err

	}
	return nil
}
