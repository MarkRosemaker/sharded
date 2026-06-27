package sharded

import (
	"strconv"
	"sync"
	"testing"
)

func TestMap_Basic(t *testing.T) {
	m := NewMap()

	// empty
	if v, ok := m.Get("missing"); ok || v != "" {
		t.Error("expected missing")
	}
	if m.Len() != 0 {
		t.Error("len should be 0")
	}

	// set + get
	m.Set("key1", "value1")
	if v, ok := m.Get("key1"); !ok || v != "value1" {
		t.Error("get failed")
	}

	// overwrite
	m.Set("key1", "newvalue")
	if v, ok := m.Get("key1"); !ok || v != "newvalue" {
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
}

func TestMap_Concurrency(t *testing.T) {
	m := NewMap()
	const goroutines = 100
	const keysPer = 100

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range keysPer {
				key := "key-" + strconv.Itoa(id) + "-" + strconv.Itoa(j)
				m.Set(key, "val")
				if v, ok := m.Get(key); !ok || v != "val" {
					t.Error("concurrent get failed")
				}
			}
		}(i)
	}
	wg.Wait()

	if m.Len() != goroutines*keysPer {
		t.Errorf("expected %d items, got %d", goroutines*keysPer, m.Len())
	}
}
