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

	"github.com/crossplane-contrib/provider-digitalocean/apis/compute/v1alpha1"
	do "github.com/crossplane-contrib/provider-digitalocean/pkg/clients"
)

// GenerateDroplet generates *godo.DropletCreateRequest instance from DropletParameters.
func GenerateDroplet(name string, in v1alpha1.DropletParameters, create *godo.DropletCreateRequest) {
	create.Name = name
	create.Region = in.Region
	create.Size = in.Size
	create.Image = generateImage(in.Image)
	create.SSHKeys = generateSSHKeys(in.SSHKeys)
	create.Backups = do.BoolValue(in.Backups)
	create.IPv6 = do.BoolValue(in.IPv6)
	create.PrivateNetworking = do.BoolValue(in.PrivateNetworking)
	create.Monitoring = do.BoolValue(in.Monitoring)
	create.Volumes = generateVolumes(in.Volumes)
	create.Tags = in.Tags
	create.VPCUUID = do.StringValue(in.VPCUUID)
	create.WithDropletAgent = in.WithDropletAgent
}

func generateImage(param string) godo.DropletCreateImage {
	image := godo.DropletCreateImage{}
	if imageID, err := strconv.Atoi(param); err == nil {
		image.ID = imageID
	} else {
		image.Slug = param
	}
	return image
}

func generateSSHKeys(param []string) []godo.DropletCreateSSHKey {
	keys := make([]godo.DropletCreateSSHKey, len(param))
	for _, k := range param {
		if id, err := strconv.Atoi(k); err == nil {
			keys = append(keys, godo.DropletCreateSSHKey{ID: id})
		} else {
			keys = append(keys, godo.DropletCreateSSHKey{Fingerprint: k})
		}
	}
	return keys
}

func generateVolumes(param []string) []godo.DropletCreateVolume {
	volumes := make([]godo.DropletCreateVolume, len(param))
	for _, v := range param {
		if v == "" {
			continue
		}
		volumes = append(volumes, godo.DropletCreateVolume{ID: v})
	}
	return volumes
}

// LateInitializeSpec updates any unset (i.e. nil) optional fields of the
// supplied DropletParameters that are set (i.e. non-zero) on the supplied
// Droplet.
func LateInitializeSpec(p *v1alpha1.DropletParameters, observed godo.Droplet) {
	p.Volumes = do.LateInitializeStringSlice(p.Volumes, observed.VolumeIDs)
	p.Tags = do.LateInitializeStringSlice(p.Tags, observed.Tags)
	p.VPCUUID = do.LateInitializeString(p.VPCUUID, observed.VPCUUID)
}
