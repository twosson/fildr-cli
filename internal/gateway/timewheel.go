package gateway

import (
	"context"
	"fildr-cli/internal/log"
	"github.com/rfyiamcool/go-timewheel"
	"time"
)

var tws *timewheel.TimeWheel
var logger log.Logger

func init() {
	tw, err := timewheel.NewTimeWheel(1*time.Second, 360)
	if err != nil {
		panic(err)
	}
	tws = tw
}

func Run(ctx context.Context) error {
	logger = log.From(ctx)
	tws, err := timewheel.NewTimeWheel(1*time.Second, 360)
	if err != nil {
		return err
	}
	tws.AddCron(time.Second*5, func() {
		datas, err := getMetrics()
		if err != nil {
			return
		}
		for i := range datas {
			postGateway(datas[i])
		}
	})

	tws.Start()
	return nil
}
