package driver_test

import (
	"context"
	"os"
	"testing"
	"time"

	driver "github.com/lestrrat-go/spanner-emulator-driver"
	"github.com/stretchr/testify/require"
)

func TestDriver(t *testing.T) {
	dsn := driver.Config{
		Project:  `driver-test-project`,
		Instance: `driver-test-instance`,
		Database: `driver-test`,
	}
	t.Logf("%s", dsn.FormatDSN())
	d, err := driver.New(dsn.FormatDSN())
	require.NoError(t, err, `driver.New should succeed`)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var options []driver.Option
	if v := os.Getenv(`SPANNER_EMULATOR_DRIVER_TEST_INSTANCE_CONFIG`); v != "" {
		options = append(options, driver.WithInstanceConfig(v))
	}

	exited := d.Run(ctx, options...)

	require.NoError(t, d.Ready(ctx), `driver should start successfully`)

	<-exited
}
