package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

type NetAddress struct {
	Host string
	Port int
}

type ConfigAgent struct {
	Net            NetAddress `env:"ADDRESS"`
	PollInterval   int        `env:"POLL_INTERVAL"`
	ReportInterval int        `env:"REPORT_INTERVAL"`
}

type ConfigServer struct {
	Net             NetAddress `env:"ADDRESS"`
	StoreIntrval    int        `env:"STORE_INTERVAL"`
	FileStoragePath string     `env:"FILE_STORAGE_PATH"`
	Restore         bool       `env:"RESTORE"`
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
	f.BoolVar(&s.Restore, "r", s.Restore, "Restore")
	err := f.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = env.Parse(s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (a *ConfigAgent) Get() {
	f := flag.NewFlagSet("Run agent", flag.ContinueOnError)
	f.Var(&a.Net, "a", "Net address host:port")
	f.IntVar(&a.ReportInterval, "r", a.ReportInterval, "ReportInterval")
	f.IntVar(&a.PollInterval, "p", a.PollInterval, "PollInterval")
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
