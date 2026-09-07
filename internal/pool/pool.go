package pool

import "sync"

type Reseter[T any] interface {
	*T
	Reset()
}

type Pool[T any, P Reseter[T]] struct {
	p sync.Pool
}

func New[T any, P Reseter[T]](newFunc func() P) *Pool[T, P] {
	if newFunc == nil {
		newFunc = func() P {
			return new(T)
		}
	}
	return &Pool[T, P]{
		p: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *Pool[T, P]) Get() P {
	return p.p.Get().(P)
}

func (p *Pool[T, P]) Put(x P) {
	x.Reset()
	p.p.Put(x)
}
