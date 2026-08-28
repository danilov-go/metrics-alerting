// Package logger реализует глобальный логер для приложения.
package logger

import (
	"go.uber.org/zap"
)

// Log является глобальным экземпляром логера.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует глобальный логер с заданным уровнем.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
