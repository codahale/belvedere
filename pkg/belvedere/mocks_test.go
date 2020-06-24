package belvedere

// nolint: lll
//go:generate mockgen -package belvedere -mock_names Builder=ResourceBuilder -destination mock_resources_test.go -source internal/resources/resources.go Builder
//go:generate mockgen -package belvedere -mock_names Manager=DeploymentsManager -destination mock_deployments_test.go -source internal/deployments/deployments.go Manager
//go:generate mockgen -package belvedere -mock_names Service=SetupService -destination mock_setup_test.go -source internal/setup/setup.go Service
//go:generate mockgen -package belvedere -mock_names Service=BackendsService -destination mock_backends_test.go -source internal/backends/backends.go Service
//go:generate mockgen -package belvedere -destination mock_health_test.go -source internal/check/health.go HealthChecker
//go:generate mockgen -package belvedere -destination mock_apps_test.go -source apps.go AppService
