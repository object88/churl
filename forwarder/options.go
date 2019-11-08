package forwarder

import "io"

type Option func(o *Options) error

type Options struct {
	err io.Writer
	out io.Writer

	ready chan struct{}
}

func NewOptions() *Options {
	return &Options{
		out: &nilwriter{},
	}
}

func Err(w io.Writer) Option {
	return func(o *Options) error {
		if w == nil {
			w = &nilwriter{}
		}
		o.err = w
		return nil
	}
}

func Out(w io.Writer) Option {
	return func(o *Options) error {
		if w == nil {
			w = &nilwriter{}
		}
		o.out = w
		return nil
	}
}

func Ready(c chan struct{}) Option {
	return func(o *Options) error {
		o.ready = c
		return nil
	}
}
