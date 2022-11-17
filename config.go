package driver

import (
	"fmt"
	"strings"
)

// Config is an utility object that works like `(github.com/go-sql-mysql).Config`
// which allows users to use existing tools to import connection information.
type Config struct {
	Project  string // Name of the project
	Instance string // Name of the instance
	Database string // Name of the database
}

const databaseMarker = `/databases/`
const instanceMarker = `/instances/`
const projectMarker = `projects/`

func ParseDSN(dsn string) (*Config, error) {
	// we won't check for i < 0 and what not.
	databaseMarkerIdx := strings.Index(dsn, databaseMarker)
	instanceMarkerIdx := strings.Index(dsn, instanceMarker)

	if databaseMarkerIdx == -1 {
		return nil, fmt.Errorf(`could not find marker %q in %q`, databaseMarker, dsn)
	}
	if instanceMarkerIdx == -1 {
		return nil, fmt.Errorf(`could not find marker %q in %q`, instanceMarker, dsn)
	}

	if databaseMarkerIdx < instanceMarkerIdx {
		return nil, fmt.Errorf(`invalid dsn: expected projects/PROJECT/instances/INSTANCE/databases/DATABASE, got %q`, dsn)
	}

	projectName := strings.TrimPrefix(dsn[:instanceMarkerIdx], projectMarker)
	instanceName := dsn[instanceMarkerIdx+len(instanceMarker) : databaseMarkerIdx]
	databaseName := dsn[databaseMarkerIdx+len(databaseMarker):]

	if projectName == "" {
		return nil, fmt.Errorf(`could not find a project name from %q`, dsn)
	}

	if instanceName == "" {
		return nil, fmt.Errorf(`could not find an instance name from %q`, dsn)
	}

	if databaseName == "" {
		return nil, fmt.Errorf(`could not find a database name from %q`, dsn)
	}

	return &Config{
		Project:  projectName,
		Instance: instanceName,
		Database: databaseName,
	}, nil
}

// FormatDSN formats the given Config into a DSN string which can be
// passed to the spanner driver
func (cfg *Config) FormatDSN() string {
	// specify defaults to make sure that the user can see unspecified fields
	// when they receive and error
	project := `!!UNSPECIFIED!!`
	instance := `!!UNSPECIFIED!!`
	database := `!!UNSPECIFIED!!`
	if v := cfg.Project; v != "" {
		project = v
	}
	if v := cfg.Instance; v != "" {
		instance = v
	}
	if v := cfg.Database; v != "" {
		database = v
	}
	return fmt.Sprintf(`%s%s%s%s%s%s`, projectMarker, project, instanceMarker, instance, databaseMarker, database)
}
