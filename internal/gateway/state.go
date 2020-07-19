package gateway

import "bytes"

var metricQueue = NewQueue()

type MetricData struct {
	instance string
	job      string
	data     *bytes.Buffer
}

func push(data *MetricData) {
	metricQueue.Push(data)
}

func pop() *MetricData {
	if metricQueue.IsEmpty() {
		return nil
	}
	return metricQueue.Pop().(*MetricData)
}
