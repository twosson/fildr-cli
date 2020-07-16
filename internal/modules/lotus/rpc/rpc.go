package rpc

import (
	"context"
	"fildr-cli/internal/log"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/filecoin-project/specs-actors/actors/abi"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"golang.org/x/xerrors"
	"net/http"
	"time"
)

func WaitForSyncComplete(ctx context.Context, napi api.FullNode) error {

	logger := log.NopLogger().Named("lotus-sync")

sync_complete:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			state, err := napi.SyncState(ctx)
			if err != nil {
				return err
			}

			for i, w := range state.ActiveSyncs {
				if w.Target == nil {
					continue
				}

				if w.Stage == api.StageSyncErrored {
					logger.Errorf("")
				} else {
					logger.Errorf("Syncing workder %d", i)
				}

				if w.Stage == api.StageSyncComplete {
					break sync_complete
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			head, err := napi.ChainHead(ctx)
			if err != nil {
				return err
			}

			timestampDelta := uint64(time.Now().Unix() - int64(head.MinTimestamp()))

			logger.Infof("Waiting for reasonable head height: %d, timestamp_delta: %s", head.Height(), timestampDelta)

			if timestampDelta < build.BlockDelaySecs*20 {
				return nil
			}
		}
	}
}

func GetTips(ctx context.Context, api api.FullNode, lastHeight abi.ChainEpoch, headlag int) (<-chan *types.TipSet, error) {
	chmain := make(chan *types.TipSet)

	hb := NewHeadBuffer(headlag)

	notif, err := api.ChainNotify(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(chmain)

		ping := time.Tick(300 * time.Second)

		for {
			select {
			case changes := <-notif:
				for _, change := range changes {
					switch change.Type {
					case store.HCCurrent:
						tipsets, err := loadTipsets(ctx, api, change.Val, lastHeight)
						if err != nil {
							return
						}

						for _, tipset := range tipsets {
							chmain <- tipset
						}
					case store.HCApply:
						if out := hb.Push(change); out != nil {
							chmain <- out.Val
						}
					case store.HCRevert:
						hb.Pop()
					}
				}
			case <-ping:
				cctx, cancel := context.WithTimeout(ctx, 5*time.Second)

				if _, err := api.ID(cctx); err != nil {
					cancel()
					return
				}

				cancel()

			case <-ctx.Done():
				return
			}
		}
	}()

	return chmain, nil
}

func loadTipsets(ctx context.Context, api api.FullNode, curr *types.TipSet, lowestHeight abi.ChainEpoch) ([]*types.TipSet, error) {
	tipsets := []*types.TipSet{}

	for {
		if curr.Height() == 0 {
			break
		}

		if curr.Height() <= lowestHeight {
			break
		}

		tipsets = append(tipsets, curr)

		tsk := curr.Parents()
		prev, err := api.ChainGetTipSet(ctx, tsk)
		if err != nil {
			return tipsets, err
		}

		curr = prev
	}

	for i, j := 0, len(tipsets)-1; i < j; i, j = i+1, j-1 {
		tipsets[i], tipsets[j] = tipsets[j], tipsets[i]
	}

	return tipsets, nil
}

func GetFullNodeApiUsingCredentials(listenAddr, token string) (api.FullNode, jsonrpc.ClientCloser, error) {
	parsedAddr, err := ma.NewMultiaddr(listenAddr)
	if err != nil {
		return nil, nil, err
	}

	_, addr, err := manet.DialArgs(parsedAddr)
	if err != nil {
		return nil, nil, err
	}

	return client.NewFullNodeRPC(apiUri(addr), apiHeaders(token))
}

func GetFullNodeApi(repo string) (api.FullNode, jsonrpc.ClientCloser, error) {
	addr, headers, err := getApi(repo)
	if err != nil {
		return nil, nil, err
	}

	return client.NewFullNodeRPC(addr, headers)
}

func getApi(path string) (string, http.Header, error) {
	r, err := repo.NewFS(path)
	if err != nil {
		return "", nil, err
	}

	ma, err := r.APIEndpoint()
	if err != nil {
		return "", nil, xerrors.Errorf("failed to get api endpoint: %w", err)
	}

	_, addr, err := manet.DialArgs(ma)
	if err != nil {
		return "", nil, err
	}

	var headers http.Header
	token, err := r.APIToken()
	if err != nil {
		// log
	} else {
		headers = apiHeaders(string(token))
	}

	return apiUri(addr), headers, nil
}

func apiUri(addr string) string {
	return "ws://" + addr + "/rpc/v0"
}

func apiHeaders(token string) http.Header {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	return headers
}
