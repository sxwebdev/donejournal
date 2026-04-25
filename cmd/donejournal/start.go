package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/sxwebdev/donejournal/internal/agent"
	"github.com/sxwebdev/donejournal/internal/agent/provider"
	"github.com/sxwebdev/donejournal/internal/agent/provider/aimlapi"
	"github.com/sxwebdev/donejournal/internal/agent/provider/baseten"
	"github.com/sxwebdev/donejournal/internal/agent/provider/groq"
	"github.com/sxwebdev/donejournal/internal/agent/provider/openrouter"
	"github.com/sxwebdev/donejournal/internal/api"
	"github.com/sxwebdev/donejournal/internal/bot"
	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/donejournal/internal/processor"
	"github.com/sxwebdev/donejournal/internal/services/baseservices"
	"github.com/sxwebdev/donejournal/internal/store"
	"github.com/sxwebdev/donejournal/internal/store/badgerdb"
	"github.com/sxwebdev/donejournal/internal/stt"
	"github.com/sxwebdev/donejournal/internal/tmanager"
	"github.com/sxwebdev/donejournal/pkg/migrator"
	"github.com/sxwebdev/donejournal/pkg/sqlite"
	"github.com/sxwebdev/donejournal/pkg/utils"
	"github.com/sxwebdev/donejournal/sql"
	"github.com/sxwebdev/tokenmanager"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/launcher/services/pingpong"
	"github.com/tkcrm/mx/logger"
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

			// set default timezone
			var err error
			time.Local, err = time.LoadLocation(conf.Timezone)
			if err != nil {
				return fmt.Errorf("failed to set timezone: %w", err)
			}

			carbon.SetTimezone(conf.Timezone)
			carbon.SetWeekStartsAt(carbon.Monday)

			// check if exists data dir, if not create it
			if _, err := os.Stat(conf.DataDir); os.IsNotExist(err) {
				l.Infof("creating data directory in %s", conf.DataDir)
				if err := os.MkdirAll(conf.DataDir, 0o700); err != nil {
					return fmt.Errorf("failed to create data dir: %w", err)
				}
			}

			var authConfig config.AuthConfig

			// get file datadir/auth_secrets.json
			// if not exists create it with generated values
			secretsFilePath := filepath.Join(conf.DataDir, "auth_secrets.json")
			if _, err := os.Stat(secretsFilePath); os.IsNotExist(err) {
				l.Infof("creating secrets file in %s", secretsFilePath)
				if err := os.MkdirAll(filepath.Dir(secretsFilePath), 0o700); err != nil {
					return fmt.Errorf("failed to create secrets dir: %w", err)
				}

				accessToken, err := utils.GenerateRandomString(48, "")
				if err != nil {
					return fmt.Errorf("failed to generate access token secret key: %w", err)
				}

				refreshToken, err := utils.GenerateRandomString(48, "")
				if err != nil {
					return fmt.Errorf("failed to generate refresh token secret key: %w", err)
				}

				authConfig = config.AuthConfig{
					AccessTokenSecretKey:  accessToken,
					RefreshTokenSecretKey: refreshToken,
				}

				data, err := json.MarshalIndent(authConfig, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal secrets: %w", err)
				}

				if err := os.WriteFile(secretsFilePath, data, 0o600); err != nil {
					return fmt.Errorf("failed to create secrets file: %w", err)
				}
			} else {
				data, err := os.ReadFile(secretsFilePath)
				if err != nil {
					return fmt.Errorf("failed to read secrets file: %w", err)
				}

				if err := json.Unmarshal(data, &authConfig); err != nil {
					return fmt.Errorf("failed to unmarshal secrets file: %w", err)
				}

				if authConfig.AccessTokenSecretKey == "" || authConfig.RefreshTokenSecretKey == "" {
					return fmt.Errorf("invalid secrets file: missing keys")
				}
			}

			badgerDbPath := filepath.Join(conf.DataDir, "badger")
			badgerDB, err := badgerdb.New(l, badgerDbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize badgerdb: %w", err)
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

			st, err := store.New(sqliteDB.DB, badgerDB)
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}

			baseService := baseservices.New(l, st)
			// Stop brokers when the context is cancelled (server shutting down)
			go func() {
				<-ctx.Done()
				baseService.Stop()
			}()

			// Initialize token manager with BadgerDB as token store
			tokenMgr := tokenmanager.New[api.TokenData](badgerDB, authConfig.AccessTokenSecretKey, 30*24*time.Hour)

			// Initialize agent LLM provider based on config (groq has priority if both enabled).
			llmProvider, err := selectLLMProvider(l, conf.Agent)
			if err != nil {
				return fmt.Errorf("failed to init agent provider: %w", err)
			}
			agentService := agent.New(l, llmProvider, baseService, badgerDB)

			// init bot service
			botService, err := bot.New(l, conf.Telegram.BotToken)
			if err != nil {
				return fmt.Errorf("failed to initialize bot service: %w", err)
			}

			// init processor service
			processorService := processor.New(l, baseService, agentService, botService)

			// init STT service (optional, only if enabled)
			var sttService *stt.Service
			if conf.STT.Enabled {
				sttService, err = stt.New(ctx, l, conf.DataDir, conf.STT.ModelPath)
				if err != nil {
					return fmt.Errorf("failed to initialize STT service: %w", err)
				}
			}

			// init task manager
			taskManager, err := tmanager.New(l, riverSqliteDB, processorService, botService, sttService, conf.STT.MaxDuration)
			if err != nil {
				return fmt.Errorf("failed to initialize task manager: %w", err)
			}

			// Initialize API service with Connect-RPC
			apiService := api.New(l, conf, baseService, st, taskManager, tokenMgr)

			// register services
			ln.ServicesRunner().Register(
				launcher.NewService(launcher.WithService(pingpong.New(l))),
				launcher.NewService(launcher.WithService(badgerDB)),
				launcher.NewService(launcher.WithService(sqliteDB)),
				launcher.NewService(launcher.WithService(riverSqliteDB)),
				launcher.NewService(
					launcher.WithService(taskManager),
					launcher.WithShutdownTimeout(time.Minute),
				),
				launcher.NewService(launcher.WithService(botService)),
				launcher.NewService(launcher.WithService(apiService)),
			)

			return ln.Run()
		},
	}
}

// selectLLMProvider picks the agent LLM provider according to the config.
// Groq has priority if both providers are enabled. Returns an error if none
// is enabled or the enabled provider is misconfigured.
func selectLLMProvider(log logger.Logger, cfg config.AgentConfig) (provider.Provider, error) {
	switch {
	case cfg.Groq.Enabled:
		if cfg.Groq.APIKey == "" {
			return nil, fmt.Errorf("groq is enabled but api_key is empty")
		}
		log.Infof("using LLM provider: groq (model=%s)", cfg.Groq.Model)
		return groq.NewClient(log, cfg.Groq.APIKey, cfg.Groq.Model), nil
	case cfg.OpenRouter.Enabled:
		if cfg.OpenRouter.APIKey == "" {
			return nil, fmt.Errorf("openrouter is enabled but api_key is empty")
		}
		log.Infof("using LLM provider: openrouter (model=%s)", cfg.OpenRouter.Model)
		return openrouter.NewClient(log, cfg.OpenRouter.APIKey, cfg.OpenRouter.Model), nil
	case cfg.Baseten.Enabled:
		if cfg.Baseten.APIKey == "" {
			return nil, fmt.Errorf("baseten is enabled but api_key is empty")
		}
		if cfg.Baseten.Model == "" {
			return nil, fmt.Errorf("baseten is enabled but model is empty")
		}
		log.Infof("using LLM provider: baseten (model=%s)", cfg.Baseten.Model)
		return baseten.NewProvider(log, cfg.Baseten.APIKey, cfg.Baseten.Model), nil
	case cfg.AIMLAPI.Enabled:
		if cfg.AIMLAPI.APIKey == "" {
			return nil, fmt.Errorf("aimlapi is enabled but api_key is empty")
		}
		if cfg.AIMLAPI.Model == "" {
			return nil, fmt.Errorf("aimlapi is enabled but model is empty")
		}
		log.Infof("using LLM provider: aimlapi (model=%s)", cfg.AIMLAPI.Model)
		return aimlapi.NewProvider(log, cfg.AIMLAPI.APIKey, cfg.AIMLAPI.Model), nil
	default:
		return nil, fmt.Errorf("no LLM provider enabled: set agent.groq.enabled, agent.openrouter.enabled, agent.baseten.enabled or agent.aimlapi.enabled to true")
	}
}
