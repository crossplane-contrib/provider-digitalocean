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

package compute

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

	"github.com/crossplane-contrib/provider-digitalocean/apis/compute/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	docompute "github.com/crossplane-contrib/provider-digitalocean/pkg/clients/compute"
)

const (
	// Error strings.
	errNotDroplet = "managed resource is not a Droplet resource"
	errGetDroplet = "cannot get droplet"

	errDropletCreateFailed = "creation of Droplet resource has failed"
	errDropletDeleteFailed = "deletion of Droplet resource has failed"
	errDropletUpdate       = "cannot update managed Droplet resource"
)

// SetupDroplet adds a controller that reconciles Droplet managed
// resources.
func SetupDroplet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DropletGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Droplet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DropletGroupVersionKind),
			managed.WithExternalConnecter(&dropletConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type dropletConnector struct {
	kube client.Client
}

func (c *dropletConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	token, err := do.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	client := godo.NewFromToken(token)
	godo.SetUserAgent("crossplane")(client) //nolint:errcheck
	return &dropletExternal{Client: client, kube: c.kube}, nil
}

type dropletExternal struct {
	kube client.Client
	*godo.Client
}

func (c *dropletExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Droplet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDroplet)
	}
	if cr.Status.AtProvider.ID == 0 {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	observed, response, err := c.Droplets.Get(ctx, cr.Status.AtProvider.ID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetDroplet)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	docompute.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	observedPrivateIPv4, _ := observed.PrivateIPv4()
	observedPublicIPv4, _ := observed.PublicIPv4()

	cr.Status.AtProvider = v1alpha1.DropletObservation{
		CreationTimestamp: observed.Created,
		ID:                observed.ID,
		PrivateIPv4:       observedPrivateIPv4,
		PublicIPv4:        observedPublicIPv4,
		Region:            observed.Region.Slug,
		Size:              observed.SizeSlug,
		Status:            observed.Status,
	}
	if err := c.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errDropletUpdate)
	}

	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errDropletUpdate)
		}
	}

	switch cr.Status.AtProvider.Status {
	case v1alpha1.StatusNew:
		cr.SetConditions(xpv1.Creating())
	case v1alpha1.StatusActive:
		cr.SetConditions(xpv1.Available())
	}

	if currentSpec.ProjectID != nil {
		page := 0
		for {
			projectResources, _, err := c.Projects.ListResources(ctx, *currentSpec.ProjectID, &godo.ListOptions{Page: page})
			if err != nil {
				return managed.ExternalObservation{}, err
			}
			if len(projectResources) == 0 {
				return managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				}, nil
			}
			for _, projectResource := range projectResources {
				if projectResource.URN == observed.URN() {
					break
				}
			}
		}
	}

	// Droplets are always "up to date" because they can't be updated. ¯\_(ツ)_/¯
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *dropletExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Droplet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDroplet)
	}

	cr.Status.SetConditions(xpv1.Creating())

	name := meta.GetExternalName(cr)

	create := &godo.DropletCreateRequest{}
	docompute.GenerateDroplet(name, cr.Spec.ForProvider, create)

	droplet, _, err := c.Droplets.Create(ctx, create)
	if err != nil || droplet == nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errDropletCreateFailed)
	}

	cr.Status.AtProvider = v1alpha1.DropletObservation{
		CreationTimestamp: droplet.Created,
		ID:                droplet.ID,
		Status:            droplet.Status,
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()

	if currentSpec.ProjectID != nil {
		_, _, err := c.Projects.AssignResources(ctx, *currentSpec.ProjectID, droplet)
		if err != nil {
			return managed.ExternalCreation{}, err
		}
	}

	if err := c.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errDropletUpdate)
	}

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *dropletExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Droplet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDroplet)
	}
	observed, response, err := c.Droplets.Get(ctx, cr.Status.AtProvider.ID)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetDroplet)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()

	if currentSpec.ProjectID != nil {
		_, _, err := c.Projects.AssignResources(ctx, *currentSpec.ProjectID, observed)
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	// Droplets cannot be updated.
	return managed.ExternalUpdate{}, nil
}

func (c *dropletExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Droplet)
	if !ok {
		return errors.New(errNotDroplet)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	response, err := c.Droplets.Delete(ctx, cr.Status.AtProvider.ID)
	return errors.Wrap(do.IgnoreNotFound(err, response), errDropletDeleteFailed)
}
