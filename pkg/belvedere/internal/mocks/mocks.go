package mocks

//go:generate mockgen -package mocks -mock_names Builder=ResourceBuilder -destination resources.go github.com/codahale/belvedere/pkg/belvedere/internal/resources Builder
//go:generate mockgen -package mocks -mock_names Manager=DeploymentsManager -destination deployments.go github.com/codahale/belvedere/pkg/belvedere/internal/deployments Manager
//go:generate mockgen -package mocks -mock_names Service=SetupService -destination setup.go github.com/codahale/belvedere/pkg/belvedere/internal/setup Service
