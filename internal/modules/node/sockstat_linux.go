// +build !nosockstat

package node

import (
	"errors"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"os"
)

const (
	sockStatSubsystem = "sockstat"
)

// Used for calculating the total memory bytes on TCP and UDP.
var pageSize = os.Getpagesize()

type sockStatCollector struct {
	logger log.Logger
}

func init() {
	registerCollector(sockStatSubsystem, NewSockStatCollector)
}

// NewSockStatCollector returns a new Collector exposing socket stats.
func NewSockStatCollector(logger log.Logger) (gateway.Collector, error) {
	return &sockStatCollector{logger}, nil
}

func (c *sockStatCollector) Update(ch chan<- prometheus.Metric) error {
	fs, err := procfs.NewFS(procPath)
	if err != nil {
		return fmt.Errorf("failed to open procfs: %w", err)
	}

	// If IPv4 and/or IPv6 are disabled on this kernel, handle it gracefully.
	stat4, err := fs.NetSockstat()
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		c.logger.Debugf("msg", "IPv4 sockstat statistics not found, skipping")
	default:
		return fmt.Errorf("failed to get IPv4 sockstat data: %w", err)
	}

	stat6, err := fs.NetSockstat6()
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		c.logger.Debugf("msg", "IPv6 sockstat statistics not found, skipping")
	default:
		return fmt.Errorf("failed to get IPv6 sockstat data: %w", err)
	}

	stats := []struct {
		isIPv6 bool
		stat   *procfs.NetSockstat
	}{
		{
			stat: stat4,
		},
		{
			isIPv6: true,
			stat:   stat6,
		},
	}

	for _, s := range stats {
		c.update(ch, s.isIPv6, s.stat)
	}

	return nil
}

func (c *sockStatCollector) update(ch chan<- prometheus.Metric, isIPv6 bool, s *procfs.NetSockstat) {
	if s == nil {
		// IPv6 disabled or similar; nothing to do.
		return
	}

	// If sockstat contains the number of used sockets, export it.
	if !isIPv6 && s.Used != nil {
		// TODO: this must be updated if sockstat6 ever exports this data.
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, sockStatSubsystem, "sockets_used"),
				"Number of IPv4 sockets in use.",
				nil,
				nil,
			),
			prometheus.GaugeValue,
			float64(*s.Used),
		)
	}

	// A name and optional value for a sockstat metric.
	type ssPair struct {
		name string
		v    *int
	}

	// Previously these metric names were generated directly from the file output.
	// In order to keep the same level of compatibility, we must map the fields
	// to their correct names.
	for _, p := range s.Protocols {
		pairs := []ssPair{
			{
				name: "inuse",
				v:    &p.InUse,
			},
			{
				name: "orphan",
				v:    p.Orphan,
			},
			{
				name: "tw",
				v:    p.TW,
			},
			{
				name: "alloc",
				v:    p.Alloc,
			},
			{
				name: "mem",
				v:    p.Mem,
			},
			{
				name: "memory",
				v:    p.Memory,
			},
		}

		// Also export mem_bytes values for sockets which have a mem value
		// stored in pages.
		if p.Mem != nil {
			v := *p.Mem * pageSize
			pairs = append(pairs, ssPair{
				name: "mem_bytes",
				v:    &v,
			})
		}

		for _, pair := range pairs {
			if pair.v == nil {
				// This value is not set for this protocol; nothing to do.
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(
						namespace,
						sockStatSubsystem,
						fmt.Sprintf("%s_%s", p.Protocol, pair.name),
					),
					fmt.Sprintf("Number of %s sockets in state %s.", p.Protocol, pair.name),
					nil,
					nil,
				),
				prometheus.GaugeValue,
				float64(*pair.v),
			)
		}
	}
}
