package kubernetes

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/digitalocean/godo"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-digitalocean/apis/kubernetes/v1alpha1"
	"github.com/crossplane-contrib/provider-digitalocean/pkg/clients/kubernetes"
	"github.com/crossplane-contrib/provider-digitalocean/pkg/clients/kubernetes/fake"
)

var (
	name             = "test"
	region           = "ams1"
	tier             = "stater"
	observedRegistry = &godo.Registry{
		Name:                       name,
		StorageUsageBytes:          0,
		StorageUsageBytesUpdatedAt: time.Now(),
		CreatedAt:                  time.Now(),
		Region:                     region,
	}
	observedSubscription = &godo.RegistrySubscription{
		Tier: &godo.RegistrySubscriptionTier{
			Name:                   tier,
			Slug:                   tier,
			IncludedRepositories:   0,
			IncludedStorageBytes:   0,
			AllowStorageOverage:    false,
			IncludedBandwidthBytes: 0,
			MonthlyPriceInCents:    0,
			Eligible:               false,
			EligibilityReasons:     nil,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
)

func genContainerRegistryObservation(tier string) v1alpha1.DOContainerRegistryObservation {
	return v1alpha1.DOContainerRegistryObservation{
		Name:                       observedRegistry.Name,
		CreatedAt:                  observedRegistry.CreatedAt.String(),
		Region:                     observedRegistry.Region,
		StorageUsageBytes:          observedRegistry.StorageUsageBytes,
		StorageUsageBytesUpdatedAt: observedRegistry.StorageUsageBytesUpdatedAt.String(),
		Subscription: v1alpha1.Subscription{
			Tier: v1alpha1.Tier{
				Name:                   tier,
				Slug:                   tier,
				IncludedRepositories:   observedSubscription.Tier.IncludedRepositories,
				IncludedStorageBytes:   observedSubscription.Tier.IncludedStorageBytes,
				AllowStorageOverage:    observedSubscription.Tier.AllowStorageOverage,
				IncludedBandwidthBytes: observedSubscription.Tier.IncludedBandwidthBytes,
				MonthlyPriceInCents:    observedSubscription.Tier.MonthlyPriceInCents,
			},
			CreatedAt: observedSubscription.CreatedAt.String(),
			UpdatedAt: observedSubscription.UpdatedAt.String(),
		},
	}
}

type args struct {
	containerRegistry kubernetes.RegistryClient
	kube              client.Client
	cr                *v1alpha1.DOContainerRegistry
}

type registryModifier func(*v1alpha1.DOContainerRegistry)

func withExternalName(name string) registryModifier {
	return func(r *v1alpha1.DOContainerRegistry) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) registryModifier {
	return func(r *v1alpha1.DOContainerRegistry) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.DOContainerRegistryParameters) registryModifier {
	return func(r *v1alpha1.DOContainerRegistry) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha1.DOContainerRegistryObservation) registryModifier {
	return func(r *v1alpha1.DOContainerRegistry) { r.Status.AtProvider = s }
}

func registry(m ...registryModifier) *v1alpha1.DOContainerRegistry {
	cr := &v1alpha1.DOContainerRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func Test_containerRegistryExternal_Create(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DOContainerRegistry
		result managed.ExternalCreation
		err    error
	}
	tests := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockCreate: func(context.Context, *godo.RegistryCreateRequest) (*godo.Registry, *godo.Response, error) {
						return &godo.Registry{
							Name: name,
						}, &godo.Response{}, nil
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				})),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"FailedToCreate": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockCreate: func(context.Context, *godo.RegistryCreateRequest) (*godo.Registry, *godo.Response, error) {
						return nil, &godo.Response{}, errors.New("")
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				})),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				}), withConditions(xpv1.Creating(), xpv1.ReconcileError(errors.Wrap(errors.New(""), errContainerRegistryCreateFailed)))),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(""), errContainerRegistryCreateFailed),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := &containerRegistryExternal{kube: tc.kube, client: tc.containerRegistry}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_containerRegistryExternal_Delete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.DOContainerRegistry
		err error
	}
	tests := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockDelete: func(context.Context) (*godo.Response, error) {
						return &godo.Response{
							Response: &http.Response{
								StatusCode: http.StatusOK,
							},
						}, nil
					}},
				cr: registry(),
			},
			want: want{
				cr:  registry(withConditions(xpv1.Deleting())),
				err: nil,
			},
		},
		"DeleteFailed": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockDelete: func(context.Context) (*godo.Response, error) {
						return &godo.Response{
							Response: &http.Response{
								StatusCode: http.StatusBadRequest,
							},
						}, errors.New("")
					}},
				cr: registry(),
			},
			want: want{
				cr:  registry(withConditions(xpv1.Deleting(), xpv1.ReconcileError(errors.Wrap(errors.New(""), errContainerRegistryDeleteFailed)))),
				err: errors.Wrap(errors.New(""), errContainerRegistryDeleteFailed),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := &containerRegistryExternal{kube: tc.kube, client: tc.containerRegistry}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_containerRegistryExternal_Observe(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DOContainerRegistry
		result managed.ExternalObservation
		err    error
	}
	tests := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockGet: func(ctx context.Context) (*godo.Registry, *godo.Response, error) {
						return observedRegistry, nil, nil
					},
					MockGetSubscription: func(ctx context.Context) (*godo.RegistrySubscription, *godo.Response, error) {
						return observedSubscription, nil, nil
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
				}), withExternalName(name), withConditions(xpv1.Creating())),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"ShouldUpdate": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockGet: func(ctx context.Context) (*godo.Registry, *godo.Response, error) {
						return observedRegistry, nil, nil
					},
					MockGetSubscription: func(ctx context.Context) (*godo.RegistrySubscription, *godo.Response, error) {
						return observedSubscription, nil, nil
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: "basic",
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: "basic",
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
					Diff:             subscriptionOutDated,
				},
				err: nil,
			},
		},
		"GetFailed": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockGet: func(ctx context.Context) (*godo.Registry, *godo.Response, error) {
						return nil, &godo.Response{
							Response: &http.Response{
								StatusCode: http.StatusBadRequest,
							},
						}, errors.New("")
					},
					MockGetSubscription: func(ctx context.Context) (*godo.RegistrySubscription, *godo.Response, error) {
						return observedSubscription, nil, nil
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockClient().Update,
				},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: tier,
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Unavailable()), withStatus(genContainerRegistryObservation(tier))),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errors.New(""), errGetContainerRegistry),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := &containerRegistryExternal{kube: tc.kube, client: tc.containerRegistry}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_containerRegistryExternal_Update(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DOContainerRegistry
		result managed.ExternalUpdate
		err    error
	}
	tests := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockUpdateSubscription: func(ctx context.Context, request *godo.RegistrySubscriptionUpdateRequest) (*godo.RegistrySubscription, *godo.Response, error) {
						return &godo.RegistrySubscription{}, nil, nil
					},
				},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: "basic",
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: "basic",
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
				result: managed.ExternalUpdate{},
				err:    nil,
			},
		},
		"UpdateFailed": {
			args: args{
				containerRegistry: &fake.MockRegistryClient{
					MockUpdateSubscription: func(ctx context.Context, request *godo.RegistrySubscriptionUpdateRequest) (*godo.RegistrySubscription, *godo.Response, error) {
						return nil, &godo.Response{
							Response: &http.Response{
								StatusCode: http.StatusBadRequest,
							},
						}, errors.New("")
					}},
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: "basic",
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available()), withStatus(genContainerRegistryObservation(tier))),
			},
			want: want{
				cr: registry(withSpec(v1alpha1.DOContainerRegistryParameters{
					SubscriptionTier: "basic",
					Region:           godo.String(region),
				}), withExternalName(name), withConditions(xpv1.Available(), xpv1.ReconcileError(errors.Wrap(errors.New(""), errContainerRegistryUpdate))), withStatus(genContainerRegistryObservation(tier))),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errors.New(""), errContainerRegistryUpdate),
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			e := &containerRegistryExternal{kube: tc.kube, client: tc.containerRegistry}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
