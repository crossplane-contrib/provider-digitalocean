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

package controller

import (
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-digitalocean/pkg/controller/compute"
	"github.com/crossplane-contrib/provider-digitalocean/pkg/controller/config"
	"github.com/crossplane-contrib/provider-digitalocean/pkg/controller/database"
	"github.com/crossplane-contrib/provider-digitalocean/pkg/controller/kubernetes"
	"github.com/crossplane-contrib/provider-digitalocean/pkg/controller/loadbalancer"
)

// Setup creates all DigitalOcean controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		config.Setup,
		compute.SetupDroplet,
		database.SetupDatabase,
		kubernetes.SetupKubernetesCluster,
		kubernetes.SetupDOContainerRegistry,
		loadbalancer.SetupLB,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}
	return nil
}
