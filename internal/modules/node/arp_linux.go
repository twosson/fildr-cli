package node

import (
	"fildr-cli/internal/log"
	"github.com/prometheus/client_golang/prometheus"
)

type arpCollector struct {
	entries *prometheus.Desc
	logger  log.Logger
}
