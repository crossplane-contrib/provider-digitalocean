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

package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	"github.com/crossplane-contrib/provider-digitalocean/apis/storage/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
	dospace "github.com/crossplane-contrib/provider-digitalocean/pkg/clients/storage"
)

const (
	errNotSpace               = "managed resource is not a Spaces resource"
	errBuildingS3Config       = "there was a problem creating the s3 configuration for talking to Spaces"
	errSpacesCreateFailed     = "creation of Spaces bucket has failed"
	errSpacesListBucketFailed = "listing of Spaces buckets for observation failed"
	errSpacesDeleteFailed     = "deletion of Spaces bucket has failed"
	errSpacesUpdate           = "cannot update managed Spaces bucket resource"
)

type spacesConnector struct {
	kube client.Client
}

type spacesExternal struct {
	kube client.Client

	// We need these because we'll have to create a new s3 client on every request and change the endpoint based on where we're going.
	accessKey       string
	accessSecretKey string
}

func (c *spacesConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	accessKey, accessSecretKey, err := do.GetS3AuthInfo(ctx, c.kube, mg)

	if err != nil {
		return nil, err
	}

	return &spacesExternal{kube: c.kube, accessKey: accessKey, accessSecretKey: accessSecretKey}, nil
}

// SetupSpaces adds a controller that reconciles DOSpaces managed resources.
func SetupSpaces(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.DOSpaceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.DOSpace{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.DOSpaceGroupVersionKind),
			managed.WithExternalConnecter(&spacesConnector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

func (c *spacesExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.DOSpace)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSpace)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	config, err := c.createS3ConfigForRegion(ctx, cr.Spec.ForProvider.Region)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errBuildingS3Config)
	}

	client := s3.NewFromConfig(*config)

	// Empty input because it accepts nothing
	output, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil || output == nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errSpacesListBucketFailed)
	}

	obsBucket := getBucket(meta.GetExternalName(cr), output.Buckets)

	if obsBucket == nil {
		return managed.ExternalObservation{}, nil
	}

	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errSpacesUpdate)
		}
	}

	cr.Status.AtProvider = v1alpha1.DOSpaceObservation{
		Name:         *obsBucket.Name,
		CreationDate: obsBucket.CreationDate.String(),
	}

	cr.Status.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (c *spacesExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.DOSpace)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSpace)
	}

	cr.Status.SetConditions(xpv1.Creating())

	config, err := c.createS3ConfigForRegion(ctx, cr.Spec.ForProvider.Region)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errBuildingS3Config)
	}

	client := s3.NewFromConfig(*config)

	name := ""
	if meta.GetExternalName(cr) != "" {
		name = meta.GetExternalName(cr)
	} else {
		name = cr.GetName()
	}

	create := &s3.CreateBucketInput{}
	dospace.GenerateSpace(name, cr.Spec.ForProvider, create)

	output, err := client.CreateBucket(ctx, create)
	if err != nil || output == nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSpacesCreateFailed)
	}

	meta.SetExternalName(cr, *output.Location)

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (c *spacesExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Buckets cannot be updated at this time
	// Coming to a provider near you soon
	return managed.ExternalUpdate{}, nil
}

func (c *spacesExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.DOSpace)
	if !ok {
		return errors.New(errNotSpace)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	config, err := c.createS3ConfigForRegion(ctx, cr.Spec.ForProvider.Region)
	if err != nil {
		return errors.Wrap(err, errBuildingS3Config)
	}

	client := s3.NewFromConfig(*config)
	delete := &s3.DeleteBucketInput{
		Bucket: &cr.Status.AtProvider.Name,
	}
	output, err := client.DeleteBucket(ctx, delete)
	if err != nil || output == nil {
		return errors.Wrap(err, errSpacesDeleteFailed)
	}

	return nil
}

func getBucket(name string, buckets []types.Bucket) *types.Bucket {
	found := false
	var obsBucket types.Bucket
	for _, bucket := range buckets {
		if *bucket.Name == name {
			found = true
			obsBucket = bucket
		}
	}

	if found {
		return &obsBucket
	}

	return nil
}

func (c *spacesExternal) createS3ConfigForRegion(ctx context.Context, doRegion string) (*aws.Config, error) {

	doEndpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.digitaloceanspaces.com", doRegion),
		}, nil
	})

	doCredsProvider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     c.accessKey,
			SecretAccessKey: c.accessSecretKey,
		}, nil
	})

	config, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(doEndpointResolver),
		config.WithCredentialsProvider(doCredsProvider),
	)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
