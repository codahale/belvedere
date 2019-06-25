package gcp

import (
	"context"
	"sync"

	"google.golang.org/api/cloudresourcemanager/v1"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/serviceusage/v1"
)

var (
	gceService     *compute.Service
	dmService      *deploymentmanager.Service
	suService      *serviceusage.Service
	crmService     *cloudresourcemanager.Service
	dnsService     *dns.Service
	loggingService *logging.Service
	m              sync.Mutex
)

func Compute(ctx context.Context) (*compute.Service, error) {
	m.Lock()
	defer m.Unlock()

	if gceService == nil {
		service, err := compute.NewService(ctx)
		if err != nil {
			return nil, err
		}
		gceService = service
	}

	return gceService, nil
}

func DeploymentManager(ctx context.Context) (*deploymentmanager.Service, error) {
	m.Lock()
	defer m.Unlock()

	if dmService == nil {
		service, err := deploymentmanager.NewService(ctx)
		if err != nil {
			return nil, err
		}
		dmService = service
	}

	return dmService, nil
}

func ServiceUsage(ctx context.Context) (*serviceusage.Service, error) {
	m.Lock()
	defer m.Unlock()

	if suService == nil {
		service, err := serviceusage.NewService(ctx)
		if err != nil {
			return nil, err
		}
		suService = service
	}

	return suService, nil
}

func CloudResourceManager(ctx context.Context) (*cloudresourcemanager.Service, error) {
	m.Lock()
	defer m.Unlock()

	if crmService == nil {
		service, err := cloudresourcemanager.NewService(ctx)
		if err != nil {
			return nil, err
		}
		crmService = service
	}

	return crmService, nil
}

func DNS(ctx context.Context) (*dns.Service, error) {
	m.Lock()
	defer m.Unlock()

	if dnsService == nil {
		service, err := dns.NewService(ctx)
		if err != nil {
			return nil, err
		}
		dnsService = service
	}

	return dnsService, nil
}

func Logging(ctx context.Context) (*logging.Service, error) {
	m.Lock()
	defer m.Unlock()

	if loggingService == nil {
		service, err := logging.NewService(ctx)
		if err != nil {
			return nil, err
		}
		loggingService = service
	}

	return loggingService, nil
}
