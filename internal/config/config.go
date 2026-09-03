// Package config управляет конфигурацией приложения.
package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v11"
)

// NetAddress определяет адрес сервера.
type NetAddress struct {
	Host string
	Port int
}

// ConfigAgent определяет конфигурацию агента.
type ConfigAgent struct {
	// Net содержит сетевой адрес для запуска агента.
	Net NetAddress `env:"ADDRESS"`
	// PollInterval определяет интервал сбора метрик.
	PollInterval int `env:"POLL_INTERVAL"`
	// ReportInterval определяет интервал отправки метрик на сервер.
	ReportInterval int `env:"REPORT_INTERVAL"`
	// Key содержит ключ для подписи данных.
	Key string `env:"KEY"`
	// RateLimit ограничивает количество исходящих запросов.
	RateLimit int `env:"RATE_LIMIT"`
}

// ConfigServer определяет конфигурацию сервера.
type ConfigServer struct {
	// Net содержит сетевой адрес для запуска сервера.
	Net NetAddress `env:"ADDRESS"`
	// StoreInterval определяет интервал времени для сохранения метрик на диск.
	StoreIntrval int `env:"STORE_INTERVAL"`
	// FileStoragePath определяет путь к файлу, куда сохраняются метрики.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	// Restore определяет, нужно ли загружать сохранённые метрики из файла при старте сервера.
	Restore bool `env:"RESTORE"`
	// DatabaseDSN содержит строку подключения к базе данных PostgreSQL.
	DatabaseDSN string `env:"DATABASE_DSN"`
	// Key содержит ключ для подписи данных.
	Key string `env:"KEY"`
	// AuditFile определяет путь к файлу логов аудита.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL содержит URL-адрес внешнего сервиса аудита.
	AuditURL string `env:"AUDIT_URL"`
	// RetryDuration определяет продолжительность попыток повтора операций.
	RetryDuration int `env:"RETRY_DURATION"`
	// RetryInterval определяет интервал между повторными попытками выполнения операций.
	RetryInterval int `env:"RETRY_INTERVAL"`
	// ValidDB флаг, указывающий на корректность строки подключения к базе данных.
	ValidDB bool
	// ValidFile флаг, указывающий на корректность пути к файлу хранилища метрик.
	ValidFile bool
	// ValidFileAudit флаг, указывающий на корректность и доступность файла аудита.
	ValidFileAudit bool
	// ValidURLAudit флаг, указывающий на корректность и доступность URL аудита.
	ValidURLAudit bool
}

// String возвращает строковое представление сетевого адреса в формате host:port.
func (n NetAddress) String() string {
	return n.Host + ":" + strconv.Itoa(n.Port)
}

// UnmarshalText десериализует сетевой адрес из текстового формата для библиотеки env.
func (n *NetAddress) UnmarshalText(adr []byte) error {
	return n.Set(string(adr))
}

// Set парсит строку в формате host:port и валидирует ее для пакета flag.
func (n *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	n.Host = hp[0]
	n.Port = port
	return nil
}

// Get парсит конфигурацию сервера.
func (s *ConfigServer) Get() {
	f := flag.NewFlagSet("Run server", flag.ContinueOnError)
	f.Var(&s.Net, "a", "Net address host:port")
	f.IntVar(&s.StoreIntrval, "i", s.StoreIntrval, "StoreIntrval")
	f.StringVar(&s.FileStoragePath, "f", s.FileStoragePath, "FileStoragePath")
	f.StringVar(&s.DatabaseDSN, "d", s.DatabaseDSN, "DatabaseDSN")
	f.StringVar(&s.Key, "k", s.Key, "Key")
	f.StringVar(&s.AuditFile, "audit-file", s.AuditFile, "AuditFile")
	f.StringVar(&s.AuditURL, "audit-url", s.AuditURL, "AuditURL")
	f.BoolVar(&s.Restore, "r", s.Restore, "Restore")
	f.IntVar(&s.RetryDuration, "retry-duration", s.RetryDuration, "RetryDuration")
	f.IntVar(&s.RetryInterval, "retry-interval", s.RetryInterval, "RetryInterval")
	err := f.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
	err = env.Parse(s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
}

// Get парсит конфигурацию агента.
func (a *ConfigAgent) Get() {
	f := flag.NewFlagSet("Run agent", flag.ContinueOnError)
	f.Var(&a.Net, "a", "Net address host:port")
	f.IntVar(&a.ReportInterval, "r", a.ReportInterval, "ReportInterval")
	f.IntVar(&a.PollInterval, "p", a.PollInterval, "PollInterval")
	f.IntVar(&a.RateLimit, "l", a.RateLimit, "RateLimit")
	f.StringVar(&a.Key, "k", a.Key, "Key")
	err := f.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = env.Parse(a)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if a.PollInterval == 0 {
		fmt.Println("pollInterval не может быть нулем")
		os.Exit(1)
	}
	if a.PollInterval > a.ReportInterval {
		fmt.Println("pollInterval не может быть больше reportInterval")
		os.Exit(1)
	}
}
