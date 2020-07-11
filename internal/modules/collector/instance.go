package collector

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Instance struct {
	R *prometheus.Registry
	C *FilCollector

	job      string
	instance string
}

func GetInstance(namespace string) *Instance {
	instance := &Instance{
		R:        prometheus.NewRegistry(),
		C:        NewFilCollector(namespace),
		job:      "defaultJobName",
		instance: "defaultInstanceName",
	}
	if instance.R == nil || instance.C == nil || instance.R.Register(instance.C) != nil {
		instance = nil
	}
	return instance
}

func (i *Instance) GetMetrics() string {
	if i == nil {
		return ""
	}

	mfs, err := i.R.Gather()
	if err != nil {
		fmt.Println("gather err = ", err)
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)

	for _, mf := range mfs {
		for _, m := range mf.GetMetric() {
			for _, l := range m.GetLabel() {
				if l.GetName() == "job" {
					fmt.Println("metric", mf.GetName(), m, "already contains a job label")
				}
			}
		}

		if err := enc.Encode(mf); err != nil {
			fmt.Println("encode error", err)
		}
	}
	return buf.String()
}

func (i *Instance) PushMetrics(gateway string, data string) error {
	sr := strings.NewReader(data)
	br := bufio.NewReader(sr)
	var url string
	if gateway[len(gateway)-1] == '/' {
		url = gateway + "metrics/job/" + i.job + "/instance/" + i.instance
	} else {
		url = gateway + "/metrics/job/" + i.job + "/instance/" + i.instance
	}

	req, err := http.NewRequest(http.MethodPost, url, br)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(resp.Body)
		errStr := fmt.Sprintf("unexpected status code %s, PushGateway url = %s, body = %s.", resp.StatusCode, url, string(body))
		return errors.New(errStr)
	}
	return nil
}

func (i *Instance) SetJob(job string) {
	i.job = job
}

func (i *Instance) GetJob() string {
	return i.job
}

func (i *Instance) SetInstance(instance string) {
	i.instance = instance
}

func (i *Instance) GetInstance() string {
	return i.instance
}
