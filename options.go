package driver

import "github.com/lestrrat-go/option"

type Option = option.Interface

type identDropDatabase struct{}

func WithDropDatabase(v bool) Option {
	return option.New(identDropDatabase{}, v)
}
