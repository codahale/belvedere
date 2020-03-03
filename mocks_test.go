package main

//go:generate mockgen -package main -destination mock_project_test.go -source pkg/belvedere/belvedere.go Project
//go:generate mockgen -package main -destination mock_secrets_test.go -source pkg/belvedere/secrets.go SecretsService
//go:generate mockgen -package main -destination mock_logs_test.go -source pkg/belvedere/logs.go LogService
//go:generate mockgen -package main -destination mock_tables_test.go -source tables.go TableWriter
//go:generate mockgen -package main -destination mock_apps_test.go -source pkg/belvedere/apps.go AppService
//go:generate mockgen -package main -destination mock_releases_test.go -source pkg/belvedere/releases.go ReleaseService
