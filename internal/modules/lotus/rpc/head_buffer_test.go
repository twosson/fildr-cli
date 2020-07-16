package rpc

import (
	"github.com/filecoin-project/lotus/api"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHeadBuffer(t *testing.T) {
	hb := NewHeadBuffer(5)
	require.Nil(t, hb.Push(&api.HeadChange{Type: "1"}))
	require.Nil(t, hb.Push(&api.HeadChange{Type: "2"}))
	require.Nil(t, hb.Push(&api.HeadChange{Type: "3"}))
	require.Nil(t, hb.Push(&api.HeadChange{Type: "4"}))
	require.Nil(t, hb.Push(&api.HeadChange{Type: "5"}))

	hc := hb.Push(&api.HeadChange{Type: "6"})

	require.Equal(t, hc.Type, "1")

}
