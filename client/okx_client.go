package client

import (
	"best-ticker/config"
	"best-ticker/utils/logger"
	"context"
	"github.com/drinkthere/okx"
	"github.com/drinkthere/okx/api"
)

type OkxClient struct {
	Client *api.Client
}

func (okxClient *OkxClient) Init(cfg *config.OkxConfig, colo string, localIP string) bool {
	var dest okx.Destination
	if colo == "zoneB" {
		dest = okx.ColoServer
	} else if colo == "zoneD" {
		dest = okx.ColoDServer
	} else {
		dest = okx.NormalServer
	}

	ctx := context.Background()

	var client *api.Client
	var err error
	if localIP == "" {
		client, err = api.NewClient(ctx, cfg.OkxAPIKey, cfg.OkxSecretKey, cfg.OkxPassword, dest)
	} else {
		client, err = api.NewClientWithIP(ctx, cfg.OkxAPIKey, cfg.OkxSecretKey, cfg.OkxPassword, dest, localIP)
	}
	if err != nil {
		logger.Error(err.Error())
		return false
	}

	okxClient.Client = client
	return true
}
