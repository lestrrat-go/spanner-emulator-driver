package emulator

import "github.com/lestrrat-go/option"

type Option = option.Interface

type identGRPCPort struct{}
type identRESTPort struct{}
type identNotifyReady struct{}
type identStopContainer struct{}
type identOnExit struct{}

func WithGRPCPort(v int) Option {
	return option.New(identGRPCPort{}, v)
}

func WithRESTPort(v int) Option {
	return option.New(identRESTPort{}, v)
}

func WithNotifyReady(notifyFunc func()) Option {
	return option.New(identNotifyReady{}, notifyFunc)
}

func WithStopContainer(v bool) Option {
	return option.New(identStopContainer{}, v)
}

func WithOnExit(f func() error) Option {
	return option.New(identOnExit{}, f)
}
