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

type NetAddress struct {
	Host string
	Port int
}

type ConfigAgent struct {
	Net            NetAddress `env:"ADDRESS"`
	PollInterval   int        `env:"POLL_INTERVAL"`
	ReportInterval int        `env:"REPORT_INTERVAL"`
	Key            string     `env:"KEY"`
	RateLimit      int        `env:"RATE_LIMIT"`
}

type ConfigServer struct {
	Net             NetAddress `env:"ADDRESS"`
	StoreIntrval    int        `env:"STORE_INTERVAL"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH"`
	Restore         bool       `env:"RESTORE"`
	DatabaseDsn     string     `env:"DATABASE_DSN"`
	Key             string     `env:"KEY"`
	AuditFile       string     `env:"AUDIT_FILE"`
	AuditUrl        string     `env:"AUDIT_URL"`
	RetryDuration   int        `env:"RETRY_DURATION"`
	RetryInterval   int        `env:"RETRY_INTERVAL"`
	ValidDB         bool
	ValidFile       bool
	ValidFileAudit  bool
	ValidUrlAudit   bool
}

func (n NetAddress) String() string {
	return n.Host + ":" + strconv.Itoa(n.Port)
}

func (n *NetAddress) UnmarshalText(adr []byte) error {
	return n.Set(string(adr))
}
func (n *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("Need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	n.Host = hp[0]
	n.Port = port
	return nil
}

func (s *ConfigServer) Get() {
	f := flag.NewFlagSet("Run server", flag.ContinueOnError)
	f.Var(&s.Net, "a", "Net address host:port")
	f.IntVar(&s.StoreIntrval, "i", s.StoreIntrval, "StoreIntrval")
	f.StringVar(&s.FileStoragePath, "f", s.FileStoragePath, "FileStoragePath")
	f.StringVar(&s.DatabaseDsn, "d", s.DatabaseDsn, "DatabaseDsn")
	f.StringVar(&s.Key, "k", s.Key, "Key")
	f.StringVar(&s.AuditFile, "audit-file", s.AuditFile, "AuditFile")
	f.StringVar(&s.AuditUrl, "audit-url", s.AuditUrl, "AuditUrl")
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
			if s.DatabaseDsn != "" {
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
			if s.AuditUrl != "" {
				s.ValidUrlAudit = true
			}
		}
	})
	err = env.Parse(s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dsn, envDb := os.LookupEnv("DATABASE_DSN")
	if envDb == true && dsn != "" {
		s.ValidDB = true
	}
	path, envPath := os.LookupEnv("FILE_STORAGE_PATH")
	_, envRestore := os.LookupEnv("RESTORE")
	if (envPath && path != "") || envRestore {
		s.ValidFile = true
	}
	pathAudit, envPathAudit := os.LookupEnv("AUDIT_FILE")
	url, envUrl := os.LookupEnv("AUDIT_URL")
	if envPathAudit && pathAudit != "" {
		s.ValidFileAudit = true
	}
	if envUrl && url != "" {
		s.ValidUrlAudit = true
	}
}

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
