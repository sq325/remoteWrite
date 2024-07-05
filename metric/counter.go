package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sq325/remoteWrite/prompb"
)

type PBCounterMeter interface {
	PBMetric
	Add(lvs []string, value float64)
	Inc(lvs []string)
	GetValue(lvs []string) (float64, error)
}

type PBCounter struct {
	vec *Vec
}

func (c *PBCounter) NewPBCounter(name string, help string, labels []string) *PBCounter {
	return &PBCounter{
		vec: NewVec(name, labels,
			&counterVec{
				cv: prometheus.NewCounterVec(
					prometheus.CounterOpts{
						Name: name,
						Help: help,
					},
					labels,
				),
			},
		),
	}
}

// TODO: implement
func (c *PBCounter) TimeSeries() []*prompb.TimeSeries {
	return nil
}

func (c *PBCounter) Add(lvs []string, value float64) {
	c.vec.Add(lvs, value)
}

func (c *PBCounter) Inc(lvs []string) {
	c.vec.Inc(lvs)
}

func (c *PBCounter) GetValue(lvs []string) (float64, error) {
	m, err := c.vec.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	return GetMetricValue(m)
}
