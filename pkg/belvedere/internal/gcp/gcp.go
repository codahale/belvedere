package gcp

import (
	"context"
	"sync"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/serviceusage/v1"
)

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

var (
	suService  *serviceusage.Service
	suErr      error
	suOnce     sync.Once
	crmService *cloudresourcemanager.Service
	crmErr     error
	crmOnce    sync.Once
)
