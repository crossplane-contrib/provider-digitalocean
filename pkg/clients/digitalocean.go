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

package clients

import (
	"context"
	"net/http"
	"strings"

	"github.com/digitalocean/godo"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-digitalocean/apis/v1alpha1"
)

// GetAuthInfo returns the necessary authentication information that is necessary
// to use when the controller connects to DigitalOcean API in order to reconcile
// the managed resource.
func GetAuthInfo(ctx context.Context, c client.Client, mg resource.Managed) (token string, err error) {
	pc, err := getProviderConfig(ctx, c, mg)
	if err != nil {
		return "", err
	}

	ref := pc.Spec.Credentials.SecretRef
	if ref == nil {
		return "", errors.New("no credentials secret reference was provided")
	}

	s := &v1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ref.Namespace}, s); err != nil {
		return "", err
	}
	return string(s.Data[ref.Key]), nil
}

// GetS3AuthInfo returns the access key and private access key needed to access DigitalOcean's S3 api
func GetS3AuthInfo(ctx context.Context, c client.Client, mg resource.Managed) (accessKey, privateAccessKey string, err error) {
	pc, err := getProviderConfig(ctx, c, mg)
	if err != nil {
		return "", "", err
	}

	spacesCredentials := pc.Spec.SpacesCredentials
	if spacesCredentials == nil {
		return "", "", errors.New("no spaces credentials was provided")
	}

	accessKeyRef := pc.Spec.SpacesCredentials.AccessKeyRef.SecretRef
	if accessKeyRef == nil {
		return "", "", errors.New("no spaces access key was provided")
	}

	secretKeyRef := pc.Spec.SpacesCredentials.SecretAccessKeyRef.SecretRef
	if secretKeyRef == nil {
		return "", "", errors.New("no spaces secret access key was provided")
	}

	accessKeyS := &v1.Secret{}
	secretKeyS := &v1.Secret{}

	if err := c.Get(ctx, types.NamespacedName{Name: accessKeyRef.Name, Namespace: accessKeyRef.Namespace}, accessKeyS); err != nil {
		return "", "", errors.Wrap(err, "failed to fetch s3 access key secret")
	}

	if err := c.Get(ctx, types.NamespacedName{Name: secretKeyRef.Name, Namespace: secretKeyRef.Namespace}, secretKeyS); err != nil {
		return "", "", errors.Wrap(err, "failed to fetch s3 access secret key")
	}

	return string(accessKeyS.Data[accessKeyRef.Key]), string(secretKeyS.Data[secretKeyRef.Key]), nil
}

func getProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*v1alpha1.ProviderConfig, error) {
	pc := &v1alpha1.ProviderConfig{}
	t := resource.NewProviderConfigUsageTracker(c, &v1alpha1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, err
	}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, err
	}

	// NOTE(muvaf): When we implement the workload identity, we will only need to
	// return a different type of option.ClientOption, which is WithTokenSource().
	if s := pc.Spec.Credentials.Source; s != xpv1.CredentialsSourceSecret {
		return nil, errors.Errorf("unsupported credentials source %q", s)
	}

	return pc, nil
}

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
func StringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// BoolValue converts the supplied bool pointer to an bool, returning false if
// the pointer is nil.
func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// Int64Value converts the supplied int64 pointer to an int, returning zero if
// the pointer is nil.
func Int64Value(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// IntValue converts the supplied int pointer to an int, returning zero if
// the pointer is nil.
func IntValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// LateInitialize functions initialize s(first argument), presumed to be an
// optional field of a Kubernetes API object's spec per Kubernetes
// "late initialization" semantics. s is returned unchanged if it is non-nil
// or from(second argument) is the empty string, otherwise a pointer to from
// is returned.
// https://github.com/kubernetes/community/blob/db7f270f/contributors/devel/sig-architecture/api-conventions.md#optional-vs-required
// https://github.com/kubernetes/community/blob/db7f270f/contributors/devel/sig-architecture/api-conventions.md#late-initialization
// TODO(muvaf): These functions will probably be needed by other providers.
// Consider moving them to crossplane-runtime.

// LateInitializeString implements late initialization for string type.
func LateInitializeString(s *string, from string) *string {
	if s != nil || from == "" {
		return s
	}
	return &from
}

// LateInitializeInt64 implements late initialization for int64 type.
func LateInitializeInt64(i *int64, from int64) *int64 {
	if i != nil || from == 0 {
		return i
	}
	return &from
}

// LateInitializeBool implements late initialization for bool type.
func LateInitializeBool(b *bool, from bool) *bool {
	if b != nil || !from {
		return b
	}
	return &from
}

// LateInitializeStringSlice implements late initialization for
// string slice type.
func LateInitializeStringSlice(s []string, from []string) []string {
	if len(s) != 0 || len(from) == 0 {
		return s
	}
	return from
}

// LateInitializeStringMap implements late initialization for
// string map type.
func LateInitializeStringMap(s map[string]string, from map[string]string) map[string]string {
	if len(s) != 0 || len(from) == 0 {
		return s
	}
	return from
}

// IgnoreNotFound checks for response of DigitalOcean GET API call
// and the content of returned error to ignore it if the response
// is a '404 not found' error otherwise bubble up the error.
func IgnoreNotFound(err error, response *godo.Response) error {
	if err != nil && strings.Contains(err.Error(), "is invalid because cannot be less than 1") {
		return nil
	}
	if response != nil && response.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}
