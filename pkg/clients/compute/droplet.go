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
	"strconv"

	"github.com/digitalocean/godo"

	"github.com/khos2ow/provider-digitalocean/apis/compute/v1alpha1"
	do "github.com/khos2ow/provider-digitalocean/pkg/clients"
)

// GenerateDroplet generates *godo.DropletCreateRequest instance from DropletParameters.
// nolint:gocyclo
func GenerateDroplet(name string, in v1alpha1.DropletParameters, create *godo.DropletCreateRequest) {
	create.Name = name
	create.Region = in.Region
	create.Size = in.Size

	create.Image = godo.DropletCreateImage{}

	if imageID, err := strconv.Atoi(in.Image); err == nil {
		create.Image.ID = imageID
	} else {
		create.Image.Slug = in.Image
	}

	if len(in.SSHKeys) > 0 {
		keys := make([]godo.DropletCreateSSHKey, len(in.Volumes))
		for _, k := range in.Volumes {
			if id, err := strconv.Atoi(k); err == nil {
				keys = append(keys, godo.DropletCreateSSHKey{ID: id})
			} else {
				keys = append(keys, godo.DropletCreateSSHKey{Fingerprint: k})
			}
		}
		create.SSHKeys = keys
	}

	if in.Backups != nil {
		create.Backups = *in.Backups
	}

	if in.IPv6 != nil {
		create.IPv6 = *in.IPv6
	}

	if in.PrivateNetworking != nil {
		create.PrivateNetworking = *in.PrivateNetworking
	}

	if in.Monitoring != nil {
		create.Monitoring = *in.Monitoring
	}

	if len(in.Volumes) > 0 {
		volumes := make([]godo.DropletCreateVolume, len(in.Volumes))
		for _, v := range in.Volumes {
			if v == "" {
				continue
			}
			volumes = append(volumes, godo.DropletCreateVolume{ID: v})
		}
		create.Volumes = volumes
	}

	if len(in.Tags) > 0 {
		create.Tags = in.Tags
	}

	if in.VPCUUID != nil {
		create.VPCUUID = *in.VPCUUID
	}
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied DropletParameters that are set (i.e. non-zero) on the supplied
// Droplet.
func LateInitializeSpec(p *v1alpha1.DropletParameters, observed godo.Droplet) {
	p.Volumes = do.LateInitializeStringSlice(p.Volumes, observed.VolumeIDs)
	p.Tags = do.LateInitializeStringSlice(p.Tags, observed.Tags)
	p.VPCUUID = do.LateInitializeString(p.VPCUUID, observed.VPCUUID)
}
