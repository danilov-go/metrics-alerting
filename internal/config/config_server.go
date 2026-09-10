package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"os"

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
	// RetryDuration определяет продолжительность попыток повтора операций.
	RetryDuration DurationSeconds `env:"RETRY_DURATION"`
	// RetryInterval определяет интервал между повторными попытками выполнения операций.
	RetryInterval DurationSeconds `env:"RETRY_INTERVAL"`
	// ValidDB флаг, указывающий на корректность строки подключения к базе данных.
	ValidDB bool
	// ValidFile флаг, указывающий на корректность пути к файлу хранилища метрик.
	ValidFile bool
	// ValidFileAudit флаг, указывающий на корректность и доступность файла аудита.
	ValidFileAudit bool
	// ValidURLAudit флаг, указывающий на корректность и доступность URL аудита.
	ValidURLAudit bool
}

// Get парсит конфигурацию сервера.
func (s *ConfigServer) Get() error {
	if path := GetPath(); path != "" {
		if err := LoadJSON(path, s); err != nil {
			return err
		}
	}
	f := flag.NewFlagSet("Run server", flag.ContinueOnError)
	var dummy string
	f.StringVar(&dummy, "c", "", "Path to config file")
	f.StringVar(&dummy, "config", "", "Path to config file")
	f.Var(&s.Net, "a", "Net address host:port")
	f.Var(&s.StoreIntrval, "i", "StoreIntrval")
	f.StringVar(&s.FileStoragePath, "f", s.FileStoragePath, "FileStoragePath")
	f.StringVar(&s.DatabaseDSN, "d", s.DatabaseDSN, "DatabaseDSN")
	f.StringVar(&s.Key, "k", s.Key, "Key")
	f.StringVar(&s.CryptoKey, "crypto-key", s.CryptoKey, "CryptoKey")
	f.StringVar(&s.AuditFile, "audit-file", s.AuditFile, "AuditFile")
	f.StringVar(&s.AuditURL, "audit-url", s.AuditURL, "AuditURL")
	f.BoolVar(&s.Restore, "r", s.Restore, "Restore")
	f.Var(&s.RetryDuration, "retry-duration", "RetryDuration")
	f.Var(&s.RetryInterval, "retry-interval", "RetryInterval")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}
	f.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "d":
			if s.DatabaseDSN != "" {
				s.ValidDB = true
			}
		case "f":
			if s.FileStoragePath != "" {
				s.ValidFile = true
			}
		case "r":
			s.ValidFile = true
		case "audit-file":
			if s.AuditFile != "" {
				s.ValidFileAudit = true
			}
		case "audit-url":
			if s.AuditURL != "" {
				s.ValidURLAudit = true
			}
		}
	})
	if err := env.Parse(s); err != nil {
		return err
	}
	dsn, envDB := os.LookupEnv("DATABASE_DSN")
	if envDB && dsn != "" {
		s.ValidDB = true
	}
	path, envPath := os.LookupEnv("FILE_STORAGE_PATH")
	_, envRestore := os.LookupEnv("RESTORE")
	if (envPath && path != "") || envRestore {
		s.ValidFile = true
	}
	pathAudit, envPathAudit := os.LookupEnv("AUDIT_FILE")
	url, envURL := os.LookupEnv("AUDIT_URL")
	if envPathAudit && pathAudit != "" {
		s.ValidFileAudit = true
	}
	if envURL && url != "" {
		s.ValidURLAudit = true
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
