// +build !noedac

package node

import (
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"path/filepath"
	"regexp"
)

const (
	edacSubsystem = "edac"
)

var (
	edacMemControllerRE = regexp.MustCompile(`.*devices/system/edac/mc/mc([0-9]*)`)
	edacMemCsrowRE      = regexp.MustCompile(`.*devices/system/edac/mc/mc[0-9]*/csrow([0-9]*)`)
)

type edacCollector struct {
	ceCount      *prometheus.Desc
	ueCount      *prometheus.Desc
	csRowCECount *prometheus.Desc
	csRowUECount *prometheus.Desc
	logger       log.Logger
}

func init() {
	registerCollector("edac", NewEdacCollector)
}

// NewEdacCollector returns a new Collector exposing edac stats.
func NewEdacCollector(logger log.Logger) (pusher.Collector, error) {
	return &edacCollector{
		ceCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, edacSubsystem, "correctable_errors_total"),
			"Total correctable memory errors.",
			[]string{"controller"}, nil,
		),
		ueCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, edacSubsystem, "uncorrectable_errors_total"),
			"Total uncorrectable memory errors.",
			[]string{"controller"}, nil,
		),
		csRowCECount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, edacSubsystem, "csrow_correctable_errors_total"),
			"Total correctable memory errors for this csrow.",
			[]string{"controller", "csrow"}, nil,
		),
		csRowUECount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, edacSubsystem, "csrow_uncorrectable_errors_total"),
			"Total uncorrectable memory errors for this csrow.",
			[]string{"controller", "csrow"}, nil,
		),
		logger: logger,
	}, nil
}

func (c *edacCollector) Update(ch chan<- prometheus.Metric) error {
	memControllers, err := filepath.Glob(sysFilePath("devices/system/edac/mc/mc[0-9]*"))
	if err != nil {
		return err
	}
	for _, controller := range memControllers {
		controllerMatch := edacMemControllerRE.FindStringSubmatch(controller)
		if controllerMatch == nil {
			return fmt.Errorf("controller string didn't match regexp: %s", controller)
		}
		controllerNumber := controllerMatch[1]

		value, err := readUintFromFile(filepath.Join(controller, "ce_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ce_count for controller %s: %w", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ceCount, prometheus.CounterValue, float64(value), controllerNumber)

		value, err = readUintFromFile(filepath.Join(controller, "ce_noinfo_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ce_noinfo_count for controller %s: %w", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.csRowCECount, prometheus.CounterValue, float64(value), controllerNumber, "unknown")

		value, err = readUintFromFile(filepath.Join(controller, "ue_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ue_count for controller %s: %w", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.ueCount, prometheus.CounterValue, float64(value), controllerNumber)

		value, err = readUintFromFile(filepath.Join(controller, "ue_noinfo_count"))
		if err != nil {
			return fmt.Errorf("couldn't get ue_noinfo_count for controller %s: %w", controllerNumber, err)
		}
		ch <- prometheus.MustNewConstMetric(
			c.csRowUECount, prometheus.CounterValue, float64(value), controllerNumber, "unknown")

		// For each controller, walk the csrow directories.
		csrows, err := filepath.Glob(controller + "/csrow[0-9]*")
		if err != nil {
			return err
		}
		for _, csrow := range csrows {
			csrowMatch := edacMemCsrowRE.FindStringSubmatch(csrow)
			if csrowMatch == nil {
				return fmt.Errorf("csrow string didn't match regexp: %s", csrow)
			}
			csrowNumber := csrowMatch[1]

			value, err = readUintFromFile(filepath.Join(csrow, "ce_count"))
			if err != nil {
				return fmt.Errorf("couldn't get ce_count for controller/csrow %s/%s: %w", controllerNumber, csrowNumber, err)
			}
			ch <- prometheus.MustNewConstMetric(
				c.csRowCECount, prometheus.CounterValue, float64(value), controllerNumber, csrowNumber)

			value, err = readUintFromFile(filepath.Join(csrow, "ue_count"))
			if err != nil {
				return fmt.Errorf("couldn't get ue_count for controller/csrow %s/%s: %w", controllerNumber, csrowNumber, err)
			}
			ch <- prometheus.MustNewConstMetric(
				c.csRowUECount, prometheus.CounterValue, float64(value), controllerNumber, csrowNumber)
		}
	}

	return err
}
