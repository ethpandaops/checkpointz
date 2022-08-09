package cache

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type item struct {
	value      interface{}
	lastAccess time.Time
}

type sortableItem struct {
	key        string
	lastAccess time.Time
}

type TTLMap struct {
	m        map[string]*item
	l        sync.Mutex
	maxItems int

	evictCallbacks []func(string, interface{})
}

// NewTTLMap returns a new TTLMap.
func NewTTLMap(maxItems int, ttl time.Duration) (m *TTLMap) {
	m = &TTLMap{
		m:        make(map[string]*item, maxItems),
		maxItems: maxItems,
	}

	go func() {
		for now := range time.Tick(time.Second * 1) {
			for k, v := range m.m {
				if v.lastAccess.Add(ttl).Before(now) {
					m.Delete(k)
				}
			}
		}
	}()

	return
}

func (m *TTLMap) OnItemEvicted(f func(string, interface{})) {
	m.evictCallbacks = append(m.evictCallbacks, f)
}

func (m *TTLMap) Delete(k string) {
	val, err := m.Get(k)
	if err != nil {
		return
	}

	m.l.Lock()

	delete(m.m, k)

	for _, f := range m.evictCallbacks {
		go f(k, val)
	}

	m.l.Unlock()
}

func (m *TTLMap) evictOldestItem() {
	// This is a very naive implementation.
	items := []sortableItem{}

	for k, v := range m.m {
		items = append(items, sortableItem{
			key:        k,
			lastAccess: v.lastAccess,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].lastAccess.Before(items[j].lastAccess)
	})

	if len(items) > 0 {
		m.Delete(items[0].key)
	}
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Add(k string, v interface{}) {
	if m.Len() >= m.maxItems {
		m.evictOldestItem()
	}

	m.l.Lock()

	defer m.l.Unlock()

	it, ok := m.m[k]
	if !ok {
		it = &item{value: v}
		m.m[k] = it
	}

	it.lastAccess = time.Now()
}

func (m *TTLMap) Get(k string) (interface{}, error) {
	m.l.Lock()

	defer m.l.Unlock()

	it, ok := m.m[k]
	if !ok {
		return nil, errors.New("not found")
	}

	it.lastAccess = time.Now()

	return it.value, nil
}
