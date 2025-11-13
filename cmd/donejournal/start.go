package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/donejournal/pkg/sqlite"
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
			// m := migrator.New(l, sql.MigrationsFS, sql.MigrationsPath, datamigrations.Migrations)
			// if err := m.MigrateUpAll(ctx, sqliteDbPath); err != nil {
			// 	return fmt.Errorf("failed to run migrations: %w", err)
			// }

			// initialize cache
			// appCache := cache.New()

			// register services
			ln.ServicesRunner().Register(
				service.New(service.WithService(pingpong.New(l))),
				// service.New(service.WithService(appCache)),
				// service.New(service.WithService(botService)),
			)

			return ln.Run()
		},
	}
}
