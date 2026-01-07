package sim

import "sync"

type StatePool struct {
	pool sync.Pool
	size int
}

func NewStatePool(stateSize int) *StatePool {
	return &StatePool{
		size: stateSize,
		pool: sync.Pool{
			New: func() interface{} {
				return make(State, stateSize)
			},
		},
	}
}

func (p *StatePool) Get() State {
	return p.pool.Get().(State)
}

func (p *StatePool) Put(s State) {
	if len(s) == p.size {
		for i := range s {
			s[i] = 0
		}
		p.pool.Put(s)
	}
}

func (p *StatePool) GetAndCopy(src State) State {
	dst := p.Get()
	copy(dst, src)
	return dst
}
