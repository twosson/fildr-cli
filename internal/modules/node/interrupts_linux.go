// +build linux openbsd
// +build !nointerrupts

package node

import (
	"bufio"
	"errors"
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	interruptLabelNames = []string{"cpu", "type", "info", "devices"}
)

type interruptsCollector struct {
	desc   gateway.TypedDesc
	logger log.Logger
}

func init() {
	//registerCollector("interrupts", NewInterruptsCollector)
}

// NewInterruptsCollector returns a new Collector exposing interrupts stats.
func NewInterruptsCollector(logger log.Logger) (gateway.Collector, error) {
	return &interruptsCollector{
		desc: gateway.TypedDesc{prometheus.NewDesc(
			namespace+"_interrupts_total",
			"Interrupt details.",
			interruptLabelNames, nil,
		), prometheus.CounterValue},
		logger: logger,
	}, nil
}

func (c *interruptsCollector) Update(ch chan<- prometheus.Metric) (err error) {
	interrupts, err := getInterrupts()
	if err != nil {
		return fmt.Errorf("couldn't get interrupts: %w", err)
	}
	for name, interrupt := range interrupts {
		for cpuNo, value := range interrupt.values {
			fv, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in interrupts: %w", value, err)
			}
			ch <- c.desc.MustNewConstMetric(fv, strconv.Itoa(cpuNo), name, interrupt.info, interrupt.devices)
		}
	}
	return err
}

type interrupt struct {
	info    string
	devices string
	values  []string
}

func getInterrupts() (map[string]interrupt, error) {
	file, err := os.Open(procFilePath("interrupts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseInterrupts(file)
}

func parseInterrupts(r io.Reader) (map[string]interrupt, error) {
	var (
		interrupts = map[string]interrupt{}
		scanner    = bufio.NewScanner(r)
	)

	if !scanner.Scan() {
		return nil, errors.New("interrupts empty")
	}
	cpuNum := len(strings.Fields(scanner.Text())) // one header per cpu

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < cpuNum+2 { // irq + one column per cpu + details,
			continue // we ignore ERR and MIS for now
		}
		intName := parts[0][:len(parts[0])-1] // remove trailing :
		intr := interrupt{
			values: parts[1 : cpuNum+1],
		}

		if _, err := strconv.Atoi(intName); err == nil { // numeral interrupt
			intr.info = parts[cpuNum+1]
			intr.devices = strings.Join(parts[cpuNum+2:], " ")
		} else {
			intr.info = strings.Join(parts[cpuNum+1:], " ")
		}
		interrupts[intName] = intr
	}

	return interrupts, scanner.Err()
}
