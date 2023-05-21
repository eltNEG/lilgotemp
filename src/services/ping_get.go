package services

import (
	"context"
	"eltneg/goliltemp/src/config"
	"eltneg/goliltemp/src/dependencies"
	"net/http"
)

type PingReq struct{}

type PingRes struct {
	Version string `json:"version,omitempty"`
}

// Controller returns the result of the logic
func (d *PingReq) Controller(ctx context.Context, cfg *config.Config, dpc *dependencies.Dependencies) (status int, msg string, data *PingRes, err error) {
	data = &PingRes{
		Version: cfg.Version,
	}

	return http.StatusOK, "pong", data, err
}
