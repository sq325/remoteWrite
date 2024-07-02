package client

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sq325/remoteWriteClient/prompb"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	url    string
	client *http.Client

	RequestCounter         *prometheus.CounterVec
	RequestBytesCounter    *prometheus.CounterVec
	WriteTimeSeriesCounter *prometheus.CounterVec
	flag                   *prometheus.GaugeVec
}

func NewClient(url string, opts ...Option) *Client {

	defaultOpt := []Option{
		WithDialTimeout(5 * time.Second),
		WithTimeout(15 * time.Second),
	}

	c := newConfig(append(defaultOpt, opts...)...)

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: c.DialTimeout,
		}).DialContext,
		ResponseHeaderTimeout: c.Timeout,
		MaxIdleConnsPerHost:   100,
	}
	httpclient := &http.Client{
		Transport: tr,
	}

	// flags
	f := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "remotewirte_client_flag",
			Help: "Flag of remote write client",
		},
		[]string{"name", "value"},
	)
	f.WithLabelValues("dialTimeout", c.DialTimeout.String()).Set(1)
	f.WithLabelValues("timeout", c.Timeout.String()).Set(1)
	f.WithLabelValues("url", url).Set(1)

	return &Client{
		url:    url,
		client: httpclient,
		RequestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "remotewrite_client_request_total",
				Help: "Total number of remote write requests sent to the remote storage",
			}, []string{"endpoint"},
		),
		RequestBytesCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "remotewrite_client_write_bytes_total",
				Help: "Total number of bytes sent to the remote storage after snappy compression",
			},
			[]string{"endpoint"},
		),
		WriteTimeSeriesCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "remotewrite_client_write_timeseries_total",
				Help: "Total number of time series sent to the remote storage",
			},
			[]string{"endpoint"},
		),
		flag: f,
	}
}

func (c *Client) Name() string {
	return "RemoteWrite Client"
}

func (c *Client) Write(series []*prompb.TimeSeries) error {
	if len(series) == 0 {
		return nil
	}

	req := &prompb.WriteRequest{
		Timeseries: series,
	}

	bys, err := proto.Marshal(req)
	if err != nil {
		log.Printf("failed to marshal WriteRequest: %v", err)
		return err
	}
	if err := c.write(snappy.Encode(nil, bys)); err != nil {
		return err
	}
	c.WriteTimeSeriesCounter.WithLabelValues(c.url).Add(float64(len(series)))
	return nil
}

func (c *Client) write(bys []byte) error {
	req, err := http.NewRequest("POST", c.url, bytes.NewReader(bys))
	if err != nil {
		log.Printf("failed to create request: %v", err)
		return err
	}
	req.Header.Add("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("User-Agent", "kube-eventer")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Println("push data with remote write request got error:", err, "response body:", string(bys))
		return err
	}
	// meter
	c.RequestCounter.WithLabelValues(c.url).Inc()
	c.RequestBytesCounter.WithLabelValues(c.url).Add(float64(len(bys)))
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("push data with remote write request got status code: %v, response body: %s", resp.StatusCode, string(bys))
		return err
	}

	return nil
}
