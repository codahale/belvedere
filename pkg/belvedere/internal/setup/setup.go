package setup

import (
	"context"
	"fmt"
	"time"

	"go.opencensus.io/trace"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/serviceusage/v1"
)

type Service interface {
	//nolint:lll // no way to shorten this URL
	// SetDMPerms binds the Deployment Service service account to the `owner` role if it has not already
	// been so bound. This allows Deployment Service to add IAM roles to service accounts per
	// https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_deployment_manager_permission_to_set_iam_policies
	SetDMPerms(ctx context.Context, project string, dryRun bool) error

	// EnableAPIs enables all required services for the given GCP project.
	EnableAPIs(ctx context.Context, project string, dryRun bool, interval time.Duration) error

	// ManagedZone returns the managed zone for the given project.
	ManagedZone(ctx context.Context, project string) (*dns.ManagedZone, error)
}

func NewService(ctx context.Context, opts ...option.ClientOption) (Service, error) {
	crm, err := cloudresourcemanager.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	su, err := serviceusage.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	ds, err := dns.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &service{crm: crm, su: su, dns: ds}, nil
}

type service struct {
	crm *cloudresourcemanager.Service
	su  *serviceusage.Service
	dns *dns.Service
}

var _ Service = &service{}

func (s *service) ManagedZone(ctx context.Context, project string) (*dns.ManagedZone, error) {
	ctx, span := trace.StartSpan(ctx, "belvedere.internal.setup.ManagedZone")
	defer span.End()

	// Find the managed zone.
	mz, err := s.dns.ManagedZones.Get(project, "belvedere").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("error getting managed zone: %w", err)
	}

	return mz, nil
}
