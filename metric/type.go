package metric

import "github.com/sq325/remoteWrite/prompb"

type PbMetric interface {
	TimeSeries() []*prompb.TimeSeries
}


