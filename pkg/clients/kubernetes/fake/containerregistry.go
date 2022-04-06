package fake

import (
	"context"

	"github.com/digitalocean/godo"

	"github.com/crossplane-contrib/provider-digitalocean/pkg/clients/kubernetes"
)

// this ensures that the mock implements the client interface
var _ kubernetes.RegistryClient = (*MockRegistryClient)(nil)

// MockRegistryClient is a type that implements all the methods for RegistryClient interface
type MockRegistryClient struct {
	MockGet                func(context.Context) (*godo.Registry, *godo.Response, error)
	MockGetSubscription    func(context.Context) (*godo.RegistrySubscription, *godo.Response, error)
	MockCreate             func(context.Context, *godo.RegistryCreateRequest) (*godo.Registry, *godo.Response, error)
	MockUpdateSubscription func(context.Context, *godo.RegistrySubscriptionUpdateRequest) (*godo.RegistrySubscription, *godo.Response, error)
	MockDelete             func(context.Context) (*godo.Response, error)
}

// Get mocks Get method
func (c *MockRegistryClient) Get(ctx context.Context) (*godo.Registry, *godo.Response, error) {
	return c.MockGet(ctx)
}

// GetSubscription mocks GetSubscription method
func (c *MockRegistryClient) GetSubscription(ctx context.Context) (*godo.RegistrySubscription, *godo.Response, error) {
	return c.MockGetSubscription(ctx)
}

// Create mocks Create method
func (c *MockRegistryClient) Create(ctx context.Context, request *godo.RegistryCreateRequest) (*godo.Registry, *godo.Response, error) {
	return c.MockCreate(ctx, request)
}

// UpdateSubscription mocks UpdateSubscription method
func (c *MockRegistryClient) UpdateSubscription(ctx context.Context, request *godo.RegistrySubscriptionUpdateRequest) (*godo.RegistrySubscription, *godo.Response, error) {
	return c.MockUpdateSubscription(ctx, request)
}

// Delete mocks Delete method
func (c *MockRegistryClient) Delete(ctx context.Context) (*godo.Response, error) {
	return c.MockDelete(ctx)
}
