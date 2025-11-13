package main

import (
	"github.com/sxwebdev/donejournal/pkg/migrator"
	"github.com/sxwebdev/donejournal/sql"
	"github.com/tkcrm/mx/logger"
	"github.com/urfave/cli/v3"
)

func migrationsCMD() *cli.Command {
	opts := append(
		defaultLoggerOpts(),
		logger.WithConsoleColored(true),
		logger.WithLogFormat(logger.LoggerFormatConsole),
	)
	l := logger.NewExtended(opts...)
	migrations := migrator.DataMigrations{}
	return migrator.CliCmd(l, sql.MigrationsFS, sql.MigrationsPath, migrations)
}
