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
	errFetchingConfig  = "fetching of DOKubernetesCluster Kubeconfig has failed"
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
	godo.SetUserAgent("crossplane")(client) //nolint:errcheck
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
		return managed.ExternalObservation{}, errors.Wrap(do.IgnoreNotFound(err, response), errGetK8s)
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	dok8s.LateInitializeSpec(&cr.Spec.ForProvider, *observed)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errK8sUpdate)
		}
	}

	cr.Status.AtProvider = dok8s.GenerateObservation(observed)
	dok8s.SetCondition(cr)

	extObs := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}

	if cr.Spec.WriteConnectionSecretToReference != nil {
		config, resp, err := c.Kubernetes.GetKubeConfig(ctx, observed.ID)

		if err != nil || resp.StatusCode >= 300 {
			return managed.ExternalObservation{}, errors.Wrap(err, errFetchingConfig)
		}

		extObs.ConnectionDetails = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretKubeconfigKey: config.KubeconfigYAML,
		}
	}

	return extObs, nil
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

	return managed.ExternalCreation{}, nil
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

	response, err := c.Kubernetes.Delete(ctx, cr.Status.AtProvider.ID)
	return errors.Wrap(do.IgnoreNotFound(err, response), errK8sDeleteFailed)
}
