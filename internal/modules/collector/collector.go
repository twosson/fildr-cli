package collector

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

var (
	factories = map[string]map[string]func() (Collector, error){
		"node":   {},
		"gpu":    {},
		"daemon": {},
		"miner":  {},
		"worker": {},
	}
)

func registerCollector(namespace string, collector string, factory func() (Collector, error)) {
	factories[namespace][collector] = factory
}

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

type FilCollector struct {
	Collectors map[string]Collector

	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
}

func NewFilCollector(namespace string) *FilCollector {
	collectors := make(map[string]Collector)

	for key, f := range factories[namespace] {
		collector, err := f()
		if err != nil {
			return nil
		}
		collectors[key] = collector
	}

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

	return &FilCollector{Collectors: collectors, scrapeDurationDesc: scrapeDurationDesc, scrapeSuccessDesc: scrapeSuccessDesc}
}

func (f *FilCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- f.scrapeDurationDesc
	ch <- f.scrapeSuccessDesc
}

func (f *FilCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(f.Collectors))

	for name, c := range f.Collectors {
		go func(name string, c Collector) {
			f.execute(name, c, ch)
			wg.Done()
		}(name, c)
	}

	wg.Wait()
}

func (f *FilCollector) execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			//level.Debug(logger).Log("msg", "collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			//level.Error(logger).Log("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		success = 1
	}

	ch <- prometheus.MustNewConstMetric(f.scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(f.scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}
