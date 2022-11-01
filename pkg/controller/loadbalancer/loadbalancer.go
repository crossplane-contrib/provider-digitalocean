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

package loadbalancer

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

	"github.com/crossplane-contrib/provider-digitalocean/apis/loadbalancer/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	dolb "github.com/crossplane-contrib/provider-digitalocean/pkg/clients/loadbalancer"
)

const (
	// Error strings.
	errNotLB = "managed resource is not a LoadBalander resource"
	errGetLB = "cannot get a loadbalancer"

	errLBCreateFailed = "creation of LoadBalancer resource has failed"
	errLBDeleteFailed = "deletion of LoadBalancer resource has failed"
	errLBUpdate       = "cannot update managed LoadBalancer resource"
)

// SetupLB adds a controller that reconciles LB managed
// resources.
func SetupLB(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.LBGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.LB{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.LBGroupVersionKind),
			managed.WithExternalConnecter(&lbConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type lbConnector struct {
	kube client.Client
}

func (c *lbConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	token, err := do.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	client := godo.NewFromToken(token)
	godo.SetUserAgent("crossplane")(client) //nolint:errcheck
	return &lbExternal{Client: client, kube: c.kube}, nil
}

type lbExternal struct {
	kube client.Client
	*godo.Client
}

func (c *lbExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.LB)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotLB)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observed, response, err := c.LoadBalancers.Get(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetLB)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	dolb.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errLBUpdate)
		}
	}

	cr.Status.AtProvider = v1alpha1.LBObservation{
		CreationTimestamp: observed.Created,
		ID:                observed.ID,
		Status:            observed.Status,
	}

	switch cr.Status.AtProvider.Status {
	case v1alpha1.StatusNew:
		cr.SetConditions(xpv1.Creating())
	case v1alpha1.StatusActive:
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *lbExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.LB)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotLB)
	}

	cr.Status.SetConditions(xpv1.Creating())

	name := meta.GetExternalName(cr)
	if meta.GetExternalName(cr) == "" {
		name = cr.GetName()
	}

	create := &godo.LoadBalancerRequest{}
	dolb.GenerateLoadBalancer(name, cr.Spec.ForProvider, create)

	lb, _, err := c.LoadBalancers.Create(ctx, create)
	if err != nil || lb == nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errLBCreateFailed)
	}

	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, lb.ID)
	}

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *lbExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Droplets cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (c *lbExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.LB)
	if !ok {
		return errors.New(errNotLB)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := c.LoadBalancers.Delete(ctx, cr.Status.AtProvider.ID)
	return errors.Wrap(do.IgnoreNotFound(err, response), errLBDeleteFailed)
}
