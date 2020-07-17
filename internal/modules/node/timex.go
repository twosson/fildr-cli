// +build linux
// +build !notimex

package node

import (
	"fildr-cli/internal/log"
	"fildr-cli/internal/pusher"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sys/unix"
)

const (
	// The system clock is not synchronized to a reliable
	// server (TIME_ERROR).
	timeError = 5
	// The timex.Status time resolution bit (STA_NANO),
	// 0 = microsecond, 1 = nanoseconds.
	staNano = 0x2000

	// 1 second in
	nanoSeconds  = 1000000000
	microSeconds = 1000000
)

type timexCollector struct {
	offset,
	freq,
	maxerror,
	esterror,
	status,
	constant,
	tick,
	ppsfreq,
	jitter,
	shift,
	stabil,
	jitcnt,
	calcnt,
	errcnt,
	stbcnt,
	tai,
	syncStatus pusher.TypedDesc
	logger log.Logger
}

func init() {
	registerCollector("timex", NewTimexCollector)
}

// NewTimexCollector returns a new Collector exposing adjtime(3) stats.
func NewTimexCollector(logger log.Logger) (pusher.Collector, error) {
	const subsystem = "timex"

	return &timexCollector{
		offset: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "offset_seconds"),
			"Time offset in between local system and reference clock.",
			nil, nil,
		), prometheus.GaugeValue},
		freq: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "frequency_adjustment_ratio"),
			"Local clock frequency adjustment.",
			nil, nil,
		), prometheus.GaugeValue},
		maxerror: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "maxerror_seconds"),
			"Maximum error in seconds.",
			nil, nil,
		), prometheus.GaugeValue},
		esterror: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "estimated_error_seconds"),
			"Estimated error in seconds.",
			nil, nil,
		), prometheus.GaugeValue},
		status: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "status"),
			"Value of the status array bits.",
			nil, nil,
		), prometheus.GaugeValue},
		constant: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "loop_time_constant"),
			"Phase-locked loop time constant.",
			nil, nil,
		), prometheus.GaugeValue},
		tick: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "tick_seconds"),
			"Seconds between clock ticks.",
			nil, nil,
		), prometheus.GaugeValue},
		ppsfreq: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_frequency_hertz"),
			"Pulse per second frequency.",
			nil, nil,
		), prometheus.GaugeValue},
		jitter: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_jitter_seconds"),
			"Pulse per second jitter.",
			nil, nil,
		), prometheus.GaugeValue},
		shift: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_shift_seconds"),
			"Pulse per second interval duration.",
			nil, nil,
		), prometheus.GaugeValue},
		stabil: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_stability_hertz"),
			"Pulse per second stability, average of recent frequency changes.",
			nil, nil,
		), prometheus.GaugeValue},
		jitcnt: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_jitter_total"),
			"Pulse per second count of jitter limit exceeded events.",
			nil, nil,
		), prometheus.CounterValue},
		calcnt: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_calibration_total"),
			"Pulse per second count of calibration intervals.",
			nil, nil,
		), prometheus.CounterValue},
		errcnt: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_error_total"),
			"Pulse per second count of calibration errors.",
			nil, nil,
		), prometheus.CounterValue},
		stbcnt: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "pps_stability_exceeded_total"),
			"Pulse per second count of stability limit exceeded events.",
			nil, nil,
		), prometheus.CounterValue},
		tai: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "tai_offset_seconds"),
			"International Atomic Time (TAI) offset.",
			nil, nil,
		), prometheus.GaugeValue},
		syncStatus: pusher.TypedDesc{prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "sync_status"),
			"Is clock synchronized to a reliable server (1 = yes, 0 = no).",
			nil, nil,
		), prometheus.GaugeValue},
		logger: logger,
	}, nil
}

func (c *timexCollector) Update(ch chan<- prometheus.Metric) error {
	var syncStatus float64
	var divisor float64
	var timex = new(unix.Timex)

	status, err := unix.Adjtimex(timex)
	if err != nil {
		return fmt.Errorf("failed to retrieve adjtimex stats: %w", err)
	}

	if status == timeError {
		syncStatus = 0
	} else {
		syncStatus = 1
	}
	if (timex.Status & staNano) != 0 {
		divisor = nanoSeconds
	} else {
		divisor = microSeconds
	}
	// See NOTES in adjtimex(2).
	const ppm16frac = 1000000.0 * 65536.0

	ch <- c.syncStatus.MustNewConstMetric(syncStatus)
	ch <- c.offset.MustNewConstMetric(float64(timex.Offset) / divisor)
	ch <- c.freq.MustNewConstMetric(1 + float64(timex.Freq)/ppm16frac)
	ch <- c.maxerror.MustNewConstMetric(float64(timex.Maxerror) / microSeconds)
	ch <- c.esterror.MustNewConstMetric(float64(timex.Esterror) / microSeconds)
	ch <- c.status.MustNewConstMetric(float64(timex.Status))
	ch <- c.constant.MustNewConstMetric(float64(timex.Constant))
	ch <- c.tick.MustNewConstMetric(float64(timex.Tick) / microSeconds)
	ch <- c.ppsfreq.MustNewConstMetric(float64(timex.Ppsfreq) / ppm16frac)
	ch <- c.jitter.MustNewConstMetric(float64(timex.Jitter) / divisor)
	ch <- c.shift.MustNewConstMetric(float64(timex.Shift))
	ch <- c.stabil.MustNewConstMetric(float64(timex.Stabil) / ppm16frac)
	ch <- c.jitcnt.MustNewConstMetric(float64(timex.Jitcnt))
	ch <- c.calcnt.MustNewConstMetric(float64(timex.Calcnt))
	ch <- c.errcnt.MustNewConstMetric(float64(timex.Errcnt))
	ch <- c.stbcnt.MustNewConstMetric(float64(timex.Stbcnt))
	ch <- c.tai.MustNewConstMetric(float64(timex.Tai))

	return nil
}
