package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

type ConfigAgent struct {
	Net            NetAddress
	PollInterval   int
	ReportInterval int
}

type ConfigServer struct {
	Net NetAddress
}

func (n NetAddress) String() string {
	return n.Host + ":" + strconv.Itoa(n.Port)
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
	err := f.Parse(os.Args[1:])
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.Net.Set(envRunAddr)
	}
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
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		a.Net.Set(envRunAddr)
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		reportInterval, err := strconv.Atoi(envReportInterval)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		a.ReportInterval = reportInterval
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		pollInterval, err := strconv.Atoi(envPollInterval)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		a.PollInterval = pollInterval
	}
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
