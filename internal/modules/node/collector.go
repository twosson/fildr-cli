package node

import (
	"errors"
	"fildr-cli/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

const namespace = "node"

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"node_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"node_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

var (
	factories      = make(map[string]func(logger log.Logger) (Collector, error))
	collectorState = make(map[string]bool)
)

func registerCollector(collector string, factory func(logger log.Logger) (Collector, error)) {
	factories[collector] = factory
	collectorState[collector] = true
}

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

type NodeCollector struct {
	Collectors map[string]Collector
	logger     log.Logger
}

func NewNodeCollector(logger log.Logger, filters ...string) (*NodeCollector, error) {

	for _, filter := range filters {
		_, exist := collectorState[filter]
		if exist {
			collectorState[filter] = false
		}
	}

	collectors := make(map[string]Collector)

	for key, enabled := range collectorState {
		if enabled {
			collector, err := factories[key](log.NopLogger().Named("collector-" + key))
			if err != nil {
				continue
			}
			collectors[key] = collector
		}
	}

	return &NodeCollector{Collectors: collectors, logger: logger}, nil
}

func (n NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

func (n NodeCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))

	for name, c := range n.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch, n.logger)
			wg.Done()
		}(name, c)
	}

	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric, logger log.Logger) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			logger.Debugf("msg collector returned no data: name: %s, duration_seconds: %s, err: %v", name, duration.Seconds(), err)
		} else {
			logger.Errorf("msg collector failed: name: %s, duration_seconds: %s, err: %v", name, duration.Seconds(), err)
		}
		success = 0
	} else {
		success = 1
	}

	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}

type typedDesc struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}

func (d *typedDesc) mustNewConstMetric(value float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(d.desc, d.valueType, value, labels...)
}

var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}
