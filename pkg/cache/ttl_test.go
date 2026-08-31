package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestItemAdds(t *testing.T) {
	instance := NewTTLMap(10, "", "")

	key := "key1"
	value := "value1"

	instance.Add(key, value, time.Now().Add(time.Hour), false)

	data, _, err := instance.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if data != value {
		t.Fatalf("Expected %s, got %s", value, data)
	}
}

func TestItemDeletes(t *testing.T) {
	instance := NewTTLMap(10, "", "")

	key := "key2"
	value := "value2"

	instance.Add(key, value, time.Now().Add(time.Hour), false)
	instance.Delete(key)

	_, _, err := instance.Get(key)
	if err == nil {
		t.Fatalf("Expected item to have been deleted")
	}
}

func TestItemDoesExpire(t *testing.T) {
	instance := NewTTLMap(10, "", "")

	key := "key3"
	value := "value3"

	instance.Add(key, value, time.Now().Add(time.Second), false)

	time.Sleep(time.Second * 3)

	_, _, err := instance.Get(key)
	if err == nil {
		t.Fatalf("Expected item to have expired")
	}
}

func TestMaxItems(t *testing.T) {
	instance := NewTTLMap(3, "", "")
	for i := 1; i <= 10; i++ {
		instance.Add(fmt.Sprintf("key%d", i), "value", time.Now().Add(time.Hour), false)
	}

	if len(instance.m) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(instance.m))
	}
}

func TestMaxItemsEvictsOldest(t *testing.T) {
	instance := NewTTLMap(3, "", "")
	for i := 1; i <= 10; i++ {
		instance.Add(fmt.Sprintf("key%d", i), "value", time.Now().Add(time.Hour).Add(time.Second*time.Duration(i)), false)
	}

	if _, _, err := instance.Get("key1"); err == nil {
		t.Fatalf("Expected item to have been evicted")
	}

	if _, _, err := instance.Get("key7"); err == nil {
		t.Fatalf("Expected item to have been evicted")
	}

	if _, _, err := instance.Get("key8"); err != nil {
		t.Fatalf("Expected item to not have been evicted")
	}

	if _, _, err := instance.Get("key9"); err != nil {
		t.Fatalf("Expected item to not have been evicted")
	}

	if _, _, err := instance.Get("key10"); err != nil {
		t.Fatalf("Expected item to not have been evicted")
	}
}

func TestCallbacks(t *testing.T) {
	instance := NewTTLMap(10, "", "")

	// Callbacks fire in their own goroutines, so synchronise on channels
	// rather than reading shared state after a sleep.
	evicted := make(chan [2]string, 1)
	added := make(chan [2]string, 1)

	instance.OnItemDeleted(func(key string, value interface{}, expiresAt time.Time) {
		str, _ := value.(string)
		evicted <- [2]string{key, str}
	})

	instance.OnItemAdded(func(key string, value interface{}, expiresAt time.Time) {
		str, _ := value.(string)
		added <- [2]string{key, str}
	})

	instance.Add("key4", "value4", time.Now().Add(time.Hour), false)
	instance.Delete("key4")

	for _, tc := range []struct {
		name string
		ch   chan [2]string
	}{
		{"added", added},
		{"evicted", evicted},
	} {
		select {
		case got := <-tc.ch:
			if got[0] != "key4" || got[1] != "value4" {
				t.Fatalf("%s callback: expected key4/value4, got %s/%s", tc.name, got[0], got[1])
			}
		case <-time.After(time.Second * 5):
			t.Fatalf("Expected %s callback to have been called", tc.name)
		}
	}
}

func TestInvincible(t *testing.T) {
	ttlMap := NewTTLMap(2, "myCache", "default")

	now := time.Now()

	// add an item with 1 second expiry
	ttlMap.Add("key1", "value1", now.Add(time.Second), false)

	// add an item with 2 second expiry and invincible
	ttlMap.Add("key2", "value2", now.Add(time.Second*2), true)

	// add an item with 3 second expiry
	ttlMap.Add("key3", "value3", now.Add(time.Second*3), false)

	time.Sleep(5 * time.Second)

	// get key2, should be present as it is invincible
	_, _, err := ttlMap.Get("key2")
	if err != nil {
		t.Error("key2 not found")
	}

	// get key1, should not be found as it has expired
	_, _, err = ttlMap.Get("key1")
	if err == nil {
		t.Error("key1 should not be found")
	}

	// get key2, should be present as it is invincible
	_, _, err = ttlMap.Get("key3")
	if err == nil {
		t.Error("key2 should not be found")
	}
}
