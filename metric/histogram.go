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
	TimeSeries() []*prompb.TimeSeries
	Labels() []string
	LabelValues() [][]string
}

// PBHistogram wraps a prometheus.CounterVec
// PBHistogram implements PBCounterMeter interface
type PBHistogram struct {
	vec     *Vec      // bucket_label must be included in end of labels
	buckets []float64 // must sorted by ascending
}

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

func (hg *PBHistogram) TimeSeries() []*prompb.TimeSeries {
	// labels := hg.Labels()
	// tsList := make([]*prompb.TimeSeries, 0, len(hg.LabelValues()))
	// for _, lvs := range hg.LabelValues() {
	// 	if len(labels) != len(lvs) {
	// 		log.Println("labels and labelvalues not match")
	// 		continue
	// 	}

	// 	ts := &prompb.TimeSeries{
	// 		Labels: make([]*prompb.Label, 0, len(labels)),
	// 		Samples: []*prompb.Sample{
	// 			{
	// 				Value:     0,
	// 				Timestamp: 0,
	// 			},
	// 		},
	// 	}

	// }
	return nil
}

// lvs 不包含 bucket_label
func (hg *PBHistogram) Observe(lvs []string, value float64) {
	b := findBucket(hg.buckets, value)
	if b <= 0 {
		log.Println("no bucket for value:", value)
		return
	}
	lvs = append(lvs, strconv.FormatFloat(b, 'f', -1, 64))
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
