package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	optionSet    string = "Set"
	optionText   string = "UnmarshalText"
	optionJSON   string = "Unmarshal"
	optionAgent  string = "Agent"
	optionServer string = "Server"
)

func TestNetAddress(t *testing.T) {
	type want struct {
		err  bool
		host string
		port int
		str  string
	}
	tests := []struct {
		name   string
		option string
		input  string
		exp    want
	}{
		{
			name:   "Set: положительный тест",
			option: optionSet,
			input:  "localhost:8080",
			exp: want{
				err:  false,
				host: "localhost",
				port: 8080,
				str:  "localhost:8080",
			},
		},
		{
			name:   "Set: нет порта",
			option: optionSet,
			input:  "localhost",
			exp: want{
				err: true,
			},
		},
		{
			name:   "Set: порт не число",
			option: optionSet,
			input:  "localhost:invalid",
			exp: want{
				err: true,
			},
		},

		{
			name:   "UnmarshalText: положительный тест",
			option: optionText,
			input:  "10.1.0.1:8080",
			exp: want{
				err:  false,
				host: "10.1.0.1",
				port: 8080,
				str:  "10.1.0.1:8080",
			},
		},
		{
			name:   "UnmarshalJSON: невалидный формат адреса внутри JSON",
			option: optionJSON,
			input:  `"invalid-address"`,
			exp: want{
				err: true,
			},
		},
		{
			name:   "UnmarshalJSON: тип не строка",
			option: optionJSON,
			input:  `12345`,
			exp: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var test NetAddress
			var err error
			switch tt.option {
			case optionSet:
				err = test.Set(tt.input)
			case optionText:
				err = test.UnmarshalText([]byte(tt.input))
			case optionJSON:
				err = json.Unmarshal([]byte(tt.input), &test)
			}
			if tt.exp.err {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.exp.host, test.Host)
			assert.Equal(t, tt.exp.port, test.Port)
			if tt.exp.str != "" {
				assert.Equal(t, tt.exp.str, test.String())
			}
		})
	}
}

func TestDurationSeconds(t *testing.T) {
	type want struct {
		err bool
		s   DurationSeconds
		str string
	}
	tests := []struct {
		name   string
		option string
		input  string
		exp    want
	}{
		{
			name:   "Set: положительный тест (число в строке)",
			option: optionSet,
			input:  "15",
			exp: want{
				err: false,
				s:   15,
				str: "15",
			},
		},
		{
			name:   "Set: положительный тест (Суффикс s)",
			option: optionSet,
			input:  "15s",
			exp: want{
				err: false,
				s:   15,
				str: "15",
			},
		},
		{
			name:   "Set: положительный тест (Суффикс m)",
			option: optionSet,
			input:  "2m",
			exp: want{
				err: false,
				s:   120,
				str: "120",
			},
		},
		{
			name:   "Set: положительный тест (Суффикс h)",
			option: optionSet,
			input:  "1h",
			exp: want{
				err: false,
				s:   3600,
				str: "3600",
			},
		},
		{
			name:   "Set: невалидный формат строки времени",
			option: optionSet,
			input:  "invalid",
			exp: want{
				err: true,
			},
		},
		{
			name:   "UnmarshalText: положительный тест",
			option: optionText,
			input:  "15s",
			exp: want{
				err: false,
				s:   15,
				str: "15",
			},
		},
		{
			name:   "UnmarshalJSON: положительный тест (число)",
			option: optionJSON,
			input:  `15`,
			exp: want{
				err: false,
				s:   15,
				str: "15",
			},
		},
		{
			name:   "UnmarshalJSON: положительный тест (с суффиксом)",
			option: optionJSON,
			input:  `"5m"`,
			exp: want{
				err: false,
				s:   300,
				str: "300",
			},
		},
		{
			name:   "UnmarshalJSON: неподдерживаемый тип",
			option: optionJSON,
			input:  `"invalid"`,
			exp: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d DurationSeconds
			var err error
			switch tt.option {
			case optionSet:
				err = d.Set(tt.input)
			case optionText:
				err = d.UnmarshalText([]byte(tt.input))
			case optionJSON:
				err = json.Unmarshal([]byte(tt.input), &d)
			default:
				t.Fatalf("неизвестная опция: %s", tt.option)
			}
			if tt.exp.err {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.exp.s, d)
			if tt.exp.str != "" {
				assert.Equal(t, tt.exp.str, d.String())
			}
		})
	}
}

func TestLoadJSON(t *testing.T) {
	tests := []struct {
		name        string
		option      string
		jsonConfig  string
		expError    bool
		invalidPath bool
		expConfig   any
	}{
		{
			name:   "положительный тест (сервер)",
			option: optionServer,
			jsonConfig: `
			{
				"address": "localhost:8080",
				"restore": true, 
				"store_interval": "1s",
				"store_file": "/path/to/file.db",
				"database_dsn": "test", 
				"crypto_key": "/path/to/key.pem", 
				"trusted_subnet": "192.168.1.0/24",
				"grpc_address": "localhost:8081"
			}`,
			expError: false,
			expConfig: &ConfigServer{
				Net: NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				Restore:         true,
				StoreIntrval:    1,
				FileStoragePath: "/path/to/file.db",
				DatabaseDSN:     "test",
				CryptoKey:       "/path/to/key.pem",
				TrustedSubnet:   "192.168.1.0/24",
				GrpcAddress:     "localhost:8081",
			},
		},
		{
			name:   "положительный тест (агент)",
			option: optionAgent,
			jsonConfig: `
			{
				"address": "localhost:8080", 
				"report_interval": "1s",
				"poll_interval": "1s", 
				"crypto_key": "/path/to/key.pem",
				"grpc_address": "localhost:8081"
			}`,
			expError: false,
			expConfig: &ConfigAgent{
				Net: NetAddress{
					Host: "localhost",
					Port: 8080,
				},
				ReportInterval: 1,
				PollInterval:   1,
				CryptoKey:      "/path/to/key.pem",
				GrpcAddress:    "localhost:8081",
			},
		},
		{
			name:        "невалидный путь",
			option:      optionServer,
			jsonConfig:  `{}`,
			invalidPath: true,
			expError:    true,
			expConfig:   nil,
		},
		{
			name:   "ошибка JSON",
			option: optionServer,
			jsonConfig: `
			{
				"address": "localhost:8080",
				"restore": "123"
			}`,
			invalidPath: false,
			expError:    true,
			expConfig:   nil,
		},
		{
			name:   "ошибка JSON",
			option: optionServer,
			jsonConfig: `
			{
				"address": "localhost:8080",
				"store_interval": []
			}`,
			invalidPath: false,
			expError:    true,
			expConfig:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configPath string
			if tt.invalidPath {
				configPath = filepath.Join(t.TempDir(), "unknow.json")
			} else {
				testDir := t.TempDir()
				configPath = filepath.Join(testDir, "config.json")
				err := os.WriteFile(configPath, []byte(tt.jsonConfig), 0644)
				require.NoError(t, err)
			}
			var expErr error
			var config any
			switch tt.option {
			case optionAgent:
				cfg := ConfigAgent{}
				expErr = LoadJSON(configPath, &cfg)
				config = &cfg
			case optionServer:
				cfg := ConfigServer{}
				expErr = LoadJSON(configPath, &cfg)
				config = &cfg
			default:
				t.Fatalf("неизвестная опция: %s", tt.option)
			}
			if tt.expError {
				assert.Error(t, expErr)
			} else {
				assert.NoError(t, expErr)
				assert.Equal(t, tt.expConfig, config)
			}
		})
	}
}

func TestPrintBuild(t *testing.T) {
	tmpl := "Build version: %s\nBuild date: %s\nBuild commit: %s\n"
	tests := []struct {
		name    string
		version string
		date    string
		commit  string
	}{
		{
			name:    "передача параметров сборки",
			version: "v1.2.3",
			date:    time.Now().Format(time.DateOnly),
			commit:  "testcomit",
		},
		{
			name:    "без передачи параметров",
			version: "N/A",
			date:    "N/A",
			commit:  "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			expData := fmt.Sprintf(tmpl, tt.version, tt.date, tt.commit)
			PrintBuild(&buf, tt.version, tt.date, tt.commit)
			assert.Equal(t, expData, buf.String())
		})
	}
}

func TestGetPath(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	tests := []struct {
		name     string
		args     []string
		envSetup map[string]string
		expected string
	}{
		{
			name: "положительный тест (env)",
			args: []string{"cmd", "-c", "flag_path.json"},
			envSetup: map[string]string{
				"CONFIG": "env_path.json",
			},
			expected: "env_path.json",
		},
		{
			name:     "положительный тест (flag)",
			args:     []string{"cmd", "-config", "flag_path.json"},
			expected: "flag_path.json",
		},
		{
			name:     "флаг без пути",
			args:     []string{"cmd", "-c"},
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetEnv(t)
			for k, v := range tt.envSetup {
				t.Setenv(k, v)
			}
			os.Args = tt.args
			assert.Equal(t, tt.expected, GetPath())
			resetEnv(t)
		})
	}
}

func TestGetHost(t *testing.T) {
	packetConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer packetConn.Close()
	validAddr := packetConn.LocalAddr().String()
	tests := []struct {
		name   string
		adr    string
		expErr bool
	}{
		{
			name:   "положительный тест",
			adr:    validAddr,
			expErr: false,
		},

		{
			name:   "некорректный формат порта",
			adr:    "localhost:invalid",
			expErr: true,
		},
		{
			name:   "пустой адрес",
			adr:    "",
			expErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := GetHost(tt.adr)
			if tt.expErr {
				assert.Error(t, err)
				assert.Empty(t, host)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, host)
			}
		})
	}
}
