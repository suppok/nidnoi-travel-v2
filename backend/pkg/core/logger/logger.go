package logger

import (
	"go.uber.org/zap"
)

// New returns a console logger locally and a JSON logger on dev/prd.
func New(appEnv string) (*zap.Logger, error) {
	if appEnv == "local" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
