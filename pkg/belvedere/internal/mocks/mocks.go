package mocks

//go:generate mockgen -package mocks -mock_names Builder=ResourceBuilder -destination resources.go github.com/codahale/belvedere/pkg/belvedere/internal/resources Builder
//go:generate mockgen -package mocks -mock_names Manager=DeploymentsManager -destination deployments.go github.com/codahale/belvedere/pkg/belvedere/internal/deployments Manager
//go:generate mockgen -package mocks -mock_names Service=SetupService -destination setup.go github.com/codahale/belvedere/pkg/belvedere/internal/setup Service
//go:generate mockgen -package mocks -mock_names Service=BackendsService -destination backends.go github.com/codahale/belvedere/pkg/belvedere/internal/backends Service
//go:generate mockgen -package mocks -destination health.go github.com/codahale/belvedere/pkg/belvedere/internal/check HealthChecker
//go:generate mockgen -package mocks -destination project.go github.com/codahale/belvedere/pkg/belvedere Project
//go:generate mockgen -package mocks -destination secrets.go github.com/codahale/belvedere/pkg/belvedere SecretsService
//go:generate mockgen -package mocks -destination logs.go github.com/codahale/belvedere/pkg/belvedere LogService
