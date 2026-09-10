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

// ConfigAgent определяет конфигурацию агента.
type ConfigAgent struct {
	// Net содержит сетевой адрес для запуска агента.
	Net NetAddress `env:"ADDRESS" json:"address"`
	// PollInterval определяет интервал сбора метрик.
	PollInterval DurationSeconds `env:"POLL_INTERVAL" json:"poll_interval"`
	// ReportInterval определяет интервал отправки метрик на сервер.
	ReportInterval DurationSeconds `env:"REPORT_INTERVAL" json:"report_interval"`
	// Key содержит ключ для подписи данных.
	Key string `env:"KEY"`
	// CryptoKey путь до файла с публичным ключом.
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	// RateLimit ограничивает количество исходящих запросов.
	RateLimit int `env:"RATE_LIMIT"`
}

// Get парсит конфигурацию агента.
func (a *ConfigAgent) Get() error {
	if path := GetPath(); path != "" {
		if err := LoadJSON(path, a); err != nil {
			return err
		}
	}
	f := flag.NewFlagSet("Run agent", flag.ContinueOnError)
	var dummy string
	f.StringVar(&dummy, "c", "", "Path to config file")
	f.StringVar(&dummy, "config", "", "Path to config file")
	f.Var(&a.Net, "a", "Net address host:port")
	f.Var(&a.ReportInterval, "r", "ReportInterval")
	f.Var(&a.PollInterval, "p", "PollInterval")
	f.IntVar(&a.RateLimit, "l", a.RateLimit, "RateLimit")
	f.StringVar(&a.Key, "k", a.Key, "Key")
	f.StringVar(&a.CryptoKey, "crypto-key", a.CryptoKey, "CryptoKey")
	if err := f.Parse(os.Args[1:]); err != nil {
		return err
	}
	if err := env.Parse(a); err != nil {
		return err
	}
	if a.PollInterval == 0 {
		return errors.New("pollInterval не может быть нулем")
	}
	if a.PollInterval > a.ReportInterval {
		return errors.New("pollInterval не может быть больше reportInterval")
	}
	return nil
}

// GetKey считывает PEM-файл с диска, декодирует его и парсит публичный RSA-ключ.
func (a *ConfigAgent) GetKey() (*rsa.PublicKey, error) {
	data, err := os.ReadFile(a.CryptoKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("*pem.Block равен nil")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("некорректный RSA-ключ")
	}
	return publicKey, nil
}
