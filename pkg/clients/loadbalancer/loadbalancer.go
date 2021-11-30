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
	"net/http"

	"github.com/digitalocean/godo"

	"github.com/crossplane-contrib/provider-digitalocean/apis/loadbalancer/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
)

// GenerateLoadBalancer generates *godo.LoadBalancerRequest instance from LBParameters.
func GenerateLoadBalancer(name string, in v1alpha1.LBParameters, create *godo.LoadBalancerRequest) {
	create.Name = name
	create.Region = in.Region
	create.Algorithm = in.Algorithm
	create.ForwardingRules = append(create.ForwardingRules, generateForwardRule(in.Port))
	create.HealthCheck = generateHealthCheck(in.HealthCheck, in.Port)
	create.Tags = in.Tags
	create.VPCUUID = do.StringValue(in.VPCUUID)
}

func generateForwardRule(param int) godo.ForwardingRule {
	if param != 0 {
		return godo.ForwardingRule{
			EntryProtocol:  "tcp",
			EntryPort:      param,
			TargetProtocol: "tcp",
			TargetPort:     param,
		}
	}

	return godo.ForwardingRule{
		EntryProtocol:  "tcp",
		EntryPort:      80,
		TargetProtocol: "tcp",
		TargetPort:     80,
	}
}

func generateHealthCheck(in v1alpha1.DOLoadBalancerHealthCheck, inPort int) *godo.HealthCheck {
	port := 80
	if inPort != 0 {
		port = inPort
	}
	return &godo.HealthCheck{
		Protocol:               "tcp",
		Port:                   port,
		CheckIntervalSeconds:   in.Interval,
		ResponseTimeoutSeconds: in.Timeout,
		UnhealthyThreshold:     in.UnhealthyThreshold,
		HealthyThreshold:       in.HealthyThreshold,
	}
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied LBParameters that are set (i.e. non-zero) on the supplied
// LB.
func LateInitializeSpec(p *v1alpha1.LBParameters, observed godo.LoadBalancer) {
	p.Tags = do.LateInitializeStringSlice(p.Tags, observed.Tags)
	p.VPCUUID = do.LateInitializeString(p.VPCUUID, observed.VPCUUID)
}

// IgnoreNotFound checks for response of DigitalOcean GET API call
// and the content of returned error to ignore it if the response
// is a '404 not found' error otherwise bubble up the error.
func IgnoreNotFound(err error, response *godo.Response) error {
	if err != nil && err.Error() == "lbID is invalid because cannot be less than 1" {
		return nil
	}
	if response != nil && response.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}
