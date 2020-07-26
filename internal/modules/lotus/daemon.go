package lotus

import (
	"context"
)

type daemon struct {
	id            string
	daemonVersion string
	apiVersion    string

	lotusApi    *LotusApi
	lotusClient *LotusClient
	isShutdown  bool
}

func newDaemon() (*daemon, error) {
	daemon := &daemon{}
	daemon.lotusApi = &LotusApi{}
	daemon.lotusClient = &LotusClient{}
	daemon.isShutdown = false
	if err := daemon.lotusClient.WithDaemonClient(daemon.lotusApi); err != nil {
		return nil, err
	}

	daemonVersion, apiVersion, err := daemon.version()
	if err != nil {
		return nil, err
	}

	netId, err := daemon.netId()
	if err != nil {
		return nil, err
	}

	daemon.id = netId
	daemon.daemonVersion = daemonVersion
	daemon.apiVersion = apiVersion

	cancelCtx, cancel := context.WithCancel(context.Background())
	closingCh, err := daemon.lotusApi.Closing(cancelCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	go func() {
		<-closingCh
		daemon.isShutdown = true
		cancel()
	}()

	return daemon, nil
}

func (d *daemon) netId() (string, error) {
	pid, err := d.lotusApi.ID()
	if err != nil {
		return "", err
	}
	return pid.String(), nil
}

func (d *daemon) version() (string, string, error) {
	v, err := d.lotusApi.Version()
	if err != nil {
		return "", "", err
	}
	return v.Version, v.APIVersion.String(), nil
}
