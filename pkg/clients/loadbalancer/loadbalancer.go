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
func GenerateLoadBalancer(name string, in v1alpha1.DOLoadBalancerParameters, create *godo.LoadBalancerRequest) {
	create.Name = name
	create.Region = in.Region
	create.Algorithm = do.StringValue(in.Algorithm)
	create.ForwardingRules = generateForwardRule(in.ForwardingRules)
	create.HealthCheck = generateHealthCheck(in.HealthCheck)
	create.Tags = in.Tags
	create.VPCUUID = do.StringValue(in.VPCUUID)
}

func generateForwardRule(in []*v1alpha1.DOForwardingRule) []godo.ForwardingRule {
	result := make([]godo.ForwardingRule, 0, len(in))
	for _, fdRule := range in {
		result = append(result, godo.ForwardingRule{
			EntryProtocol:  fdRule.EntryProtocol,
			EntryPort:      fdRule.EntryPort,
			TargetProtocol: fdRule.TargetProtocol,
			TargetPort:     fdRule.TargetPort,
			CertificateID:  fdRule.CertificateID,
			TlsPassthrough: fdRule.TLSPassthrough,
		})
	}

	return result
}

func generateHealthCheck(in *v1alpha1.DOLoadBalancerHealthCheck) *godo.HealthCheck {
	return &godo.HealthCheck{
		Protocol:               do.StringValue(in.Protocol),
		Port:                   do.IntValue(in.Port),
		CheckIntervalSeconds:   do.IntValue(in.CheckIntervalSeconds),
		ResponseTimeoutSeconds: do.IntValue(in.ResponseTimeoutSeconds),
		UnhealthyThreshold:     do.IntValue(in.UnhealthyThreshold),
		HealthyThreshold:       do.IntValue(in.HealthyThreshold),
	}
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied LBParameters that are set (i.e. non-zero) on the supplied
// LB.
func LateInitializeSpec(p *v1alpha1.DOLoadBalancerParameters, observed godo.LoadBalancer) {
	p.Algorithm = do.LateInitializeString(p.Algorithm, observed.Algorithm)
	p.Tags = do.LateInitializeStringSlice(p.Tags, observed.Tags)
	p.VPCUUID = do.LateInitializeString(p.VPCUUID, observed.VPCUUID)

	if len(p.ForwardingRules) == 0 && len(observed.ForwardingRules) != 0 {
		p.ForwardingRules = make([]*v1alpha1.DOForwardingRule, len(observed.ForwardingRules))
		for i, val := range observed.ForwardingRules {
			p.ForwardingRules[i] = &v1alpha1.DOForwardingRule{
				EntryProtocol:  val.EntryProtocol,
				EntryPort:      val.EntryPort,
				TargetProtocol: val.TargetProtocol,
				TargetPort:     val.TargetPort,
				CertificateID:  val.CertificateID,
				TLSPassthrough: val.TlsPassthrough,
			}
		}
	}

	if observed.HealthCheck != nil {
		p.HealthCheck.HealthyThreshold = do.LateInitializeInt(
			p.HealthCheck.HealthyThreshold,
			observed.HealthCheck.HealthyThreshold)
		p.HealthCheck.CheckIntervalSeconds = do.LateInitializeInt(
			p.HealthCheck.CheckIntervalSeconds,
			observed.HealthCheck.CheckIntervalSeconds)
		p.HealthCheck.UnhealthyThreshold = do.LateInitializeInt(
			p.HealthCheck.UnhealthyThreshold,
			observed.HealthCheck.UnhealthyThreshold)
		p.HealthCheck.ResponseTimeoutSeconds = do.LateInitializeInt(
			p.HealthCheck.ResponseTimeoutSeconds,
			observed.HealthCheck.ResponseTimeoutSeconds)
		p.HealthCheck.Port = do.LateInitializeInt(
			p.HealthCheck.Port,
			observed.HealthCheck.Port)
		p.HealthCheck.Protocol = do.LateInitializeString(
			p.HealthCheck.Protocol,
			observed.HealthCheck.Protocol)
	}
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
