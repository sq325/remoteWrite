package metric

import (
	"github.com/sq325/remoteWrite/prompb"
)

type PBCounterMeter interface {
	PBMetric
	Add(lvs []string, value float64)
}

type PBCounter struct {
	vec *Vec
}

func (c *PBCounter) TimeSeries() []*prompb.TimeSeries {
	return nil
}

func (c *PBCounter) Add(lvs []string, value float64) {
	c.vec.Add(lvs, value)
}

func (c *PBCounter) Inc(lvs []string) {
	c.vec.Inc(lvs)
}

func (c *PBCounter) GetValues(lvs []string) (float64, error) {
	m, err := c.vec.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	return GetMetricValue(m)
}
