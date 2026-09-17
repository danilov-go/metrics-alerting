// Package config управляет конфигурацией приложения.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// NetAddress определяет адрес сервера.
type NetAddress struct {
	Host string
	Port int
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

// UnmarshalJSON десериализует сетевой адрес из JSON формата.
func (n *NetAddress) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return n.Set(str)
}

// DurationSeconds определяет интервал времени в секундах.
type DurationSeconds int

// String возвращает строковое представление интервала времени в секундах.
func (d DurationSeconds) String() string {
	return strconv.Itoa(int(d))
}

// Set парсит строку времени или просто число.
func (d *DurationSeconds) Set(str string) error {
	if strings.HasSuffix(str, "s") || strings.HasSuffix(str, "m") || strings.HasSuffix(str, "h") {
		duration, err := time.ParseDuration(str)
		if err != nil {
			return err
		}
		*d = DurationSeconds(duration.Seconds())
		return nil
	}
	num, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*d = DurationSeconds(num)
	return nil
}

// UnmarshalJSON десериализует интервал времени из JSON формата.
func (d *DurationSeconds) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		return d.Set(str)
	}
	var num int
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*d = DurationSeconds(num)
	return nil
}

// UnmarshalText десериализует интервал времени из текстового формата для библиотеки env.
func (d *DurationSeconds) UnmarshalText(data []byte) error {
	return d.Set(string(data))
}

// GetPath определяет путь к файлу конфигурации.
func GetPath() string {
	if path := os.Getenv("CONFIG"); path != "" {
		return path
	}
	for i := 1; i < len(os.Args)-1; i++ {
		if os.Args[i] == "-c" || os.Args[i] == "-config" {
			return os.Args[i+1]
		}
	}
	return ""
}

// LoadJSON десериализует конфигурацию из JSON файла.
func LoadJSON[T any](path string, t *T) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, t); err != nil {
		return err
	}
	return nil
}

// PrintBuild выводит в консоль информацию о текущей сборке приложения.
func PrintBuild(w io.Writer, version, date, commit string) {
	fmt.Fprintf(w, "Build version: %s\n", version)
	fmt.Fprintf(w, "Build date: %s\n", date)
	fmt.Fprintf(w, "Build commit: %s\n", commit)
}
