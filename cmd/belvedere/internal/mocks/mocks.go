package mocks

//go:generate mockgen -package mocks -destination mock_project.go -source ../../../../pkg/belvedere/belvedere.go Project
//go:generate mockgen -package mocks -destination mock_secrets.go -source ../../../../pkg/belvedere/secrets.go SecretsService
//go:generate mockgen -package mocks -destination mock_logs.go -source ../../../../pkg/belvedere/logs.go LogService
//go:generate mockgen -package mocks -destination mock_output.go -source ../cli/output.go Output
//go:generate mockgen -package mocks -destination mock_apps.go -source ../../../../pkg/belvedere/apps.go AppService
//go:generate mockgen -package mocks -destination mock_releases.go -source ../../../../pkg/belvedere/releases.go ReleaseService
