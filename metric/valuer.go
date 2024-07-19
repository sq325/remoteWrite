package metric

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Valuer is a getter to get the value of a metric
type Valuer interface {
	GetValue() float64
}

type gaugeValuer struct{ d *dto.Metric }
type counterValuer struct{ d *dto.Metric }

// Support prometheus.Gauge and prometheus.Counter
func NewValuer(m prometheus.Metric) Valuer {
	d := &dto.Metric{}
	m.Write(d)

	switch m.(type) {
	case prometheus.Gauge:
		return gaugeValuer{d}
	case prometheus.Counter:
		return counterValuer{d}
	default:
		return nil
	}
}

func (g gaugeValuer) GetValue() float64 {
	return g.d.GetGauge().GetValue()
}

func (c counterValuer) GetValue() float64 {
	return c.d.GetCounter().GetValue()
}

// GetMetricVal is a helper function to get the value of a metric
func GetMetricValue(pm prometheus.Metric) (float64, error) {
	m := NewValuer(pm)
	if m == nil {
		return 0, errors.New("unsupported metric type, only Gauge and Counter are supported")
	}
	return m.GetValue(), nil
}

// VecType is a prometheus.MetricVec
type IVec interface {
	GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error)
}

// Vec implement IVec interface
type Vec struct {
	name        string // metric name
	labels      []string
	labelvalues [][]string
	vec         IVec
}

func NewVec(name string, labels []string, vec IVec) *Vec {
	return &Vec{
		name:        name,
		labels:      labels,
		labelvalues: [][]string{},
		vec:         vec,
	}
}

func (v *Vec) Name() string {
	return v.name
}

func (v *Vec) Labels() []string {
	return v.labels
}

func (v *Vec) LabelValues() [][]string {
	return v.labelvalues
}

func (v *Vec) Set(labelvalues []string, value float64) {
	switch any(v.vec).(type) {
	case *counterVec:
		any(v.vec).(*counterVec).cv.WithLabelValues(labelvalues...).Add(value)
	case *gaugeVec:
		any(v.vec).(*gaugeVec).gv.WithLabelValues(labelvalues...).Set(value)
	}
	v.labelvalues = append(v.labelvalues, labelvalues)
}

func (v *Vec) Add(labelvalues []string, value float64) {
	switch any(v.vec).(type) {
	case *counterVec:
		any(v.vec).(*counterVec).cv.WithLabelValues(labelvalues...).Add(value)
	case *gaugeVec:
		any(v.vec).(*gaugeVec).gv.WithLabelValues(labelvalues...).Add(value)
	}
	v.labelvalues = append(v.labelvalues, labelvalues)
}

func (v *Vec) Inc(labelvalues []string) {
	switch any(v.vec).(type) {
	case *counterVec:
		any(v.vec).(*counterVec).cv.WithLabelValues(labelvalues...).Inc()
	case *gaugeVec:
		any(v.vec).(*gaugeVec).gv.WithLabelValues(labelvalues...).Inc()
	}
	v.labelvalues = append(v.labelvalues, labelvalues)
}

func (v *Vec) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	return v.vec.GetMetricWithLabelValues(lvs...)
}

// gaugeVec is a wrapper for prometheus.GaugeVec
// gaugeVec implement IVec interface
type gaugeVec struct {
	gv *prometheus.GaugeVec
}

func (mt *gaugeVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	m, err := mt.gv.GetMetricWithLabelValues(lvs...)
	return m.(prometheus.Metric), err
}

type counterVec struct {
	cv *prometheus.CounterVec
}

func (mt *counterVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	m, err := mt.cv.GetMetricWithLabelValues(lvs...)
	return m.(prometheus.Metric), err
}
