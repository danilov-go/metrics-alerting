package repository

import (
	"bufio"
	"encoding/json"
	"os"
)

type MemStorage struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

func (g *MemStorage) SaveGauges(name string, value float64) {
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	g.Gauges[name] = value
}

func (c *MemStorage) SaveCounters(name string, value int64) {
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	if val, ok := c.Counters[name]; ok {
		c.Counters[name] = value + val
	} else {
		c.Counters[name] = value
	}
}

func (g *MemStorage) GetGauges(name string) (float64, bool) {
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	val, ok := g.Gauges[name]
	return val, ok
}

func (c *MemStorage) GetCounters(name string) (int64, bool) {
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	val, ok := c.Counters[name]
	return val, ok
}

func (g *MemStorage) GetAllGauges() map[string]float64 {
	if g.Gauges == nil {
		g.Gauges = make(map[string]float64)
	}
	return g.Gauges
}

func (c *MemStorage) GetAllCounters() map[string]int64 {
	if c.Counters == nil {
		c.Counters = make(map[string]int64)
	}
	return c.Counters
}

func (c *MemStorage) SaveFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := bufio.NewWriter(file)
	err = json.NewEncoder(buf).Encode(c)
	if err != nil {
		return err

	}
	return buf.Flush()
}

func (c *MemStorage) LoadFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	err = json.NewDecoder(buf).Decode(c)
	if err != nil {
		return err
	}
	return nil
}
