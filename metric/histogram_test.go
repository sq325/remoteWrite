package metric

import (
	"log"
	"reflect"
	"sort"
	"testing"

	"github.com/sq325/remoteWrite/prompb"
)

func TestPBHistogram_TimeSeries(t *testing.T) {
	hg := NewPBHistogram("test_histogram", "Test Histogram", []string{"label1", "label2"}, []float64{50, 100, 200})
	hg.Add([]string{"value1", "value2"}, 5, 50)
	hg.Add([]string{"value1", "value2"}, 8, 100)
	hg.Add([]string{"value1", "value2"}, 13, 200)
	hg.Add([]string{"value1", "value3"}, 11, 50)
	hg.Add([]string{"value1", "value3"}, 40, 100)
	hg.Add([]string{"value1", "value3"}, 80, 200)
	hg.AddCount([]string{"value1", "value2"}, 14)
	hg.AddCount([]string{"value1", "value3"}, 88)
	hg.AddSum([]string{"value1", "value2"}, 1500)
	hg.AddSum([]string{"value1", "value3"}, 3300)
	got := hg.TimeSeries(1722838400634)

	want := []*prompb.TimeSeries{
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value2",
				},
				{
					Name:  "le",
					Value: "50",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     5,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value2",
				},
				{
					Name:  "le",
					Value: "100",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     8,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value2",
				},
				{
					Name:  "le",
					Value: "200",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     13,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value2",
				},
				{
					Name:  "le",
					Value: "+Inf",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     14,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_sum",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value2",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     1500,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_count",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value2",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     14,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value3",
				},
				{
					Name:  "le",
					Value: "50",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     11,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value3",
				},
				{
					Name:  "le",
					Value: "100",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     40,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value3",
				},
				{
					Name:  "le",
					Value: "200",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     80,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_bucket",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value3",
				},
				{
					Name:  "le",
					Value: "+Inf",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     88,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_sum",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value3",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     3300,
					Timestamp: 1722838400634,
				},
			},
		},
		{
			Labels: []*prompb.Label{
				{
					Name:  "__name__",
					Value: "test_histogram_count",
				},
				{
					Name:  "label1",
					Value: "value1",
				},
				{
					Name:  "label2",
					Value: "value3",
				},
			},
			Samples: []*prompb.Sample{
				{
					Value:     88,
					Timestamp: 1722838400634,
				},
			},
		},
	}

	// sort want and got
	sort.Slice(got, func(i, j int) bool {
		return got[i].Labels[0].Value < got[j].Labels[0].Value
	})
	sort.Slice(got, func(i, j int) bool {
		return got[i].Samples[0].Value < got[j].Samples[0].Value
	})

	sort.Slice(want, func(i, j int) bool {
		return want[i].Labels[0].Value < want[j].Labels[0].Value
	})
	sort.Slice(want, func(i, j int) bool {
		return want[i].Samples[0].Value < want[j].Samples[0].Value
	})

	for _, ts := range got {
		TSDebug(ts)
	}
	log.Println("=====================================")
	for _, ts := range want {
		TSDebug(ts)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("PBHistogram.TimeSeries() = %v, want %v", got, want)
	}
}

func TSDebug(ts *prompb.TimeSeries) {
	var (
		metricName string
		kv         [][]string
		value      float64
		labels     string // {...}
		timestamp  int64
	)

	for _, lv := range ts.Labels {
		if lv.Name == "__name__" {
			metricName = lv.Value
			continue
		}
		kv = append(kv, []string{lv.Name, lv.Value})
	}
	if len(ts.Samples) == 1 {
		value = ts.Samples[0].Value
		timestamp = ts.Samples[0].Timestamp
	}
	for i, lv := range kv {
		if i == len(kv)-1 {
			labels += lv[0] + "=" + lv[1]
		} else {
			labels += lv[0] + "=" + lv[1] + ","
		}
	}
	log.Printf("%s{%s} @%d %f", metricName, labels, timestamp, value)
}
