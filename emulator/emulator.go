// Package emulator provides simple utilities to control the lifecycle of
// a spanner emulator. The emulator is run through Docker, therefore
// you must have the `docker` binary vailable in your environment.

package emulator

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

var DockerImage = `gcr.io/cloud-spanner-emulator/emulator`
var DefaultRESTPort = 9020
var DefaultGRPCPort = 9010

func Run(ctx context.Context, options ...Option) error {
	grpcSrcPort := DefaultGRPCPort
	restSrcPort := DefaultRESTPort
	stopContainer := true
	var onExit func() error
	var notifyReady func()

	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case identGRPCPort{}:
			grpcSrcPort = option.Value().(int)
		case identRESTPort{}:
			restSrcPort = option.Value().(int)
		case identNotifyReady{}:
			notifyReady = option.Value().(func())
		case identStopContainer{}:
			stopContainer = option.Value().(bool)
		case identOnExit{}:
			onExit = option.Value().(func() error)
		}
	}
	grpcPortPublishSpec := fmt.Sprintf(`%d:9010`, grpcSrcPort)
	restPortPublishSpec := fmt.Sprintf(`%d:9020`, restSrcPort)

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	name := fmt.Sprintf("spanner-emulator-%d-%d", grpcSrcPort, restSrcPort)
	cmd := exec.CommandContext(childCtx, "docker", "run", "-i", "--rm", "-p", grpcPortPublishSpec, "-p", restPortPublishSpec, "--name", name, DockerImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	exited := make(chan error, 1)
	go func() {
		defer close(exited)
		err := cmd.Run()

		if onExit != nil {
			if err := onExit(); err != nil {
				log.Printf(`failed to execute onExit hook: %s`, err)
			}
		}

		if stopContainer {
			if err := exec.Command("docker", "stop", name).Run(); err != nil {
				log.Printf("failed to stop docker container %q: %s", name, err)
			}
		}
		exited <- err
	}()

	time.Sleep(time.Second)

	// Once the process starts, wait until the ports are available
	timeout := time.NewTimer(30 * time.Second)
OUTER:
	for {
		select {
		case err := <-exited:
			return fmt.Errorf(`emulator prematurely exited: %w`, err)
		case <-timeout.C:
			return fmt.Errorf(`timed out while waiting for emulator to be available`)
		default:
			time.Sleep(500 * time.Millisecond)
		}

		for _, port := range []int{grpcSrcPort, restSrcPort} {
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			log.Printf("connecting to %q", addr)
			_, err := net.DialTimeout("tcp", addr, time.Second)
			if err == nil {
				break OUTER
			}
			log.Printf("failed to connect to %s (%s)", addr, err)
			break
		}
	}

	if notifyReady != nil {
		notifyReady()
	}

	<-ctx.Done()
	<-exited
	return nil
}
