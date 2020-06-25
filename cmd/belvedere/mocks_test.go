package main

//nolint:lll // no way to make these shorter
//go:generate mockgen -package main -destination mock_project_test.go -source ../../pkg/belvedere/belvedere.go Project
//go:generate mockgen -package main -destination mock_secrets_test.go -source ../../pkg/belvedere/secrets.go SecretsService
//go:generate mockgen -package main -destination mock_logs_test.go -source ../../pkg/belvedere/logs.go LogService
//go:generate mockgen -package main -destination mock_output_test.go -source ./internal/cli/output.go Output
//go:generate mockgen -package main -destination mock_apps_test.go -source ../../pkg/belvedere/apps.go AppService
//go:generate mockgen -package main -destination mock_releases_test.go -source ../../pkg/belvedere/releases.go ReleaseService
