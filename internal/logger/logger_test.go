package logger_test

import (
	"testing"

	"github.com/danilov-go/metrics-alerting.git/internal/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "положительный тест",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "Невалидный уровень логирования",
			level:   "unknown",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logger.Initialize(tt.level)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
