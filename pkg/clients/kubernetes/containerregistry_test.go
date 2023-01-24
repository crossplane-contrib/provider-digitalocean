package kubernetes

import (
	"testing"
	"time"

	"github.com/digitalocean/godo"
	"github.com/stretchr/testify/assert"

	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
)

var (
	name                       = "test"
	storageUsageBytes          = 0
	storageUsageBytesUpdatedAt = time.Now()
	createdAt                  = time.Now()
	region                     = "test"
	slug                       = "test"
	includedRepositories       = 0
	includedStorageBytes       = 0
	AllowStorageOverage        = false
	IncludedBandwidthBytes     = 0
	MonthlyPriceInCents        = 0
	eligible                   = false
	eligibilityReasons         = []string{"test"}
	UpdatedAt                  = time.Now()
)

func TestGenerateContainerRegistryObservation(t *testing.T) {
	type args struct {
		registry     *godo.Registry
		subscription *godo.RegistrySubscription
	}
	tests := map[string]struct {
		args args
		want v1alpha1.DOContainerRegistryObservation
	}{
		"AllFilled": {
			args: args{
				registry: &godo.Registry{
					Name:                       name,
					StorageUsageBytes:          uint64(storageUsageBytes),
					StorageUsageBytesUpdatedAt: storageUsageBytesUpdatedAt,
					CreatedAt:                  createdAt,
					Region:                     region,
				},
				subscription: &godo.RegistrySubscription{
					Tier: &godo.RegistrySubscriptionTier{
						Name:                   name,
						Slug:                   slug,
						IncludedRepositories:   uint64(includedRepositories),
						IncludedStorageBytes:   uint64(includedStorageBytes),
						AllowStorageOverage:    AllowStorageOverage,
						IncludedBandwidthBytes: uint64(IncludedBandwidthBytes),
						MonthlyPriceInCents:    uint64(MonthlyPriceInCents),
						Eligible:               eligible,
						EligibilityReasons:     eligibilityReasons,
					},
					CreatedAt: createdAt,
					UpdatedAt: UpdatedAt,
				},
			},
			want: v1alpha1.DOContainerRegistryObservation{
				Name:                       name,
				CreatedAt:                  createdAt.String(),
				Region:                     region,
				StorageUsageBytes:          uint64(storageUsageBytes),
				StorageUsageBytesUpdatedAt: storageUsageBytesUpdatedAt.String(),
				Subscription: v1alpha1.Subscription{
					Tier: v1alpha1.Tier{
						Name:                   name,
						Slug:                   slug,
						IncludedRepositories:   uint64(includedRepositories),
						IncludedStorageBytes:   uint64(includedStorageBytes),
						AllowStorageOverage:    AllowStorageOverage,
						IncludedBandwidthBytes: uint64(includedStorageBytes),
						MonthlyPriceInCents:    uint64(MonthlyPriceInCents),
					},
					CreatedAt: createdAt.String(),
					UpdatedAt: UpdatedAt.String(),
				},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			r := GenerateContainerRegistryObservation(tc.args.registry, tc.args.subscription)
			assert.Equal(t, tc.want, r)
		})
	}
}
