package cache

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type item struct {
	value     interface{}
	expiresAt time.Time
}

type sortableItem struct {
	key       string
	expiresAt time.Time
}

type TTLMap struct {
	m        map[string]*item
	l        sync.Mutex
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
			for k, v := range m.m {
				if v.expiresAt.Before(now) {
					m.Delete(k)
				}
			}
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
	val, expiresAt, err := m.Get(k)
	if err != nil {
		return
	}

	m.l.Lock()

	delete(m.m, k)

	m.metrics.ObserveOperations(OperationDEL, 1)

	for _, f := range m.deletedCallbacks {
		go f(k, val, expiresAt)
	}

	m.l.Unlock()
}

func (m *TTLMap) evictItemToClosestToExpiry() {
	// This is a very naive implementation.
	items := []sortableItem{}

	for k, v := range m.m {
		items = append(items, sortableItem{
			key:       k,
			expiresAt: v.expiresAt,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].expiresAt.Before(items[j].expiresAt)
	})

	if len(items) > 0 {
		m.Delete(items[0].key)
		m.metrics.ObserveOperations(OperationEVICT, 1)
	}
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Add(k string, v interface{}, expiresAt time.Time) {
	if m.Len() >= m.maxItems {
		m.evictItemToClosestToExpiry()
	}

	m.l.Lock()

	defer m.l.Unlock()

	it, ok := m.m[k]
	if !ok {
		it = &item{
			value:     v,
			expiresAt: expiresAt,
		}
		m.m[k] = it
	}

	m.metrics.ObserveOperations(OperationADD, 1)

	for _, f := range m.addedCallbacks {
		go f(k, v, it.expiresAt)
	}
}

func (m *TTLMap) Get(k string) (interface{}, time.Time, error) {
	m.metrics.ObserveOperations(OperationGET, 1)

	m.l.Lock()

	defer m.l.Unlock()

	it, ok := m.m[k]
	if !ok {
		m.metrics.ObserveMiss()

		return nil, time.Now(), errors.New("not found")
	}

	m.metrics.ObserveHit()

	return it.value, it.expiresAt, nil
}
