package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	optionSet  string = "Set"
	optionText string = "UnmarshalText"
	optionJSON string = "Unmarshal"
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
