package gcp

import (
	"context"

	"google.golang.org/api/cloudresourcemanager/v1"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/deploymentmanager/v2"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/serviceusage/v1"
)

type gceKey struct{}

func Compute(ctx context.Context) (context.Context, *compute.Service, error) {
	if service, ok := ctx.Value(gceKey{}).(*compute.Service); ok {
		return ctx, service, nil
	}

	service, err := compute.NewService(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, gceKey{}, service), service, nil
}

type dmKey struct{}

func DeploymentManager(ctx context.Context) (context.Context, *deploymentmanager.Service, error) {
	if service, ok := ctx.Value(dmKey{}).(*deploymentmanager.Service); ok {
		return ctx, service, nil
	}

	service, err := deploymentmanager.NewService(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, dmKey{}, service), service, nil
}

type suKey struct{}

func ServiceUsage(ctx context.Context) (context.Context, *serviceusage.Service, error) {
	if service, ok := ctx.Value(suKey{}).(*serviceusage.Service); ok {
		return ctx, service, nil
	}

	service, err := serviceusage.NewService(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, suKey{}, service), service, nil
}

type crmKey struct{}

func CloudResourceManager(ctx context.Context) (context.Context, *cloudresourcemanager.Service, error) {
	if service, ok := ctx.Value(crmKey{}).(*cloudresourcemanager.Service); ok {
		return ctx, service, nil
	}

	service, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, crmKey{}, service), service, nil
}

type dnsKey struct{}

func DNS(ctx context.Context) (context.Context, *dns.Service, error) {
	if service, ok := ctx.Value(dnsKey{}).(*dns.Service); ok {
		return ctx, service, nil
	}

	service, err := dns.NewService(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, dnsKey{}, service), service, nil
}

type logKey struct{}

func Logging(ctx context.Context) (context.Context, *logging.Service, error) {
	if service, ok := ctx.Value(logKey{}).(*logging.Service); ok {
		return ctx, service, nil
	}

	service, err := logging.NewService(ctx)
	if err != nil {
		return nil, nil, err
	}

	return context.WithValue(ctx, logKey{}, service), service, nil
}
