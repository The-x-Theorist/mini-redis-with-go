package main

import (
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	s := &Store{
		mu: sync.RWMutex{},
		data: make(map[string]StoreData),
	}

	s.Set("foo", "bar")
	val := s.Get("foo")

	if val != "bar" {
		t.Errorf("expected bar, got %s", val)
	}
}

func TestDel(t *testing.T) {
	s := &Store{
		mu: sync.RWMutex{},
		data: make(map[string]StoreData),
	}

	s.Set("foo", "bar")
	var val = s.Get("foo")

	if val != "bar" {
		t.Errorf("expected bar, got %s", val)
	}

	s.Del("foo")

	val = s.Get("foo")

	if val == "bar" {
		t.Error("Value was not deleted")
	}
}

func TestTTL (t *testing.T) {
	s := &Store{
		mu: sync.RWMutex{},
		data: make(map[string]StoreData),
	}

	s.mu.Lock()
	s.data["foo"] = StoreData{
		value: "bar",
		expiresAt: time.Now().Add(1 * time.Second),
	}
	s.mu.Unlock()

	if s.Get("foo") != "bar" {
		t.Errorf("Expected bar before expiry")
	}

	time.Sleep(2 * time.Second)

	if s.Get("foo") == "bar" {
		t.Error("Value didn't expire")
	}

	if s.Get("foo") != "ERR data doesn't exist" {
		t.Error("Data didn't expire")
	}
}

func TestSetAndGetCases(t *testing.T) {
	s := &Store{
		mu: sync.RWMutex{},
		data: make(map[string]StoreData),
	}

	tests := []struct{
		key string
		value string
	}{
		{"a", "1"},
		{"b", "2"},
		{"c", "3"},
	}

	for _, tc := range tests {
		s.Set(tc.key, tc.value)
		got := s.Get(tc.key)

		if got != tc.value {
			t.Error("Wrong value")
		}
	}
}

func TestConcurrency(t *testing.T) {
	s := &Store{
		mu: sync.RWMutex{},
		data: make(map[string]StoreData),
	}

	done := make(chan bool)

	for i := range 100 {
		go func(i int) {
			key := "k" + time.Now().String()
			s.Set(key, "value")
			_ = s.Get(key)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}