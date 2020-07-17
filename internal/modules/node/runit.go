// +build !norunit

package node

import (
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/soundcloud/go-runit/runit"
)

var runitServiceDir = "/etc/service"

type runitCollector struct {
	state          pusher.TypedDesc
	stateDesired   pusher.TypedDesc
	stateNormal    pusher.TypedDesc
	stateTimestamp pusher.TypedDesc
	logger         log.Logger
}

func init() {
	//registerCollector("runit", NewRunitCollector)
}

// NewRunitCollector returns a new Collector exposing runit statistics.
func NewRunitCollector(logger log.Logger) (pusher.Collector, error) {
	var (
		subsystem   = "service"
		constLabels = prometheus.Labels{"supervisor": "runit"}
		labelNames  = []string{"service"}
	)

	return &runitCollector{
		state: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "state"),
			"State of runit service.",
			labelNames, constLabels,
		), prometheus.GaugeValue},
		stateDesired: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "desired_state"),
			"Desired state of runit service.",
			labelNames, constLabels,
		), prometheus.GaugeValue},
		stateNormal: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "normal_state"),
			"Normal state of runit service.",
			labelNames, constLabels,
		), prometheus.GaugeValue},
		stateTimestamp: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "state_last_change_timestamp_seconds"),
			"Unix timestamp of the last runit service state change.",
			labelNames, constLabels,
		), prometheus.GaugeValue},
		logger: logger,
	}, nil
}

func (c *runitCollector) Update(ch chan<- prometheus.Metric) error {
	services, err := runit.GetServices(runitServiceDir)
	if err != nil {
		return err
	}

	for _, service := range services {
		status, err := service.Status()
		if err != nil {
			c.logger.Debugf("msg", "Couldn't get status", "service", service.Name, "err", err)
			continue
		}

		c.logger.Debugf("msg", "duration", "service", service.Name, "status", status.State, "pid", status.Pid, "duration_seconds", status.Duration)
		ch <- c.state.MustNewConstMetric(float64(status.State), service.Name)
		ch <- c.stateDesired.MustNewConstMetric(float64(status.Want), service.Name)
		ch <- c.stateTimestamp.MustNewConstMetric(float64(status.Timestamp.Unix()), service.Name)
		if status.NormallyUp {
			ch <- c.stateNormal.MustNewConstMetric(1, service.Name)
		} else {
			ch <- c.stateNormal.MustNewConstMetric(0, service.Name)
		}
	}
	return nil
}
