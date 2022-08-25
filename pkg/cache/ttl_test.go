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

	instance.Add(key, value, time.Now().Add(time.Hour))

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

	instance.Add(key, value, time.Now().Add(time.Hour))
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

	instance.Add(key, value, time.Now().Add(time.Second))

	time.Sleep(time.Second * 3)

	_, _, err := instance.Get(key)
	if err == nil {
		t.Fatalf("Expected item to have expired")
	}
}

func TestMaxItems(t *testing.T) {
	instance := NewTTLMap(3, "", "")
	for i := 1; i <= 10; i++ {
		instance.Add(fmt.Sprintf("key%d", i), "value", time.Now().Add(time.Hour))
	}

	if len(instance.m) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(instance.m))
	}
}

func TestMaxItemsEvictsOldest(t *testing.T) {
	instance := NewTTLMap(3, "", "")
	for i := 1; i <= 10; i++ {
		instance.Add(fmt.Sprintf("key%d", i), "value", time.Now().Add(time.Hour).Add(time.Second*time.Duration(i)))
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

	evictedCallback := false

	instance.OnItemDeleted(func(key string, value interface{}, expiresAt time.Time) {
		if key != "key4" {
			t.Fatalf("Expected key to be key4, got %s", key)
		}

		if value != "value4" {
			t.Fatalf("Expected value to be value4, got %s", value)
		}

		evictedCallback = true
	})

	addedCallback := false

	instance.OnItemAdded(func(key string, value interface{}, expiresAt time.Time) {
		if key != "key4" {
			t.Fatalf("Expected key to be key4, got %s", key)
		}

		if value != "value4" {
			t.Fatalf("Expected value to be value4, got %s", value)
		}

		addedCallback = true
	})

	instance.Add("key4", "value4", time.Now().Add(time.Hour))
	instance.Delete("key4")

	time.Sleep(time.Second * 1)

	if !evictedCallback {
		t.Fatalf("Expected evicted callback to have been called")
	}

	if !addedCallback {
		t.Fatalf("Expected added callback to have been called")
	}
}
