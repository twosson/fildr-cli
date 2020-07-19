// +build !nontp

package node

import (
	"fildr-cli/internal/gateway"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/beevik/ntp"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"sync"
	"time"
)

const (
	hour24       = 24 * time.Hour // `time` does not export `Day` as Day != 24h because of DST
	ntpSubsystem = "ntp"
)

var (
	ntpServer          = "127.0.0.1"
	ntpProtocolVersion = 4
	ntpServerIsLocal   = false
	ntpIPTTL           = 1
	// 3.46608s ~ 1.5s + PHI * (1 << maxPoll), where 1.5s is MAXDIST from ntp.org, it is 1.0 in RFC5905
	// max-distance option is used as-is without phi*(1<<poll)
	ntpMaxDistance, _     = time.ParseDuration("3.46608s")
	ntpOffsetTolerance, _ = time.ParseDuration("1ms")

	leapMidnight      time.Time
	leapMidnightMutex = &sync.Mutex{}
)

type ntpCollector struct {
	stratum, leap, rtt, offset, reftime, rootDelay, rootDispersion, sanity gateway.TypedDesc
	logger                                                                 log.Logger
}

func init() {
	//registerCollector("ntp", NewNtpCollector)
}

// NewNtpCollector returns a new Collector exposing sanity of local NTP server.
// Default definition of "local" is:
// - collector.ntp.server address is a loopback address (or collector.ntp.server-is-mine flag is turned on)
// - the server is reachable with outgoin IP_TTL = 1
func NewNtpCollector(logger log.Logger) (gateway.Collector, error) {
	ipaddr := net.ParseIP(ntpServer)
	if !ntpServerIsLocal && (ipaddr == nil || !ipaddr.IsLoopback()) {
		return nil, fmt.Errorf("only IP address of local NTP server is valid for --collector.ntp.server")
	}

	if ntpProtocolVersion < 2 || ntpProtocolVersion > 4 {
		return nil, fmt.Errorf("invalid NTP protocol version %d; must be 2, 3, or 4", ntpProtocolVersion)
	}

	if ntpOffsetTolerance < 0 {
		return nil, fmt.Errorf("offset tolerance must be non-negative")
	}

	return &ntpCollector{
		stratum: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "stratum"),
			"NTPD stratum.",
			nil, nil,
		), prometheus.GaugeValue},
		leap: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "leap"),
			"NTPD leap second indicator, 2 bits.",
			nil, nil,
		), prometheus.GaugeValue},
		rtt: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "rtt_seconds"),
			"RTT to NTPD.",
			nil, nil,
		), prometheus.GaugeValue},
		offset: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "offset_seconds"),
			"ClockOffset between NTP and local clock.",
			nil, nil,
		), prometheus.GaugeValue},
		reftime: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "reference_timestamp_seconds"),
			"NTPD ReferenceTime, UNIX timestamp.",
			nil, nil,
		), prometheus.GaugeValue},
		rootDelay: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "root_delay_seconds"),
			"NTPD RootDelay.",
			nil, nil,
		), prometheus.GaugeValue},
		rootDispersion: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "root_dispersion_seconds"),
			"NTPD RootDispersion.",
			nil, nil,
		), prometheus.GaugeValue},
		sanity: gateway.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, ntpSubsystem, "sanity"),
			"NTPD sanity according to RFC5905 heuristics and configured limits.",
			nil, nil,
		), prometheus.GaugeValue},
		logger: logger,
	}, nil
}

func (c *ntpCollector) Update(ch chan<- prometheus.Metric) error {
	resp, err := ntp.QueryWithOptions(ntpServer, ntp.QueryOptions{
		Version: ntpProtocolVersion,
		TTL:     ntpIPTTL,
		Timeout: time.Second, // default `ntpdate` timeout
	})
	if err != nil {
		return fmt.Errorf("couldn't get SNTP reply: %w", err)
	}

	ch <- c.stratum.MustNewConstMetric(float64(resp.Stratum))
	ch <- c.leap.MustNewConstMetric(float64(resp.Leap))
	ch <- c.rtt.MustNewConstMetric(resp.RTT.Seconds())
	ch <- c.offset.MustNewConstMetric(resp.ClockOffset.Seconds())
	if resp.ReferenceTime.Unix() > 0 {
		// Go Zero is   0001-01-01 00:00:00 UTC
		// NTP Zero is  1900-01-01 00:00:00 UTC
		// UNIX Zero is 1970-01-01 00:00:00 UTC
		// so let's keep ALL ancient `reftime` values as zero
		ch <- c.reftime.MustNewConstMetric(float64(resp.ReferenceTime.UnixNano()) / 1e9)
	} else {
		ch <- c.reftime.MustNewConstMetric(0)
	}
	ch <- c.rootDelay.MustNewConstMetric(resp.RootDelay.Seconds())
	ch <- c.rootDispersion.MustNewConstMetric(resp.RootDispersion.Seconds())

	// Here is SNTP packet sanity check that is exposed to move burden of
	// configuration from node_exporter user to the developer.

	maxerr := ntpOffsetTolerance
	leapMidnightMutex.Lock()
	if resp.Leap == ntp.LeapAddSecond || resp.Leap == ntp.LeapDelSecond {
		// state of leapMidnight is cached as leap flag is dropped right after midnight
		leapMidnight = resp.Time.Truncate(hour24).Add(hour24)
	}
	if leapMidnight.Add(-hour24).Before(resp.Time) && resp.Time.Before(leapMidnight.Add(hour24)) {
		// tolerate leap smearing
		maxerr += time.Second
	}
	leapMidnightMutex.Unlock()

	if resp.Validate() == nil && resp.RootDistance <= ntpMaxDistance && resp.MinError <= maxerr {
		ch <- c.sanity.MustNewConstMetric(1)
	} else {
		ch <- c.sanity.MustNewConstMetric(0)
	}

	return nil
}
