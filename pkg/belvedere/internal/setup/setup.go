package setup

import (
	"context"
	"time"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/serviceusage/v1"
)

type Service interface {
	// SetDMPerms binds the Deployment Service service account to the `owner` role if it has not already
	// been so bound. This allows Deployment Service to add IAM roles to service accounts per
	// https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_deployment_manager_permission_to_set_iam_policies
	SetDMPerms(ctx context.Context, project string, dryRun bool) error

	// EnableAPIs enables all required services for the given GCP project.
	EnableAPIs(ctx context.Context, project string, dryRun bool, interval time.Duration) error
}

func NewService(ctx context.Context) (Service, error) {
	crm, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	su, err := serviceusage.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &service{crm: crm, su: su}, nil
}

type service struct {
	crm *cloudresourcemanager.Service
	su  *serviceusage.Service
}

var _ Service = &service{}
