// +build darwin freebsd openbsd linux
// +build !nouname

package node

import (
	"bytes"
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/unix"
)

var unameDesc = prometheus.NewDesc(
	prometheus.BuildFQName(namespace, "uname", "info"),
	"Labeled system information as provided by the uname system call.",
	[]string{
		"sysname",
		"release",
		"version",
		"machine",
		"nodename",
		"domainname",
	},
	nil,
)

type unameCollector struct {
	logger log.Logger
}
type uname struct {
	SysName    string
	Release    string
	Version    string
	Machine    string
	NodeName   string
	DomainName string
}

func init() {
	registerCollector("uname", newUnameCollector)
}

// NewUnameCollector returns new unameCollector.
func newUnameCollector(logger log.Logger) (pusher.Collector, error) {
	return &unameCollector{logger}, nil
}

func (c *unameCollector) Update(ch chan<- prometheus.Metric) error {
	uname, err := getUname()
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(unameDesc, prometheus.GaugeValue, 1,
		uname.SysName,
		uname.Release,
		uname.Version,
		uname.Machine,
		uname.NodeName,
		uname.DomainName,
	)

	return nil
}

func getUname() (uname, error) {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return uname{}, err
	}

	output := uname{
		SysName:    string(utsname.Sysname[:bytes.IndexByte(utsname.Sysname[:], 0)]),
		Release:    string(utsname.Release[:bytes.IndexByte(utsname.Release[:], 0)]),
		Version:    string(utsname.Version[:bytes.IndexByte(utsname.Version[:], 0)]),
		Machine:    string(utsname.Machine[:bytes.IndexByte(utsname.Machine[:], 0)]),
		NodeName:   string(utsname.Nodename[:bytes.IndexByte(utsname.Nodename[:], 0)]),
		DomainName: string(utsname.Domainname[:bytes.IndexByte(utsname.Domainname[:], 0)]),
	}

	return output, nil
}
