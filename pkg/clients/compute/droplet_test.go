package compute

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-digitalocean/apis/compute/v1alpha1"

	"github.com/digitalocean/godo"
)

var (
	name              = "mock-droplet"
	region            = "mock-region"
	size              = "mock-v1cpu-1gb"
	image             = "mock-ubuntu-20-04-x64"
	sshKeys           = []string{"mock-pub-key"}
	backups           = false
	IPv6              = false
	privateNetworking = false
	monitoring        = false
	userData          = "mock-user-data"
	volumes           = []string{"mock-volume"}
	tags              = []string{"mock-tag"}
	VPCUUID           = "mock-vpcuuid"
	withDropletAgent  = false
)

func TestGenerateDroplet(t *testing.T) {
	type args struct {
		name   string
		params v1alpha1.DropletParameters
		create godo.DropletCreateRequest
	}

	tests := map[string]struct {
		args args
		want *godo.DropletCreateRequest
	}{
		"AllFilled": {
			args: args{
				name: name,
				params: v1alpha1.DropletParameters{
					Region:            region,
					Size:              size,
					Image:             image,
					SSHKeys:           sshKeys,
					Backups:           &backups,
					IPv6:              &IPv6,
					PrivateNetworking: &privateNetworking,
					Monitoring:        &monitoring,
					UserData:          &userData,
					Volumes:           volumes,
					Tags:              tags,
					VPCUUID:           &VPCUUID,
					WithDropletAgent:  &withDropletAgent,
				},
				create: godo.DropletCreateRequest{},
			},
			want: &godo.DropletCreateRequest{
				Name:              name,
				Region:            region,
				Size:              size,
				Image:             generateImage(image),
				SSHKeys:           generateSSHKeys(sshKeys),
				Backups:           backups,
				IPv6:              IPv6,
				PrivateNetworking: privateNetworking,
				Monitoring:        monitoring,
				UserData:          userData,
				Volumes:           generateVolumes(volumes),
				Tags:              tags,
				VPCUUID:           VPCUUID,
				WithDropletAgent:  &withDropletAgent,
			},
		},
	}

	for tName, tc := range tests {
		t.Run(tName, func(t *testing.T) {
			GenerateDroplet(name, tc.args.params, &tc.args.create)

			if diff := cmp.Diff(tc.want, &tc.args.create); diff != "" {
				t.Errorf("GenerateDroplet(...): -want, +got:\n%s", diff)
			}
		})
	}
}
