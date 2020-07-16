package node

import (
	"bufio"
	"bytes"
	"errors"
	"fildr-cli/internal/log"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Instance struct {
	R *prometheus.Registry
	C *NodeCollector

	job      string
	instance string
	logger   log.Logger
}

func GetInstance(logger log.Logger) (*Instance, error) {
	node, err := NewNodeCollector(logger)
	if err != nil {
		return nil, err
	}
	instance := &Instance{
		R:        prometheus.NewRegistry(),
		C:        node,
		job:      "defaultJobName",
		instance: "defaultInstanceName",
		logger:   logger,
	}
	if instance.R == nil || instance.C == nil || instance.R.Register(instance.C) != nil {
		return nil, err
	}
	return instance, nil
}

func (i *Instance) GetMetrics() (string, error) {
	if i == nil {
		return "", nil
	}

	mfs, err := i.R.Gather()
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)

	for _, mf := range mfs {
		if err := enc.Encode(mf); err != nil {
			return "", nil
		}
	}
	return buf.String(), nil
}

func (i *Instance) PushMetrics(gateway string, token string, data string) error {
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
	req.Header.Add("blade-auth", "Bearer "+token)
	req.Header.Add("Content-Type", "text/plain")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		errStr := fmt.Sprintf("unexpected status code %d, PushGateway url = %s, body = %s.", resp.StatusCode, url, string(body))
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
