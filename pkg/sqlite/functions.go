package sqlite

import (
	"database/sql/driver"
	"strings"

	"modernc.org/sqlite"
)

func init() {
	sqlite.MustRegisterDeterministicScalarFunction("unicode_lower", 1,
		func(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
			s, ok := args[0].(string)
			if !ok {
				return args[0], nil
			}
			return strings.ToLower(s), nil
		},
	)
}
