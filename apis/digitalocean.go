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

// Package apis contains Kubernetes API for the DigitalOcean provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	computev1alpha1 "github.com/crossplane-contrib/provider-digitalocean/apis/compute/v1alpha1"
	dbv1alpha1 "github.com/crossplane-contrib/provider-digitalocean/apis/database/v1alpha1"
	kubev1alpha1 "github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	lbv1alpha1 "github.com/crossplane-contrib/provider-digitalocean/apis/loadbalancer/v1alpha1"
	storv1alpha1 "github.com/crossplane-contrib/provider-digitalocean/apis/storage/v1alpha1"
	dov1alpha1 "github.com/crossplane-contrib/provider-digitalocean/apis/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		dov1alpha1.SchemeBuilder.AddToScheme,
		computev1alpha1.SchemeBuilder.AddToScheme,
		dbv1alpha1.SchemeBuilder.AddToScheme,
		kubev1alpha1.SchemeBuilder.AddToScheme,
		lbv1alpha1.SchemeBuilder.AddToScheme,
		storv1alpha1.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
