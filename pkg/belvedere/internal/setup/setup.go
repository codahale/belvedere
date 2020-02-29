package setup

import (
	"context"
	"time"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/serviceusage/v1"
)

type Manager interface {
	// SetDMPerms binds the Deployment Manager service account to the `owner` role if it has not already
	// been so bound. This allows Deployment Manager to add IAM roles to service accounts per
	// https://cloud.google.com/deployment-manager/docs/configuration/set-access-control-resources#granting_deployment_manager_permission_to_set_iam_policies
	SetDMPerms(ctx context.Context, project string, dryRun bool) error

	// EnableAPIs enables all required services for the given GCP project.
	EnableAPIs(ctx context.Context, project string, dryRun bool, interval time.Duration) error
}

func NewManager(ctx context.Context) (Manager, error) {
	crm, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	su, err := serviceusage.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &manager{crm: crm, su: su}, nil
}

type manager struct {
	crm *cloudresourcemanager.Service
	su  *serviceusage.Service
}

var _ Manager = &manager{}
