package gateway

import (
	"bytes"
	"context"
	"errors"
	"fildr-cli/internal/config"
	"fildr-cli/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"sync"
	"time"
)

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

var registries = make(map[string]*prometheus.Registry)
var pcs = make(map[string]*promCollector)

// 注册收集器
func Registry(namespace string, name string, c Collector) {
	pc, ok := pcs[namespace]
	if !ok {
		pc = newPromCollector(namespace)
		pcs[namespace] = pc
		nr := prometheus.NewRegistry()
		nr.Register(pc)
		registries[namespace] = nr
	}
	pc.collectors[name] = c
}

type promCollector struct {
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
	instance           string
	job                string
	registry           *prometheus.Registry
	collectors         map[string]Collector
	logger             log.Logger
}

func newPromCollector(namespace string) *promCollector {
	logger := log.From(context.Background())
	scrapeDurationDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		namespace+"_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		namespace+"_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
	collectors := make(map[string]Collector)

	cfg := config.Get()

	return &promCollector{
		scrapeDurationDesc: scrapeDurationDesc,
		scrapeSuccessDesc:  scrapeSuccessDesc,
		collectors:         collectors,
		logger:             logger,
		job:                namespace,
		instance:           cfg.Gateway.Instance,
	}
}

func (p *promCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.scrapeDurationDesc
	ch <- p.scrapeSuccessDesc
}

func (p *promCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(p.collectors))
	for name, c := range p.collectors {
		go func(name string, c Collector) {

			begin := time.Now()
			err := c.Update(ch)
			duration := time.Since(begin)
			var success float64

			if err != nil {
				if IsNoDataError(err) {
					p.logger.Debugf("msg collector returned no data: name: %s, duration_seconds: %s, err: %v", name, duration.Seconds(), err)
				} else {
					p.logger.Debugf("msg collector failed: name: %s, duration_seconds: %s, err: %v", name, duration.Seconds(), err)
				}
				success = 0
			} else {
				success = 1
			}

			ch <- prometheus.MustNewConstMetric(p.scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
			ch <- prometheus.MustNewConstMetric(p.scrapeSuccessDesc, prometheus.GaugeValue, success, name)

			wg.Done()
		}(name, c)
	}

	wg.Wait()
}

func getMetrics() ([]*MetricData, error) {
	datas := make([]*MetricData, 0)
	for k, v := range registries {
		mfs, err := v.Gather()
		if err != nil {
			return nil, err
		}
		buf := &bytes.Buffer{}
		enc := expfmt.NewEncoder(buf, expfmt.FmtText)
		for _, mf := range mfs {
			if err := enc.Encode(mf); err != nil {
				return nil, err
			}
		}
		pc := pcs[k]
		datas = append(datas, &MetricData{instance: pc.instance, job: pc.job, data: buf})
	}
	return datas, nil
}

var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

type TypedDesc struct {
	Desc      *prometheus.Desc
	ValueType prometheus.ValueType
}

func (d *TypedDesc) MustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.Desc, d.ValueType, value, labels...)
}
