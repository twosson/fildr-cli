package pusher

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

type FildrCollector struct {
	Collectors         map[string]Collector
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
	logger             log.Logger
}

func NewFildrCollector(ctx context.Context, namespace string) (*FildrCollector, error) {
	fc := &FildrCollector{
		Collectors: make(map[string]Collector),
		logger:     log.From(ctx),
	}

	fc.scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		namespace+"_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)

	fc.scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		namespace+"_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)

	return fc, nil
}

func (fc *FildrCollector) Registry(name string, c Collector) {
	fc.Collectors[name] = c
}

func (fc *FildrCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- fc.scrapeDurationDesc
	ch <- fc.scrapeSuccessDesc
}

var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

func (fc *FildrCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(fc.Collectors))
	for name, c := range fc.Collectors {
		go func(name string, c Collector) {

			begin := time.Now()
			err := c.Update(ch)
			duration := time.Since(begin)
			var success float64

			if err != nil {
				if IsNoDataError(err) {
					fc.logger.Debugf("msg collector returned no data: name: %s, duration_seconds: %s, err: %v", name, duration.Seconds(), err)
				} else {
					fc.logger.Debugf("msg collector failed: name: %s, duration_seconds: %s, err: %v", name, duration.Seconds(), err)
				}
				success = 0
			} else {
				success = 1
			}

			ch <- prometheus.MustNewConstMetric(fc.scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(fc.scrapeSuccessDesc, prometheus.GaugeValue, success, name)

			wg.Done()
		}(name, c)
	}

	wg.Wait()
}

type PromInstance struct {
	R *prometheus.Registry
	C *FildrCollector

	gateway  string
	token    string
	job      string
	instance string
	logger   log.Logger
}

func GetPromInstance(ctx context.Context, job string, fc *FildrCollector) (*PromInstance, error) {

	promInstance := &PromInstance{
		R:      prometheus.NewRegistry(),
		C:      fc,
		job:    job,
		logger: log.From(ctx),
	}

	cfg := config.Get()

	if len(cfg.Gateway.Url) == 0 {
		return nil, fmt.Errorf("toml config gateway url is nil")
	}

	promInstance.gateway = cfg.Gateway.Url

	if len(cfg.Gateway.Token) == 0 {
		return nil, fmt.Errorf("toml config gateway token is nil")
	}

	promInstance.token = cfg.Gateway.Token

	if len(cfg.Gateway.Instance) == 0 {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		promInstance.instance = hostname
	} else {
		promInstance.instance = cfg.Gateway.Instance
	}

	if promInstance.R == nil {
		return nil, fmt.Errorf("prom new registry err")
	}

	if promInstance.C == nil {
		return nil, fmt.Errorf("prom collector is nil")
	}

	if len(job) == 0 {
		return nil, fmt.Errorf("prom job is nil")
	}

	if err := promInstance.R.Register(promInstance.C); err != nil {
		return nil, err
	}

	return promInstance, nil
}

func (i *PromInstance) GetMetrics() (string, error) {
	mfs, err := i.R.Gather()
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)

	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			return "", nil
		}
	}
	return buf.String(), nil
}

func (i *PromInstance) PushMetrics(data string) error {
	sr := strings.NewReader(data)
	br := bufio.NewReader(sr)
	var url string
	if i.gateway[len(i.gateway)-1] == '/' {
		url = i.gateway + "metrics/job/" + i.job + "/instance/" + i.instance
	} else {
		url = i.gateway + "/metrics/job/" + i.job + "/instance/" + i.instance
	}

	req, err := http.NewRequest(http.MethodPost, url, br)
	if err != nil {
		return err
	}
	req.Header.Add("blade-auth", "Bearer "+i.token)
	req.Header.Add("Content-Type", "text/plain")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		errStr := fmt.Sprintf("unexpected status code %d, PushGateway url = %s, body = %s.", resp.StatusCode, url, string(body))
		return errors.New(errStr)
	}
	return nil
}

type TypedDesc struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

func (d *TypedDesc) MustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.Desc, d.ValueType, value, labels...)
}
