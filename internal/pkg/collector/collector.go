package collector

import "github.com/prometheus/client_golang/prometheus"

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}
