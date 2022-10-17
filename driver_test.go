package driver_test

import (
	"context"
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

	exited := d.Run(ctx)

	require.NoError(t, d.Ready(ctx), `driver should start successfully`)

	<-exited
}
