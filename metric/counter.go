package metric

import (
	"github.com/sq325/remoteWrite/prompb"
)

type PbCounter struct {
	vec *Vec
}

func (c *PbCounter) TimeSeries() []*prompb.TimeSeries {
	return nil
}

func (c *PbCounter) Add(lvs []string, value float64) {
	c.vec.Add(lvs, value)
}

func (c *PbCounter) Inc(lvs []string) {
	c.vec.Inc(lvs)
}

func (c *PbCounter) GetValues(lvs []string) (float64, error) {
	m, err := c.vec.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	return GetMetricValue(m)
}
