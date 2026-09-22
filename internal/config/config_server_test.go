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

func TestConfigServer_Get(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name          string
		args          []string
		envSetup      map[string]string
		initialConfig ConfigServer
		wantErr       bool
		want          ConfigServer
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
				"-t", "192.168.1.0/24",
				"-g", "localhost:8081",
				"-r", "true",
			},
			wantErr: false,
			want: ConfigServer{
				StoreIntrval:    DurationSeconds(250),
				FileStoragePath: "test_flag.txt",
				DatabaseDSN:     "test_flag_db",
				Key:             "test_flag_key",
				AuditFile:       "test_flag.log",
				AuditURL:        "test_flag",
				TrustedSubnet:   "192.168.1.0/24",
				Restore:         true,
				GrpcAddress:     "localhost:8081",
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
				"-t", "192.168.1.0/24",
				"-g", "localhost:8081",
				"-r", "true",
			},
			envSetup: map[string]string{
				"STORE_INTERVAL":    "20",
				"FILE_STORAGE_PATH": "test_env.txt",
				"DATABASE_DSN":      "test_env_db",
				"KEY":               "test_env_key",
				"AUDIT_FILE":        "test_env.log",
				"AUDIT_URL":         "test_env",
				"TRUSTED_SUBNET":    "192.168.1.0/25",
				"RESTORE":           "false",
				"GRPC_ADDRESS":      "localhost:8082",
			},
			wantErr: false,
			want: ConfigServer{
				StoreIntrval:    DurationSeconds(20),
				FileStoragePath: "test_env.txt",
				DatabaseDSN:     "test_env_db",
				Key:             "test_env_key",
				AuditFile:       "test_env.log",
				AuditURL:        "test_env",
				TrustedSubnet:   "192.168.1.0/25",
				Restore:         false,
				GrpcAddress:     "localhost:8082",
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
			var cfg ConfigServer
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

func resetEnv(t *testing.T) {
	envs := []string{
		"ADDRESS",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
		"STORE_INTERVAL",
		"FILE_STORAGE_PATH",
		"RESTORE",
		"DATABASE_DSN",
		"AUDIT_FILE",
		"AUDIT_URL",
		"RETRY_DURATION",
		"RETRY_INTERVAL",
		"TRUSTED_SUBNET",
		"POLL_INTERVAL",
		"REPORT_INTERVAL",
		"RATE_LIMIT",
		"GRPC_ADDRESS",
	}
	for _, env := range envs {
		if err := os.Unsetenv(env); err != nil {
			assert.NoError(t, err)
		}
	}
}

func TestConfigServer_GetKey(t *testing.T) {
	expKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	bytesKey, err := x509.MarshalPKCS8PrivateKey(expKey)
	require.NoError(t, err)
	validBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bytesKey,
	}
	validPEM := pem.EncodeToMemory(validBlock)
	invalidBlock := &pem.Block{
		Type:  "PRIVATE KEY",
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
			path := filepath.Join(t.TempDir(), "private.pem")
			if !tt.setupPath {
				err := os.WriteFile(path, tt.pemData, 0644)
				require.NoError(t, err)
			} else {
				path = filepath.Join(t.TempDir(), "unknown.pem")
			}
			server := &ConfigServer{
				CryptoKey: path,
			}
			privateKey, err := server.GetKey()
			if tt.expErr {
				assert.Error(t, err)
				assert.Nil(t, privateKey)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expKey, privateKey)
			}
		})
	}
}

func TestParseTrustedSubnet(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		expNet        string
		wantErr       bool
	}{
		{
			name:          "положительный тест",
			trustedSubnet: "192.168.1.0/24",
			expNet:        "192.168.1.0/24",
			wantErr:       false,
		},
		{
			name:          "пустая строка",
			trustedSubnet: "",
			expNet:        "",
			wantErr:       false,
		},
		{
			name:          "передан IP без маски",
			trustedSubnet: "192.168.1.0",
			wantErr:       true,
		},
		{
			name:          "некорректный формат маски",
			trustedSubnet: "192.168.1.0/72",
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipNet, err := ParseTrustedSubnet(tt.trustedSubnet)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ipNet)
			} else {
				assert.NoError(t, err)
				if tt.trustedSubnet == "" {
					assert.Nil(t, ipNet)
				} else {
					assert.Equal(t, tt.expNet, ipNet.String())
				}
			}
		})
	}
}
