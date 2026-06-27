package sharded

import (
	"maps"
	"strconv"
	"sync"
	"testing"
)

func TestMap_Basic(t *testing.T) {
	m := NewStringMap[int]()

	// empty map
	if got, want := m.Len(), 0; got != want {
		t.Errorf("Len() = %d, want %d", got, want)
	}

	if got, ok := m.Get("missing"); ok || got != 0 {
		t.Errorf("Get(missing) = %v,%v; want zero,false", got, ok)
	}

	// set + get
	m.Set("key1", 42)
	if got, want := m.Len(), 1; got != want {
		t.Errorf("Len after Set = %d, want %d", got, want)
	}
	if got, ok := m.Get("key1"); !ok || got != 42 {
		t.Errorf("Get(key1) = %v,%v; want 42,true", got, ok)
	}

	// overwrite
	m.Set("key1", 99)
	if got, ok := m.Get("key1"); !ok || got != 99 {
		t.Error("overwrite failed")
	}

	// delete
	m.Delete("key1")
	if got, want := m.Len(), 0; got != want {
		t.Errorf("Len after Delete = %d, want %d", got, want)
	}
	if _, ok := m.Get("key1"); ok {
		t.Error("deleted key still present")
	}

	// Clear
	m.Set("a", 1)
	m.Set("b", 2)
	m.Clear()
	if got, want := m.Len(), 0; got != want {
		t.Errorf("Len after Clear = %d, want %d", got, want)
	}
}

func TestMap_Range(t *testing.T) {
	m := NewStringMap[int]()
	data := map[string]int{"a": 10, "b": 20, "c": 30}
	for k, v := range data {
		m.Set(k, v)
	}

	seen := maps.Collect(m.Range)

	if got, want := len(seen), len(data); got != want {
		t.Errorf("Range saw %d keys, want %d", got, want)
	}
	for k, wantV := range data {
		if got, ok := seen[k]; !ok || got != wantV {
			t.Errorf("key %s: got %d, want %d", k, got, wantV)
		}
	}

	// early stop
	count := 0
	for range m.Range {
		count++
		if count == 2 {
			break
		}
	}
	if got, want := count, 2; got != want {
		t.Errorf("early stop count = %d, want %d", got, want)
	}
}

func TestMap_Concurrency(t *testing.T) {
	m := NewStringMap[int]()
	const goroutines = 64
	const ops = 200

	wg := sync.WaitGroup{}
	for id := range goroutines {
		wg.Go(func() {
			for j := range ops {
				key := strconv.Itoa(id*ops + j)
				m.Set(key, id)
				if _, ok := m.Get(key); !ok {
					t.Error("get after set failed")
				}
				if j%7 == 0 {
					m.Delete(key)
				}
			}
		})
	}
	wg.Wait()

	// final len rough check (deletes make exact hard)
	if got := m.Len(); got > goroutines*ops || got < 0 {
		t.Errorf("final Len = %d out of range", got)
	}
}

func TestMap_Stress(t *testing.T) {
	m := NewStringMap[int]()
	const workers = 32
	const keys = 10000

	wg := sync.WaitGroup{}
	for i := range workers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for k := range keys {
				key := strconv.Itoa(workerID*keys + k)
				m.Set(key, k)
				if _, ok := m.Get(key); !ok {
					t.Error("stress get failed")
					return
				}
				if k%10 == 0 {
					m.Delete(key) // occasional delete
				}
			}
		}(i)
	}
	wg.Wait()

	if m.Len() == 0 {
		t.Error("stress test ended with empty map")
	}
}
