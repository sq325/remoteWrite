package metric

import (
	"log/slog"
	"math"
	"slices"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sq325/remoteWrite/prompb"
)

const (
	bucket_label = "le"
)

var (
	defaultBuckets = []float64{25, 50, 75, 100, 200, 500, 1000, math.Inf(1)}
)

type HistogramMeter interface {
	PBMetric

	HistogramReader
	HistogramWriter
}

type HistogramWriter interface {
	Add(lvs []string, value float64, le float64)
	AddSum(lvs []string, value float64)
	Observe(lvs []string, value float64)
	Reset()
}

type HistogramReader interface {
	Name() string
	Labels() []string
	LabelValues() [][]string
	Buckets() []float64
}

// A PBHistogram is composed of several bucket metrics and a sum and a count metric
// labelValues should all have the same value except for 'le'
// PBHistogram implements HistogramMeter
type PBHistogram struct {
	vec     *Vec      // bucket_label must be included in end of labels
	buckets []float64 // must sorted by ascending
	count   *Vec
	sum     *Vec
}

var _ HistogramMeter = (*PBHistogram)(nil)

// name is the name of histogram without _bucket suffix
// labels must not include bucket_label le
func NewPBHistogram(name string, help string, labels []string, buckets []float64) *PBHistogram {
	if len(buckets) == 0 {
		buckets = defaultBuckets
	}

	vecCount := NewVec(
		name+"_count",
		labels,
		&counterVec{
			cv: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: name + "_count",
					Help: help,
				},
				labels,
			),
		},
	)

	vecSum := NewVec(
		name+"_sum",
		labels,
		&counterVec{
			cv: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: name + "_sum",
					Help: help,
				},
				labels,
			),
		},
	)

	// add bucket_label
	labels = append(labels, bucket_label)
	if buckets == nil {
		buckets = defaultBuckets
	}

	vec := NewVec(
		name+"_bucket",
		labels,
		&counterVec{
			cv: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: name + "_bucket",
					Help: help,
				},
				labels,
			),
		},
	)

	return &PBHistogram{
		vec:     vec,
		buckets: buckets,
		count:   vecCount,
		sum:     vecSum,
	}
}

// Name return the name of metric
func (hg *PBHistogram) Name() string {
	return hg.vec.Name()
}

// Implement PBMetric interface
// timestamp: timestamp is in ms format
func (hg *PBHistogram) TimeSeries(timestamp int64) []*prompb.TimeSeries {
	cn := len(hg.count.LabelValues())
	if cn == 0 {
		return nil
	}

	tsGenerator := func(vec *Vec) []*prompb.TimeSeries {
		tsList := make([]*prompb.TimeSeries, 0, len(vec.LabelValues()))
		for _, lvs := range vec.LabelValues() {
			if len(lvs) != len(vec.Labels()) {
				slog.Error("labels and labelvalues not match", "labels", vec.Labels(), "labelvalues", lvs)
				continue
			}
			m, err := vec.GetMetricWithLabelValues(lvs...)
			if err != nil {
				continue
			}
			v, err := GetMetricValue(m)
			if err != nil {
				slog.Error("GetMetricValue failed", "err", err)
			}

			pblabels := prompbLabels(vec.Name(), vec.Labels(), lvs)
			sample := &prompb.Sample{
				Value:     v,
				Timestamp: timestamp,
			}
			ts := &prompb.TimeSeries{
				Labels:  pblabels,
				Samples: []*prompb.Sample{sample},
			}
			tsList = append(tsList, ts)
		}
		return tsList
	}

	tsList := make([]*prompb.TimeSeries, 0, cn+cn+cn+cn*len(hg.buckets)) // sum, count, inf = cn, cn*len(buckets) = bucket
	tsList = append(tsList, tsGenerator(hg.vec)...)

	var ts_sum = make([]*prompb.TimeSeries, 0, cn)
	ts_sum = append(ts_sum, tsGenerator(hg.sum)...)

	var ts_count = make([]*prompb.TimeSeries, 0, cn)
	tslist := tsGenerator(hg.count)
	ts_count = append(ts_count, tslist...)

	var ts_inf = make([]*prompb.TimeSeries, 0, cn)
	{
		tslist := tsGenerator(hg.count)
		for _, ts := range tslist {
			// 把末尾的_count 改成 _bucket
			for _, l := range ts.Labels {

				if l.Name == "__name__" {
					l.Value = hg.vec.Name()
				}
			}

			// 给每个ts增加le=+Inf
			ts.Labels = append(ts.Labels, &prompb.Label{
				Name:  bucket_label,
				Value: strconv.FormatFloat(math.Inf(1), 'f', -1, 64),
			})
			ts_inf = append(ts_inf, ts)
		}
	}

	tsList = append(tsList, ts_sum...)
	tsList = append(tsList, ts_count...)
	tsList = append(tsList, ts_inf...)

	return tsList
}

// prompbLabels generates []*prompb.Label based on lv and lvs, and adds the name label
func prompbLabels(name string, lv, lvs []string) []*prompb.Label {
	if len(lv) != len(lvs) {
		slog.Error("labels and labelvalues not match", "labels", lv, "labelvalues", lvs)
		return nil
	}

	labels := make([]*prompb.Label, 0, len(lv)+1) // +1 for __name__

	for i, label := range lv {
		labels = append(labels, &prompb.Label{
			Name:  label,
			Value: lvs[i],
		})
	}

	// add __name__
	labels = slices.Insert(labels, 0,
		&prompb.Label{
			Name:  "__name__",
			Value: name,
		},
	)
	return labels
}

// lvs 不包含 bucket_label
// Observe adds a single observation to the histogram.
func (hg *PBHistogram) Observe(lvs []string, value float64) {
	b := findBucket(hg.buckets, value)
	if b <= 0 {
		slog.Error("findBucket failed, no bucket found with the value", "value", value)
		return
	}

	hg.count.Add(lvs, 1)
	hg.sum.Add(lvs, value)

	lvs = append(lvs, strconv.FormatFloat(b, 'f', -1, 64)) // add bucket_label
	hg.vec.Add(lvs, 1)
}

// Add add the value to the corresponding bucket.
// Do not add a bucket with le=+Inf, as the +Inf bucket will be automatically generated in the TimeSeries
// lvs must not include bucket_label
func (hg *PBHistogram) Add(lvs []string, value float64, le float64) {
	hg.count.Add(lvs, value)

	lvs = append(lvs, strconv.FormatFloat(le, 'f', -1, 64)) // add bucket_label
	hg.vec.Add(lvs, value)
}

// func (hg *PBHistogram) AddCount(lvs []string, c int) {
// 	hg.count.Add(lvs, float64(c))
// }

func (hg *PBHistogram) AddSum(lvs []string, s float64) {
	hg.sum.Add(lvs, s)
}

// TODO: eset count and sum to zero, reinitialize vec
func (hg *PBHistogram) Reset() {
}

func (hg *PBHistogram) Buckets() []float64 {
	return hg.buckets
}

// return value:
//  1. <0: error
//  2. 0: no bucket
//  3. >0: bucket value
func findBucket(buckets []float64, value float64) float64 {
	if !slices.IsSorted(buckets) {
		return -1
	}
	for _, b := range buckets {
		if value <= b {
			return b
		}
	}
	return 0
}

// Labels contains bucket_label but not include __name__
// bucket_label is at the end of labels
func (hg *PBHistogram) Labels() []string {
	return hg.vec.Labels()
}

// LabelValues contains bucket_label but not include __name__
// bucket_label value is at the end of labelValues
func (hg *PBHistogram) LabelValues() [][]string {
	return hg.vec.LabelValues()
}

// lvs must not include bucket_label
func (hg *PBHistogram) GetBucketValue(lvs []string) (float64, error) {
	m, err := hg.vec.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	return GetMetricValue(m)
}

// XXX_count
func (hg *PBHistogram) GetCountValue(lvs []string) (int, error) {
	m, err := hg.count.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	f, err := GetMetricValue(m)
	if err != nil {
		return 0, err
	}

	return int(f), nil
}

// XXX_sum
func (hg *PBHistogram) Sum(lvs []string) (float64, error) {
	m, err := hg.sum.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	return GetMetricValue(m)
}
