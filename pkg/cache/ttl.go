package cache

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiresAt  time.Time
	invincible bool
}

type sortableItem struct {
	key       string
	expiresAt time.Time
}

type TTLMap struct {
	m        map[string]*item
	l        sync.RWMutex
	maxItems int

	metrics Metrics

	deletedCallbacks []func(string, interface{}, time.Time)
	addedCallbacks   []func(string, interface{}, time.Time)
}

// NewTTLMap returns a new TTLMap.
func NewTTLMap(maxItems int, name, namespace string) (m *TTLMap) {
	m = &TTLMap{
		m:        make(map[string]*item, maxItems),
		maxItems: maxItems,
		metrics:  NewMetrics(name, namespace+"_ttlmap"),
	}

	go func() {
		for now := range time.Tick(time.Second * 1) {
			m.l.Lock()

			for k, v := range m.m {
				if v.invincible {
					continue
				}

				if v.expiresAt.Before(now) {
					m.delete(k, v.value, v.expiresAt)
				}
			}

			m.l.Unlock()
		}
	}()

	return
}

func (m *TTLMap) EnableMetrics(namespace string) {
	m.metrics.Register()

	m.OnItemAdded(func(k string, v interface{}, e time.Time) {
		m.metrics.ObserveLen(m.Len())
	})

	m.OnItemDeleted(func(k string, v interface{}, e time.Time) {
		m.metrics.ObserveLen(m.Len())
	})
}

func (m *TTLMap) OnItemDeleted(f func(string, interface{}, time.Time)) {
	m.deletedCallbacks = append(m.deletedCallbacks, f)
}

func (m *TTLMap) OnItemAdded(f func(string, interface{}, time.Time)) {
	m.addedCallbacks = append(m.addedCallbacks, f)
}

func (m *TTLMap) Delete(k string) {
	m.l.Lock()
	defer m.l.Unlock()

	val, expiresAt, err := m.get(k)
	if err != nil {
		return
	}

	m.delete(k, val, expiresAt)
}

func (m *TTLMap) delete(k string, val interface{}, expiresAt time.Time) {
	delete(m.m, k)

	m.metrics.ObserveOperations(OperationDEL, 1)

	for _, f := range m.deletedCallbacks {
		go f(k, val, expiresAt)
	}
}

func (m *TTLMap) evictItemToClosestToExpiry() {
	// This is a very naive implementation.
	evictableItems := make([]sortableItem, 0, len(m.m))

	// Get all non-invincible items.
	for k, v := range m.m {
		if v.invincible {
			continue
		}

		evictableItems = append(evictableItems, sortableItem{
			key:       k,
			expiresAt: v.expiresAt,
		})
	}

	if len(evictableItems) == 0 {
		return
	}

	sort.Slice(evictableItems, func(i, j int) bool {
		return evictableItems[i].expiresAt.Before(evictableItems[j].expiresAt)
	})

	m.delete(evictableItems[0].key, evictableItems[0].expiresAt, evictableItems[0].expiresAt)
	m.metrics.ObserveOperations(OperationEVICT, 1)
}

func (m *TTLMap) Len() int {
	m.l.RLock()
	defer m.l.RUnlock()

	return m.len()
}

func (m *TTLMap) len() int {
	return len(m.m)
}

func (m *TTLMap) Add(k string, v interface{}, expiresAt time.Time, invincible bool) {
	m.l.Lock()
	defer m.l.Unlock()

	m.add(k, v, expiresAt, invincible)
}

func (m *TTLMap) add(k string, v interface{}, expiresAt time.Time, invincible bool) {
	if m.len() >= m.maxItems {
		m.evictItemToClosestToExpiry()
	}

	it, ok := m.m[k]
	if !ok {
		it = &item{
			value:      v,
			expiresAt:  expiresAt,
			invincible: invincible,
		}
		m.m[k] = it
	}

	m.metrics.ObserveOperations(OperationADD, 1)

	for _, f := range m.addedCallbacks {
		go f(k, v, it.expiresAt)
	}
}

func (m *TTLMap) Get(k string) (interface{}, time.Time, error) {
	m.l.RLock()
	itv, expires, err := m.get(k)
	m.l.RUnlock()

	if err != nil {
		return nil, time.Now(), err
	}

	return itv, expires, err
}

func (m *TTLMap) get(k string) (interface{}, time.Time, error) {
	m.metrics.ObserveOperations(OperationGET, 1)

	it, ok := m.m[k]
	if !ok {
		m.metrics.ObserveMiss()
		return nil, time.Now(), errors.New("not found")
	}

	m.metrics.ObserveHit()

	return it.value, it.expiresAt, nil
}
