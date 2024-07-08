package metric

import "github.com/sq325/remoteWrite/prompb"

type PBMetric interface {
	TimeSeries(timestamp int64) []*prompb.TimeSeries
}
