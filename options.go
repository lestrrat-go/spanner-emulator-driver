package driver

import "github.com/lestrrat-go/option"

type Option = option.Interface

type identDropDatabase struct{}
type identDDLDirectory struct{}
type identUseEmulator struct{}
type identInstanceConfig struct{}

func WithDropDatabase(v bool) Option {
	return option.New(identDropDatabase{}, v)
}

func WithDDLDirectory(dir string) Option {
	return option.New(identDDLDirectory{}, dir)
}

func WithUseEmulator(v bool) Option {
	return option.New(identUseEmulator{}, v)
}

func WithInstanceConfig(s string) Option {
	return option.New(identInstanceConfig{}, s)
}
