package metric

import (
	"log"

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

func NewPBCounter(name string, help string, labels []string) *PBCounter {
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

// timestamp: timestamp is in ms format
func (c *PBCounter) TimeSeries(timestamp int64) []*prompb.TimeSeries {
	n := len(c.vec.LabelValues())
	if n == 0 {
		return nil
	}

	tsList := make([]*prompb.TimeSeries, 0, n)
	// A lvs generate a TimeSeries
	for _, lvs := range c.vec.LabelValues() {
		if len(lvs) != len(c.vec.Labels()) {
			log.Println("labels and labelvalues not match")
			continue
		}
		m, err := c.vec.GetMetricWithLabelValues(lvs...)
		if err != nil {
			continue
		}
		v, err := GetMetricValue(m)
		if err != nil {
			log.Println(err)
		}

		labels := make([]*prompb.Label, 0, len(c.vec.Labels())+1) // +1 for __name__
		{
			labels = append(labels, &prompb.Label{
				Name:  "__name__",
				Value: c.vec.Name(),
			})
			for i, label := range c.vec.Labels() {
				labels = append(labels, &prompb.Label{
					Name:  label,
					Value: lvs[i],
				})
			}
		}

		sample := &prompb.Sample{
			Value:     v,
			Timestamp: timestamp,
		}

		ts := &prompb.TimeSeries{
			Labels:  labels,
			Samples: []*prompb.Sample{sample},
		}
		tsList = append(tsList, ts)
	}

	return tsList
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
