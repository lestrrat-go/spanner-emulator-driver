package emulator

import "github.com/lestrrat-go/option"

type Option = option.Interface

type identGRPCPort struct{}
type identRESTPort struct{}
type identNotifyReady struct{}

func WithGRPCPort(v int) Option {
	return option.New(identGRPCPort{}, v)
}

func WithRESTPort(v int) Option {
	return option.New(identRESTPort{}, v)
}

func WithNotifyReady(notifyFunc func()) Option {
	return option.New(identNotifyReady{}, notifyFunc)
}
