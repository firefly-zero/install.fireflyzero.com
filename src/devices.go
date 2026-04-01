package src

import (
	"sync"
)

type Devices = SyncMap[uint32, DeviceConn]

func newDevices() Devices {
	return Devices{
		mx:    &sync.Mutex{},
		items: map[uint32]DeviceConn{},
	}
}

type SyncMap[K comparable, V any] struct {
	mx    *sync.Mutex
	items map[K]V
}

func (m SyncMap[K, V]) add(k K, v V) {
	m.mx.Lock()
	m.items[k] = v
	m.mx.Unlock()
}

func (m SyncMap[K, V]) pop(k K) (V, bool) {
	m.mx.Lock()
	v, found := m.items[k]
	if found {
		delete(m.items, k)
	}
	m.mx.Unlock()
	return v, found
}
