package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigAgent_Get(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	tests := []struct {
		name          string
		args          []string
		envSetup      map[string]string
		initialConfig ConfigAgent
		wantErr       bool
		want          ConfigAgent
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
				"-g", "localhost:8081",
			},
			wantErr: false,
			want: ConfigAgent{
				PollInterval:   DurationSeconds(2),
				ReportInterval: DurationSeconds(10),
				Key:            "test_flag_key",
				CryptoKey:      "test_flag.pem",
				RateLimit:      5,
				GrpcAddress:    "localhost:8081",
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
				"-g", "localhost:8081",
			},
			envSetup: map[string]string{
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "20",
				"KEY":             "test_env_key",
				"CRYPTO_KEY":      "test_env.pem",
				"RATE_LIMIT":      "15",
				"GRPC_ADDRESS":    "localhost:8082",
			},
			wantErr: false,
			want: ConfigAgent{
				PollInterval:   DurationSeconds(5),
				ReportInterval: DurationSeconds(20),
				Key:            "test_env_key",
				CryptoKey:      "test_env.pem",
				RateLimit:      15,
				GrpcAddress:    "localhost:8082",
			},
		},
		{
			name: "pollInterval равен нулю",
			args: []string{
				"cmd",
				"-p", "0",
				"-r", "10",
				"-k", "test_flag_key",
				"-crypto-key", "test_flag.pem",
				"-l", "5",
			},
			envSetup: map[string]string{
				"POLL_INTERVAL":   "0",
				"REPORT_INTERVAL": "20",
				"KEY":             "test_env_key",
				"CRYPTO_KEY":      "test_env.pem",
				"RATE_LIMIT":      "15",
			},
			wantErr: true,
		},
		{
			name: "pollInterval больше reportInterval",
			args: []string{
				"cmd",
				"-p", "2",
				"-r", "10",
				"-k", "test_flag_key",
				"-crypto-key", "test_flag.pem",
				"-l", "5",
			},
			envSetup: map[string]string{
				"POLL_INTERVAL":   "10",
				"REPORT_INTERVAL": "2",
				"KEY":             "test_env_key",
				"CRYPTO_KEY":      "test_env.pem",
				"RATE_LIMIT":      "15",
			},
			wantErr: true,
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
			assert.Equal(t, tt.want, cfg)
			resetEnv(t)
		})
	}
}

func TestConfigAgent_GetKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	bytesKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	validBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: bytesKey,
	}
	validPEM := pem.EncodeToMemory(validBlock)
	invalidBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte("test"),
	}
	invalidPEM := pem.EncodeToMemory(invalidBlock)
	tests := []struct {
		name      string
		pemData   []byte
		setupPath bool
		expErr    bool
	}{
		{
			name:      "положительный тест",
			pemData:   validPEM,
			setupPath: false,
			expErr:    false,
		},
		{
			name:      "нет файла",
			setupPath: true,
			expErr:    true,
		},
		{
			name:    "текст в файле",
			pemData: []byte("test"),
			expErr:  true,
		},
		{
			name:    "невалидный PEM",
			pemData: invalidPEM,
			expErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "public.pem")

			if !tt.setupPath {
				err := os.WriteFile(path, tt.pemData, 0644)
				require.NoError(t, err)
			} else {
				path = filepath.Join(t.TempDir(), "unknow.pem")
			}
			agent := &ConfigAgent{
				CryptoKey: path,
			}
			publicKey, err := agent.GetKey()
			if tt.expErr {
				assert.Error(t, err)
				assert.Nil(t, publicKey)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &privateKey.PublicKey, publicKey)
			}
		})
	}
}
