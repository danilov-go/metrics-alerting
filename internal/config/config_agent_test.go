package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigAgent_Get(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	type expData struct {
		pollInterval   DurationSeconds
		reportInterval DurationSeconds
		key            string
		cryptoKey      string
		rateLimit      int
	}

	tests := []struct {
		name          string
		args          []string
		envSetup      map[string]string
		initialConfig ConfigAgent
		wantErr       bool
		want          expData
	}{
		{
			name: "положительный тест(flag)",
			args: []string{
				"cmd",
				"-p", "2",
				"-r", "10",
				"-k", "test_flag_key",
				"-crypto-key", "test_flag.pem",
				"-l", "5",
			},
			wantErr: false,
			want: expData{
				pollInterval:   DurationSeconds(2),
				reportInterval: DurationSeconds(10),
				key:            "test_flag_key",
				cryptoKey:      "test_flag.pem",
				rateLimit:      5,
			},
		},
		{
			name: "положительный тест(env)",
			args: []string{
				"cmd",
				"-p", "2",
				"-r", "10",
				"-k", "test_flag_key",
				"-crypto-key", "test_flag.pem",
				"-l", "5",
			},
			envSetup: map[string]string{
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "20",
				"KEY":             "test_env_key",
				"CRYPTO_KEY":      "test_env.pem",
				"RATE_LIMIT":      "15",
			},
			wantErr: false,
			want: expData{
				pollInterval:   DurationSeconds(5),
				reportInterval: DurationSeconds(20),
				key:            "test_env_key",
				cryptoKey:      "test_env.pem",
				rateLimit:      15,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetEnv(t)
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
			assert.Equal(t, tt.want.pollInterval, cfg.PollInterval)
			assert.Equal(t, tt.want.reportInterval, cfg.ReportInterval)
			assert.Equal(t, tt.want.key, cfg.Key)
			assert.Equal(t, tt.want.cryptoKey, cfg.CryptoKey)
			assert.Equal(t, tt.want.rateLimit, cfg.RateLimit)
			resetEnv(t)
		})
	}
}
