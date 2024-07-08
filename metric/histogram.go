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
	Observe(lvs []string, value float64)
	Name() string
	Labels() []string
	LabelValues() [][]string
	GetBucketValue(lvs []string) (float64, error)
}

// A PBHistogram is a series of histogram metric with different pb.Labels
// pb.Labels have same names but different values
// PBHistogram implements PBCounterMeter interface
type PBHistogram struct {
	vec     *Vec      // bucket_label must be included in end of labels
	buckets []float64 // must sorted by ascending
	count   int
	sum     float64
}

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
	n := len(hg.vec.LabelValues())
	if n == 0 {
		return nil
	}

	tsList := make([]*prompb.TimeSeries, 0, n)
	// A lvs generate a TimeSeries
	for _, lvs := range hg.vec.LabelValues() {
		if len(lvs) != len(hg.vec.Labels()) {
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

		labels := make([]*prompb.Label, 0, len(hg.vec.Labels()))
		{
			for i, label := range hg.vec.Labels() {
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

// lvs 不包含 bucket_label
func (hg *PBHistogram) Observe(lvs []string, value float64) {
	b := findBucket(hg.buckets, value)
	if b <= 0 {
		log.Println("no bucket for value:", value)
		return
	}
	lvs = append(lvs, strconv.FormatFloat(b, 'f', -1, 64)) // add bucket_label
	hg.vec.Add(lvs, value)
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

func (hg *PBHistogram) Labels() []string {
	return hg.vec.Labels()
}

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
func (hg *PBHistogram) GetCount() float64 {
	return float64(hg.count)
}

func (hg *PBHistogram) SetCount(v int) {
	hg.count = v
}

// XXX_sum
func (hg *PBHistogram) GetSum() float64 {
	return hg.sum
}

func (hg *PBHistogram) SetSum(v float64) {
	hg.sum = v
}
