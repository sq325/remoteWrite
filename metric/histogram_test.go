package metric

import (
	"log"
	"reflect"
	"testing"

	"github.com/sq325/remoteWrite/prompb"
)

func TestPBHistogram_TimeSeries(t *testing.T) {

	type args struct {
		name      string
		help      string
		labels    []string
		buckets   []float64
		timestamp int64
	}
	tests := []struct {
		name  string
		args  args
		want  []*prompb.TimeSeries
		want2 []*prompb.TimeSeries
	}{
		{
			name: "test1",
			args: args{
				name:      "app1",
				help:      "help1",
				labels:    []string{"l1", "l2"},
				buckets:   []float64{50, 100, 200},
				timestamp: 1722838400634,
			},
			want: []*prompb.TimeSeries{
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
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
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
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
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
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
							Value: "app1_sum",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
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
							Value: "app1_count",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
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
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
						{
							Name:  "le",
							Value: "+Inf",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     13,
							Timestamp: 1722838400634,
						},
					},
				},
			},
			want2: []*prompb.TimeSeries{
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
						{
							Name:  "le",
							Value: "50",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     10,
							Timestamp: 1722838430634,
						},
					},
				},
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
						{
							Name:  "le",
							Value: "100",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     16,
							Timestamp: 1722838430634,
						},
					},
				},
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
						{
							Name:  "le",
							Value: "200",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     26,
							Timestamp: 1722838430634,
						},
					},
				},
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_sum",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     3000,
							Timestamp: 1722838430634,
						},
					},
				},
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_count",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     26,
							Timestamp: 1722838430634,
						},
					},
				},
				{
					Labels: []*prompb.Label{
						{
							Name:  "__name__",
							Value: "app1_bucket",
						},
						{
							Name:  "l1",
							Value: "v1",
						},
						{
							Name:  "l2",
							Value: "v2",
						},
						{
							Name:  "le",
							Value: "+Inf",
						},
					},
					Samples: []*prompb.Sample{
						{
							Value:     26,
							Timestamp: 1722838430634,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hg := NewPBHistogram(tt.args.name, tt.args.help, tt.args.labels, tt.args.buckets)
			hg.Add([]string{"v1", "v2"}, 5, 50)
			hg.Add([]string{"v1", "v2"}, 8, 100)
			hg.Add([]string{"v1", "v2"}, 13, 200)
			hg.AddCount(13)
			hg.AddSum(1500)
			got := hg.TimeSeries(tt.args.timestamp)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PBHistogram.TimeSeries() = %v, want %v", got, tt.want)
			}

			for _, ts := range got {
				TSDebug(ts)
			}

			hg.Add([]string{"v1", "v2"}, 5, 50)
			hg.Add([]string{"v1", "v2"}, 8, 100)
			hg.Add([]string{"v1", "v2"}, 13, 200)
			hg.AddCount(13)
			hg.AddSum(1500)
			got2 := hg.TimeSeries(tt.args.timestamp + 30000)

			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("PBHistogram.TimeSeries() = %v, want %v", got, tt.want)
			}
			for _, ts := range got2 {
				TSDebug(ts)
			}
		})
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
