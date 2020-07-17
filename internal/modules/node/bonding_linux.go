// +build !nobonding

package node

import (
	"errors"
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type bondingCollector struct {
	slaves, active pusher.TypedDesc
	logger         log.Logger
}

func init() {
	registerCollector("bonding", NewBondingCollector)
}

// NewBondingCollector returns a newly allocated bondingCollector.
// It exposes the number of configured and active slave of linux bonding interfaces.
func NewBondingCollector(logger log.Logger) (pusher.Collector, error) {
	return &bondingCollector{
		slaves: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bonding", "slaves"),
			"Number of configured slaves per bonding interface.",
			[]string{"master"}, nil,
		), prometheus.GaugeValue},
		active: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "bonding", "active"),
			"Number of active slaves per bonding interface.",
			[]string{"master"}, nil,
		), prometheus.GaugeValue},
		logger: logger,
	}, nil
}

// Update reads and exposes bonding states, implements Collector interface. Caution: This works only on linux.
func (c *bondingCollector) Update(ch chan<- prometheus.Metric) error {
	statusfile := sysFilePath("class/net")
	bondingStats, err := readBondingStats(statusfile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.logger.Debugf("msg", "Not collecting bonding, file does not exist", "file", statusfile)
			return pusher.ErrNoData
		}
		return err
	}
	for master, status := range bondingStats {
		ch <- c.slaves.MustNewConstMetric(float64(status[0]), master)
		ch <- c.active.MustNewConstMetric(float64(status[1]), master)
	}
	return nil
}

func readBondingStats(root string) (status map[string][2]int, err error) {
	status = map[string][2]int{}
	masters, err := ioutil.ReadFile(filepath.Join(root, "bonding_masters"))
	if err != nil {
		return nil, err
	}
	for _, master := range strings.Fields(string(masters)) {
		slaves, err := ioutil.ReadFile(filepath.Join(root, master, "bonding", "slaves"))
		if err != nil {
			return nil, err
		}
		sstat := [2]int{0, 0}
		for _, slave := range strings.Fields(string(slaves)) {
			state, err := ioutil.ReadFile(filepath.Join(root, master, fmt.Sprintf("lower_%s", slave), "bonding_slave", "mii_status"))
			if errors.Is(err, os.ErrNotExist) {
				// some older? kernels use slave_ prefix
				state, err = ioutil.ReadFile(filepath.Join(root, master, fmt.Sprintf("slave_%s", slave), "bonding_slave", "mii_status"))
			}
			if err != nil {
				return nil, err
			}
			sstat[0]++
			if strings.TrimSpace(string(state)) == "up" {
				sstat[1]++
			}
		}
		status[master] = sstat
	}
	return status, err
}
