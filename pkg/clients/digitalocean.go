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

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/khos2ow/provider-digitalocean/apis/v1alpha1"
)

// GetAuthInfo returns the necessary authentication information that is necessary
// to use when the controller connects to DigitalOcean API in order to reconcile
// the managed resource.
func GetAuthInfo(ctx context.Context, c client.Client, mg resource.Managed) (token string, err error) {
	pc := &v1alpha1.ProviderConfig{}
	t := resource.NewProviderConfigUsageTracker(c, &v1alpha1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return "", err
	}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return "", err
	}

	// NOTE(muvaf): When we implement the workload identity, we will only need to
	// return a different type of option.ClientOption, which is WithTokenSource().
	if s := pc.Spec.Credentials.Source; s != runtimev1alpha1.CredentialsSourceSecret {
		return "", errors.Errorf("unsupported credentials source %q", s)
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
