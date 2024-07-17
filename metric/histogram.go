package metric

import (
	"log"
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
	Add(lvs []string, value float64, le float64, sum float64)
	Observe(lvs []string, value float64)
	Reset()
}

type HistogramReader interface {
	Name() string
	Labels() []string
	LabelValues() [][]string
	Buckets() []float64
	Count() int
	Sum() float64
}

// A PBHistogram is composed of several bucket metrics and a sum and a count metric
// labelValues should all have the same value except for 'le'
// PBHistogram implements HistogramMeter
type PBHistogram struct {
	vec     *Vec      // bucket_label must be included in end of labels
	buckets []float64 // must sorted by ascending
	count   int
	sum     float64
}

var _ HistogramMeter = (*PBHistogram)(nil)

// labels must not include bucket_label
func NewPBHistogram(name string, help string, labels []string, buckets []float64) *PBHistogram {
	if len(buckets) == 0 {
		buckets = defaultBuckets
	}

	// add bucket_label
	labels = append(labels, bucket_label)
	if buckets == nil {
		buckets = defaultBuckets
	}

	vec := NewVec(
		name,
		labels,
		&counterVec{
			cv: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: name,
					Help: help,
				},
				labels,
			),
		},
	)

	return &PBHistogram{
		vec:     vec,
		buckets: buckets,
	}
}

// Name return the name of histogram without _bucket suffix
func (hg *PBHistogram) Name() string {
	return hg.vec.Name()
}

// Implement PBMetric interface
// TODO:
// 1. 生成 {name}_sum 和 {name}_count 指标
// 2. 生成 bucket 指标
// 3. 生成 []*prompb.TimeSeries
// timestamp: timestamp is in ms format
func (hg *PBHistogram) TimeSeries(timestamp int64) []*prompb.TimeSeries {
	n := len(hg.LabelValues())
	if n == 0 {
		return nil
	}

	tsList := make([]*prompb.TimeSeries, 0, n+2) // +2 for sum and count
	// A lvs generate a TimeSeries
	// lvs containers bucket_label
	for _, lvs := range hg.LabelValues() {
		if len(lvs) != len(hg.Labels()) {
			log.Println("labels and labelvalues not match")
			continue
		}
		m, err := hg.vec.GetMetricWithLabelValues(lvs...)
		if err != nil {
			continue
		}
		v, err := GetMetricValue(m)
		if err != nil {
			log.Println(err)
		}

		// __name__ is the first label
		// bucket_label is the last label
		pblabels := prompbLabels(hg.Name()+"_bucket", hg.Labels(), lvs)

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

	// sum and count
	lvs := hg.LabelValues()[0]
	lvs = lvs[:len(lvs)-1]                 // remove bucket_label
	lv := hg.Labels()[:len(hg.Labels())-1] // remove bucket_label
	if len(lvs) != len(lv) {
		log.Println("labels and labelvalues not match, can not generate sum and count")
	}

	var ts_sum *prompb.TimeSeries
	{
		pblabels := prompbLabels(hg.Name()+"_sum", lv, lvs)
		sample := &prompb.Sample{
			Value:     hg.Sum(),
			Timestamp: timestamp,
		}
		ts_sum = &prompb.TimeSeries{
			Labels:  pblabels,
			Samples: []*prompb.Sample{sample},
		}
	}

	var ts_count *prompb.TimeSeries
	{
		pblabels := prompbLabels(hg.Name()+"_count", lv, lvs)
		sample := &prompb.Sample{
			Value:     float64(hg.Count()),
			Timestamp: timestamp,
		}
		ts_count = &prompb.TimeSeries{
			Labels:  pblabels,
			Samples: []*prompb.Sample{sample},
		}
	}

	tsList = append(tsList, ts_sum, ts_count)
	return tsList
}

// prompbLabels generates []*prompb.Label based on lv and lvs, and adds the name label
func prompbLabels(name string, lv, lvs []string) []*prompb.Label {
	if len(lv) != len(lvs) {
		log.Println("labels and labelvalues not match")
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
		log.Println("no bucket for value:", value)
		return
	}
	lvs = append(lvs, strconv.FormatFloat(b, 'f', -1, 64)) // add bucket_label
	hg.vec.Add(lvs, value)
	hg.count++      // update count
	hg.sum += value // update sum
}

// Add add the value to the corresponding bucket.
// lvs must not include bucket_label
func (hg *PBHistogram) Add(lvs []string, value float64, le float64, sum float64) {
	lvs = append(lvs, strconv.FormatFloat(le, 'f', -1, 64)) // add bucket_label
	hg.vec.Add(lvs, value)
	hg.AddCount(int(value)) // update count
	hg.AddSum(sum)          // update sum
}

func (hg *PBHistogram) AddCount(c int) {
	hg.count += c
}
func (hg *PBHistogram) AddSum(s float64) {
	hg.sum += s
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

func (hg *PBHistogram) GetBucketValue(lvs []string) (float64, error) {
	m, err := hg.vec.GetMetricWithLabelValues(lvs...)
	if err != nil {
		return 0, err
	}
	return GetMetricValue(m)
}

// XXX_count
func (hg *PBHistogram) Count() int {
	return hg.count
}

// XXX_sum
func (hg *PBHistogram) Sum() float64 {
	return hg.sum
}
