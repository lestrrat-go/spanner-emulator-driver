// Package driver provides utilities to control Google's Spanner emulator
// for testing purposes.
package driver

import (
	"context"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"github.com/lestrrat-go/spanner-emulator-driver/emulator"
	databasepb "google.golang.org/genproto/googleapis/spanner/admin/database/v1"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
	"google.golang.org/grpc/codes"
)

// Driver is the main object to control the spanner emulator.
// The zero value should not be used. Always use the value
// returned from driver.New
type Driver struct {
	mu         *sync.RWMutex
	cond       *sync.Cond
	dsn        string
	config     *Config
	ready      bool
	setupError error
}

func New(dsn string) (*Driver, error) {
	// Kind of silly, but we parse back the dsn
	config, err := ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse DSN: %w`, err)
	}

	mu := &sync.RWMutex{}
	return &Driver{
		mu:     mu,
		cond:   sync.NewCond(mu),
		config: config,
		dsn:    dsn,
	}, nil
}

// Run controls the emulator running in docker. The environment variable
// SPANNER_EMULATOR_HOST will also be set to the appropriate value
func (d *Driver) Run(ctx context.Context) {
	defer d.cond.Broadcast()

	// channel to notify readiness to the user
	d.mu.Lock()
	d.ready = false
	d.mu.Unlock()

	// Setup environment variable
	os.Setenv(`SPANNER_EMULATOR_HOST`, fmt.Sprintf(`localhost:%d`, emulator.DefaultGRPCPort))

	// channel to notify _US_ that the emulator is ready
	emuReady := make(chan struct{})

	exited := make(chan error, 1)
	go func(ctx context.Context) {
		defer close(exited)
		if err := emulator.Run(ctx, emulator.WithNotifyReady(func() { close(emuReady) })); err != nil {
			select {
			case <-ctx.Done():
			case exited <- err:
			}
		}
	}(ctx)

	select {
	case <-ctx.Done():
		d.notifyReady(fmt.Errorf(`context canceled exited before emulator became ready`))
		return
	case err := <-exited:
		// WHAT?!
		d.notifyReady(fmt.Errorf(`emulator exited before becoming ready: %w`, err))
		return
	case <-emuReady:
		// ready, go on
	}

	// start preparing
	if err := d.setup(ctx); err != nil {
		d.notifyReady(fmt.Errorf(`failed to setup spanner emulator: %w`, err))
		return
	}

	d.notifyReady(nil)
	return
}

func (d *Driver) notifyReady(err error) {
	d.mu.Lock()
	d.ready = true
	d.setupError = err
	d.mu.Unlock()
	d.cond.Broadcast()
}

func (d *Driver) Ready(ctx context.Context) error {
	d.cond.L.Lock()
	for !d.ready {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		d.cond.Wait()
	}
	d.cond.L.Unlock()

	return d.setupError
}

func (d *Driver) setup(ctx context.Context) error {
	if err := d.createSpannerInstance(ctx); err != nil {
		return fmt.Errorf(`failed to create spanner instance: %w`, err)
	}

	if err := d.createSpannerDatabase(ctx); err != nil {
		return fmt.Errorf(`failed to create spanner database: %w`, err)
	}
	return nil
}

func (d *Driver) createSpannerInstance(ctx context.Context) error {
	instanceAdminClient, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create instance admin client: %w`, err)
	}
	defer instanceAdminClient.Close()

	if _, err := instanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: projectMarker + d.config.Project + instanceMarker + d.config.Instance,
	}); err == nil {
		// instance already exists
		return nil
	}

	if err != nil && spanner.ErrCode(err) != codes.NotFound {
		return fmt.Errorf(`unexpected error while retrieving instance: %w`, err)
	}

	if _, err := instanceAdminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     projectMarker + d.config.Project,
		InstanceId: d.config.Instance,
	}); err != nil {
		return fmt.Errorf(`failed to create instance %q: %w`, d.config.Instance, err)
	}

	return nil
}

func (d *Driver) createSpannerDatabase(ctx context.Context) error {
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return fmt.Errorf(`failed to create a database admin client: %w`, err)
	}

	_, err = adminClient.GetDatabase(ctx, &databasepb.GetDatabaseRequest{
		Name: d.dsn,
	})
	switch {
	case err == nil:
		fmt.Printf("Database %q already exist\n", d.dsn)
		// if the database exists, we just use it
		return nil
	case err != nil && spanner.ErrCode(err) != codes.NotFound:
		return fmt.Errorf(`unexpected error while retrieving database: %w`, err)
	default:
		// no op, go to next
	}

	op, err := adminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          projectMarker + d.config.Project + instanceMarker + d.config.Instance,
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", d.config.Database),
		//		ExtraStatements: extraStatements,
	})
	if err != nil {
		return fmt.Errorf(`create database call failed: %w`, err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf(`create database failed while waiting for the operation to complete: %w`, err)
	}

	return nil
}