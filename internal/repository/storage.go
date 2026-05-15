package repository

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

type Storage interface {
	SaveGauges(name string, value float64)
	SaveCounters(name string, value int64)
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
