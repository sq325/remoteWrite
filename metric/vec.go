package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Item is a single item
// A item represents a MetricVec with the same labels
type Item interface {
	// Instrument return prometheus client_golang MetricType
	Instrument() MetricType
	Name() string

	// For get vec values
	Labels() []string // 一个 Item 中 多个 T 的 Labels 必须相同
	LabelValues() [][]string
}

type MetricType interface {
	GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error)
}

// metricsVec is a wrapper for prometheus.GaugeVec or prometheus.ConterVec
// metricsVec implement Item interface
type MetricVec struct {
	name        string
	labels      []string
	labelvalues [][]string
	vec         MetricType
}

func NewMetricVec(name string, labels []string, vec MetricType) *MetricVec {
	return &MetricVec{
		name:        name,
		labels:      labels,
		labelvalues: [][]string{},
		vec:         vec,
	}
}

func (m *MetricVec) Instrument() MetricType {
	return m.vec
}

func (m *MetricVec) Name() string {
	return m.name
}

func (m *MetricVec) Labels() []string {
	return m.labels
}

func (m *MetricVec) LabelValues() [][]string {
	return m.labelvalues
}

func (m *MetricVec) Set(labelvalues []string, value float64) {
	switch any(m.vec).(type) {
	case *GvMetricType:
		any(m.vec).(*GvMetricType).Vec.WithLabelValues(labelvalues...).Set(value)
	case *CvMetricType:
		any(m.vec).(*CvMetricType).Vec.WithLabelValues(labelvalues...).Add(value)
	}
	m.labelvalues = append(m.labelvalues, labelvalues)
}

func (m *MetricVec) Inc(labelvalues []string) {
	switch any(m.vec).(type) {
	case *GvMetricType:
		any(m.vec).(*GvMetricType).Vec.WithLabelValues(labelvalues...).Inc()
	case *CvMetricType:
		any(m.vec).(*CvMetricType).Vec.WithLabelValues(labelvalues...).Inc()
	}
	m.labelvalues = append(m.labelvalues, labelvalues)
}

// GvMetricType is a wrapper for prometheus.GaugeVec
// GvMetricType implement MetricType interface
type GvMetricType struct {
	Vec *prometheus.GaugeVec
}

func (mt *GvMetricType) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	m, err := mt.Vec.GetMetricWithLabelValues(lvs...)
	return m.(prometheus.Metric), err
}

type CvMetricType struct {
	Vec *prometheus.CounterVec
}

func (mt *CvMetricType) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	m, err := mt.Vec.GetMetricWithLabelValues(lvs...)
	return m.(prometheus.Metric), err
}
