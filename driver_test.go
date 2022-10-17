package driver_test

import (
	"context"
	"testing"

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go d.Run(ctx)

	require.NoError(t, d.Ready(ctx), `driver should start successfully`)
}
