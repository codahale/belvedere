package gcp

import (
	"context"
	"sync"

	"google.golang.org/api/cloudresourcemanager/v1"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/logging/v2"
	secretmanager "google.golang.org/api/secretmanager/v1beta1"
	"google.golang.org/api/serviceusage/v1"
)

// Compute creates a new Compute client or returns a previously-created one.
func Compute(ctx context.Context) (*compute.Service, error) {
	gceOnce.Do(func() {
		gceService, gceErr = compute.NewService(ctx)
	})
	return gceService, gceErr
}

// DeploymentManager creates a new Deployment Manager client or returns a previously-created one.
func DeploymentManager(ctx context.Context) (*deploymentmanager.Service, error) {
	dmOnce.Do(func() {
		dmService, dmErr = deploymentmanager.NewService(ctx)
	})
	return dmService, dmErr
}

// ServiceUsage creates a new Service Usage client or returns a previously-created one.
func ServiceUsage(ctx context.Context) (*serviceusage.Service, error) {
	suOnce.Do(func() {
		suService, suErr = serviceusage.NewService(ctx)
	})
	return suService, suErr
}

// CloudResourceManager creates a new Cloud Resource Manager client or returns a previously-created
// one.
func CloudResourceManager(ctx context.Context) (*cloudresourcemanager.Service, error) {
	crmOnce.Do(func() {
		crmService, crmErr = cloudresourcemanager.NewService(ctx)
	})
	return crmService, crmErr
}

// DNS creates a new DNS client or returns a previously-created one.
func DNS(ctx context.Context) (*dns.Service, error) {
	dnsOnce.Do(func() {
		dnsService, dnsErr = dns.NewService(ctx)
	})
	return dnsService, dnsErr
}

// Logging creates a new Logging client or returns a previously-created one.
func Logging(ctx context.Context) (*logging.Service, error) {
	loggingOnce.Do(func() {
		loggingService, loggingErr = logging.NewService(ctx)
	})
	return loggingService, loggingErr
}

// SecretManager creates a new Secret Manager client or returns a previously-created one.
func SecretManager(ctx context.Context) (*secretmanager.Service, error) {
	smOnce.Do(func() {
		smService, smEr = secretmanager.NewService(ctx)
	})
	return smService, smEr
}

var (
	gceService     *compute.Service
	gceErr         error
	gceOnce        sync.Once
	dmService      *deploymentmanager.Service
	dmErr          error
	dmOnce         sync.Once
	suService      *serviceusage.Service
	suErr          error
	suOnce         sync.Once
	crmService     *cloudresourcemanager.Service
	crmErr         error
	crmOnce        sync.Once
	dnsService     *dns.Service
	dnsErr         error
	dnsOnce        sync.Once
	loggingService *logging.Service
	loggingErr     error
	loggingOnce    sync.Once
	smService      *secretmanager.Service
	smEr           error
	smOnce         sync.Once
)
