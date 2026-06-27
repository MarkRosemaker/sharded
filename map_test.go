package sharded

import (
	"strconv"
	"sync"
	"testing"
)

func TestMap_Basic(t *testing.T) {
	m := NewStringMap[int]()

	// empty map
	if v, ok := m.Get("missing"); ok || v != 0 {
		t.Error("expected zero value for missing key")
	}
	if m.Len() != 0 {
		t.Error("len of empty map")
	}

	// set + get
	m.Set("key1", 42)
	if v, ok := m.Get("key1"); !ok || v != 42 {
		t.Error("get after set failed")
	}

	// overwrite
	m.Set("key1", 99)
	if v, ok := m.Get("key1"); !ok || v != 99 {
		t.Error("overwrite failed")
	}

	// delete
	m.Delete("key1")
	if _, ok := m.Get("key1"); ok {
		t.Error("delete failed")
	}
	if m.Len() != 0 {
		t.Error("len after delete")
	}

	// Clear
	m.Set("a", 1)
	m.Set("b", 2)
	m.Clear()
	if m.Len() != 0 {
		t.Error("clear failed")
	}
}

func TestMap_Range(t *testing.T) {
	m := NewStringMap[int]()
	keys := []string{"a", "b", "c", "d"}
	for i, k := range keys {
		m.Set(k, i+10)
	}

	seen := make(map[string]int)
	m.Range(func(k string, v int) bool {
		seen[k] = v
		return true
	})

	if len(seen) != len(keys) {
		t.Error("range missed keys")
	}
	for _, k := range keys {
		if _, ok := seen[k]; !ok {
			t.Errorf("missing key %s", k)
		}
	}

	// early stop
	count := 0
	m.Range(func(string, int) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Error("early stop failed")
	}
}

func TestMap_Concurrency(t *testing.T) {
	m := NewStringMap[int]()
	const goroutines = 64
	const ops = 200

	wg := sync.WaitGroup{}
	for i := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range ops {
				key := strconv.Itoa(id*ops + j)
				m.Set(key, id)
				if v, ok := m.Get(key); !ok || v != id {
					t.Error("concurrent get mismatch")
				}
				if j%5 == 0 {
					m.Delete(key) // occasional delete
				}
			}
		}(i)
	}
	wg.Wait()

	// rough check
	if m.Len() > goroutines*ops {
		t.Error("too many items")
	}
}
