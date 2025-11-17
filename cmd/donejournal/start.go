package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/sxwebdev/donejournal/internal/api"
	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/donejournal/internal/mcp"
	"github.com/sxwebdev/donejournal/internal/mcp/provider/groq"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/tmanager"
	"github.com/sxwebdev/donejournal/pkg/migrator"
	"github.com/sxwebdev/donejournal/pkg/sqlite"
	"github.com/sxwebdev/donejournal/sql"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/service"
	"github.com/tkcrm/mx/service/pingpong"
	"github.com/urfave/cli/v3"
)

func startCMD() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "start the server",
		Flags: []cli.Flag{cfgPathsFlag()},
		Action: func(ctx context.Context, cl *cli.Command) error {
			conf := new(config.Config)
			if err := config.Load(conf, envPrefix, cl.StringSlice("config")); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			loggerOpts := append(defaultLoggerOpts(), logger.WithConfig(conf.Log))

			l := logger.NewExtended(loggerOpts...)
			defer func() {
				_ = l.Sync()
			}()

			// init launcher
			ln := launcher.New(
				launcher.WithVersion(version),
				launcher.WithName(appName),
				launcher.WithLogger(l),
				launcher.WithContext(ctx),
				launcher.WithRunnerServicesSequence(launcher.RunnerServicesSequenceLifo),
				launcher.WithOpsConfig(conf.Ops),
				launcher.WithAppStartStopLog(true),
			)

			// check if exists data dir, if not create it
			if _, err := os.Stat(conf.DataDir); os.IsNotExist(err) {
				l.Infof("creating data directory in %s", conf.DataDir)
				if err := os.MkdirAll(conf.DataDir, 0o700); err != nil {
					return fmt.Errorf("failed to create data dir: %w", err)
				}
			}

			sqliteDbPath := filepath.Join(conf.DataDir, "sqlite", "db.sqlite")

			// init sqlite
			sqliteDB, err := sqlite.New(ctx, sqliteDbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize sqlite: %w", err)
			}

			// Print SQLite version if using SQLite storage
			sqliteVersion, err := sqliteDB.GetSQLiteVersion(ctx)
			if err != nil {
				return fmt.Errorf("failed to get SQLite version: %w", err)
			}
			l.Infof("SQLite version: %s", sqliteVersion)

			// check and run all migrations
			m := migrator.New(l, sql.MigrationsFS, sql.MigrationsPath, migrator.DataMigrations{})
			if err := m.MigrateUpAll(ctx, sqliteDbPath); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			sqliteRiverDbPath := filepath.Join(conf.DataDir, "sqlite", "river.sqlite")
			riverSqliteDB, err := sqlite.New(ctx, sqliteRiverDbPath, sqlite.WithName("river-sqlite"))
			if err != nil {
				return fmt.Errorf("failed to initialize sqlite: %w", err)
			}

			migrator, err := rivermigrate.New(riversqlite.New(riverSqliteDB.DB), nil)
			if err != nil {
				return fmt.Errorf("failed to create river sqlite migrator: %w", err)
			}

			_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
			if err != nil {
				return fmt.Errorf("failed to migrate river sqlite database: %w", err)
			}

			st, err := store.New(sqliteDB.DB)
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}

			baseService := baseservices.New(l, st)

			// Initialize MCP provider and service
			provider := groq.NewClient(l, conf.MCP.Groq.APIKey, conf.MCP.Groq.Model)
			mcpService := mcp.New(l, provider)

			// init task manager
			taskManager, err := tmanager.New(riverSqliteDB, baseService, mcpService)
			if err != nil {
				return fmt.Errorf("failed to initialize task manager: %w", err)
			}

			// Initialize API service
			apiService := api.New(l, conf, taskManager)

			// init processor service
			// processorService := processor.New(l, baseService, mcpService)

			// register services
			ln.ServicesRunner().Register(
				service.New(service.WithService(pingpong.New(l))),
				service.New(service.WithService(sqliteDB)),
				service.New(service.WithService(riverSqliteDB)),
				service.New(
					service.WithService(taskManager),
					service.WithShutdownTimeout(time.Minute),
				),
				service.New(service.WithService(apiService)),
				// service.New(
				// 	service.WithService(processorService),
				// 	service.WithShutdownTimeout(time.Minute),
				// ),
				// service.New(service.WithService(appCache)),
				// service.New(service.WithService(botService)),
			)

			return ln.Run()
		},
	}
}
