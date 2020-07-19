package gateway

import "bytes"

type MetricData struct {
	instance string
	job      string
	data     *bytes.Buffer
}
