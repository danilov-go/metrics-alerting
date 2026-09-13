package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigServer_Get(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	type want struct {
		storeInterval   DurationSeconds
		fileStoragePath string
		databaseDSN     string
		key             string
		auditFile       string
		auditURL        string
		restore         bool
	}
	tests := []struct {
		name          string
		args          []string
		envSetup      map[string]string
		initialConfig ConfigServer
		wantErr       bool
		exp           want
	}{
		{
			name: "положительный тест(flag)",
			args: []string{
				"cmd",
				"-i", "250",
				"-f", "test_flag.txt",
				"-d", "test_flag_db",
				"-k", "test_flag_key",
				"-audit-file", "test_flag.log",
				"-audit-url", "test_flag",
			},
			wantErr: false,
			exp: want{
				storeInterval:   DurationSeconds(250),
				fileStoragePath: "test_flag.txt",
				databaseDSN:     "test_flag_db",
				key:             "test_flag_key",
				auditFile:       "test_flag.log",
				auditURL:        "test_flag",
				restore:         false,
			},
		},
		{
			name: "положительный тест(env)",
			args: []string{
				"cmd",
				"-i", "250",
				"-f", "test_flag.txt",
				"-d", "test_flag_db",
				"-k", "test_flag_key",
				"-audit-file", "test_flag.log",
				"-audit-url", "test_flag",
			},
			envSetup: map[string]string{
				"STORE_INTERVAL":    "20",
				"FILE_STORAGE_PATH": "test_env.txt",
				"DATABASE_DSN":      "test_env_db",
				"KEY":               "test_env_key",
				"AUDIT_FILE":        "test_env.log",
				"AUDIT_URL":         "test_env",
			},
			wantErr: false,
			exp: want{
				storeInterval:   DurationSeconds(20),
				fileStoragePath: "test_env.txt",
				databaseDSN:     "test_env_db",
				key:             "test_env_key",
				auditFile:       "test_env.log",
				auditURL:        "test_env",
				restore:         false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ADDRESS", "")
			t.Setenv("STORE_INTERVAL", "")
			t.Setenv("FILE_STORAGE_PATH", "")
			t.Setenv("RESTORE", "")
			t.Setenv("DATABASE_DSN", "")
			t.Setenv("KEY", "")
			t.Setenv("CRYPTO_KEY", "")
			t.Setenv("AUDIT_FILE", "")
			t.Setenv("AUDIT_URL", "")
			t.Setenv("RETRY_DURATION", "")
			t.Setenv("RETRY_INTERVAL", "")
			for k, v := range tt.envSetup {
				t.Setenv(k, v)
			}
			os.Args = tt.args
			cfg := tt.initialConfig
			err := cfg.Get()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.exp.storeInterval, cfg.StoreIntrval)
			assert.Equal(t, tt.exp.fileStoragePath, cfg.FileStoragePath)
			assert.Equal(t, tt.exp.databaseDSN, cfg.DatabaseDSN)
			assert.Equal(t, tt.exp.key, cfg.Key)
			assert.Equal(t, tt.exp.auditFile, cfg.AuditFile)
			assert.Equal(t, tt.exp.auditURL, cfg.AuditURL)
			assert.Equal(t, tt.exp.restore, cfg.Restore)
		})
	}
}
