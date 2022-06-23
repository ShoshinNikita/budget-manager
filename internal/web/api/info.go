package api

import (
	"context"
	"time"
)

type getInfoResp struct {
	Version string `json:"version"`
	GitHash string `json:"git_hash"`
	Uptime  string `json:"uptime"`
}

func (api API) getInfo(context.Context, *emptyReq) (*getInfoResp, error) {
	return &getInfoResp{
		Version: api.version,
		GitHash: api.gitHash,
		Uptime:  time.Since(api.startTime).String(),
	}, nil
}
