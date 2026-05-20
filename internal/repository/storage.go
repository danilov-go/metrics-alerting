package repository

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

type Storage interface {
	SaveGauges(name string, value float64)
	SaveCounters(name string, value int64)
	GetGauges(name string) (float64, bool)
	GetCounters(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
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
