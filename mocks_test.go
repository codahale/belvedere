package main

//go:generate mockgen -package main -destination mock_project_test.go github.com/codahale/belvedere/pkg/belvedere Project
//go:generate mockgen -package main -destination mock_secrets_test.go github.com/codahale/belvedere/pkg/belvedere SecretsService
//go:generate mockgen -package main -destination mock_logs_test.go github.com/codahale/belvedere/pkg/belvedere LogService
//go:generate mockgen -package main -destination mock_tables_test.go -source tables.go TableWriter
//go:generate mockgen -package main -destination mock_apps_test.go github.com/codahale/belvedere/pkg/belvedere AppService
