package belvedere

//go:generate mockgen -package belvedere -mock_names Builder=ResourceBuilder -destination mock_resources_test.go github.com/codahale/belvedere/pkg/belvedere/internal/resources Builder
//go:generate mockgen -package belvedere -mock_names Manager=DeploymentsManager -destination mock_deployments_test.go github.com/codahale/belvedere/pkg/belvedere/internal/deployments Manager
//go:generate mockgen -package belvedere -mock_names Service=SetupService -destination mock_setup_test.go github.com/codahale/belvedere/pkg/belvedere/internal/setup Service
//go:generate mockgen -package belvedere -mock_names Service=BackendsService -destination mock_backends_test.go github.com/codahale/belvedere/pkg/belvedere/internal/backends Service
//go:generate mockgen -package belvedere -destination mock_health_test.go github.com/codahale/belvedere/pkg/belvedere/internal/check HealthChecker
