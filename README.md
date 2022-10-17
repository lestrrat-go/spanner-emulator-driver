spanner-emulator-driver
=======================

Simple utility to run google's Spanner database emulator, aimed for
testing purposes.

# SYNOPSIS

```go
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

	d, err := driver.New(dsn.FormatDSN())
	require.NoError(t, err, `driver.New should succeed`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exited := d.Run(ctx)

	require.NoError(t, d.Ready(ctx), `driver should start successfully`)

	// Your tests go here

	// Make sure to wait for the emulator to exit
	<-exited
}
```
