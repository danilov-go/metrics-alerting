package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"os"

	"dario.cat/mergo"
	"github.com/caarlos0/env/v11"
)

// ConfigServer определяет конфигурацию сервера.
type ConfigServer struct {
	// Net содержит сетевой адрес для запуска сервера.
	Net NetAddress `env:"ADDRESS" json:"address"`
	// StoreInterval определяет интервал времени для сохранения метрик на диск.
	StoreIntrval DurationSeconds `env:"STORE_INTERVAL" json:"store_interval"`
	// FileStoragePath определяет путь к файлу, куда сохраняются метрики.
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	// Restore определяет, нужно ли загружать сохранённые метрики из файла при старте сервера.
	Restore bool `env:"RESTORE" json:"restore"`
	// DatabaseDSN содержит строку подключения к базе данных PostgreSQL.
	DatabaseDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// Key содержит ключ для подписи данных.
	Key string `env:"KEY"`
	// CryptoKey определяет путь до файла с приватным ключом.
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// AuditFile определяет путь к файлу логов аудита.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL содержит URL-адрес внешнего сервиса аудита.
	AuditURL string `env:"AUDIT_URL"`
	//TrustedSubnet определяет строковое представление бесклассовой адресации (CIDR).
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	// RetryDuration определяет продолжительность попыток повтора операций.
	RetryDuration DurationSeconds `env:"RETRY_DURATION"`
	// RetryInterval определяет интервал между повторными попытками выполнения операций.
	RetryInterval DurationSeconds `env:"RETRY_INTERVAL"`
}

// Get парсит конфигурацию сервера.
func (s *ConfigServer) Get() error {
	var cfgJSON ConfigServer
	var cfgEnv ConfigServer
	var cfgFlags ConfigServer
	if path := GetPath(); path != "" {
		if err := LoadJSON(path, &cfgJSON); err != nil {
			return err
		}
	}
	f := flag.NewFlagSet("Run server", flag.ContinueOnError)
	var dummy string
	f.StringVar(&dummy, "c", "", "Path to config file")
	f.StringVar(&dummy, "config", "", "Path to config file")
	f.Var(&cfgFlags.Net, "a", "Net address host:port")
	f.Var(&cfgFlags.StoreIntrval, "i", "StoreIntrval")
	f.StringVar(&cfgFlags.FileStoragePath, "f", "", "FileStoragePath")
	f.StringVar(&cfgFlags.DatabaseDSN, "d", "", "DatabaseDSN")
	f.StringVar(&cfgFlags.Key, "k", "", "Key")
	f.StringVar(&cfgFlags.CryptoKey, "crypto-key", "", "CryptoKey")
	f.StringVar(&cfgFlags.AuditFile, "audit-file", "", "AuditFile")
	f.StringVar(&cfgFlags.AuditURL, "audit-url", "", "AuditURL")
	f.StringVar(&cfgFlags.TrustedSubnet, "t", "", "TrustedSubnet")
	f.BoolVar(&cfgFlags.Restore, "r", false, "Restore")
	f.Var(&cfgFlags.RetryDuration, "retry-duration", "RetryDuration")
	f.Var(&cfgFlags.RetryInterval, "retry-interval", "RetryInterval")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}
	if err := env.Parse(&cfgEnv); err != nil {
		return err
	}
	if err := mergo.Merge(s, cfgJSON, mergo.WithOverride); err != nil {
		return err
	}
	if err := mergo.Merge(s, cfgFlags, mergo.WithOverride); err != nil {
		return err
	}
	if err := mergo.Merge(s, cfgEnv, mergo.WithOverride); err != nil {
		return err
	}
	if _, ok := os.LookupEnv("RESTORE"); ok {
		s.Restore = cfgEnv.Restore
	} else {
		f.Visit(func(fl *flag.Flag) {
			if fl.Name == "r" {
				s.Restore = cfgFlags.Restore
			}
		})
	}
	return nil
}

// GetKey считывает PEM-файл с диска, декодирует его и парсит приватный RSA-ключ.
func (s *ConfigServer) GetKey() (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(s.CryptoKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("*pem.Block равен nil")
	}
	rawKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	privateKey, ok := rawKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("некорректный RSA-ключ")
	}
	return privateKey, nil
}
