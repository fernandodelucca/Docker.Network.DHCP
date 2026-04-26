package util

import (
	"context"
	"time"

	"github.com/moby/moby/client"
	log "github.com/sirupsen/logrus"
)

const (
	OptionsKeyGeneric = "com.docker.network.generic"
)

func AwaitContainerInspect(ctx context.Context, docker *client.Client, id string, interval time.Duration) (client.ContainerInspectResult, error) {
	var err error
	ctrChan := make(chan client.ContainerInspectResult)
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			var ctr client.ContainerInspectResult
			ctr, err = docker.ContainerInspect(ctx, id, client.ContainerInspectOptions{})
			if err == nil {
				ctrChan <- ctr
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
			}
		}
	}()

	var dummy client.ContainerInspectResult
	select {
	case ctr := <-ctrChan:
		return ctr, nil
	case <-ctx.Done():
		if err != nil {
			log.WithError(err).WithField("id", id).Error("Failed to await container by ID")
		}
		return dummy, ctx.Err()
	}
}
