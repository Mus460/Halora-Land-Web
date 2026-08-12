package cache

import (
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := New(time.Minute)
	c.Set("k", "v")
	v, ok := c.Get("k")
	if !ok {
		t.Fatal("expected hit")
	}
	if v != "v" {
		t.Errorf("got %v want v", v)
	}
}

func TestGetMiss(t *testing.T) {
	c := New(time.Minute)
	if _, ok := c.Get("nope"); ok {
		t.Fatal("expected miss for missing key")
	}
}

func TestExpiry(t *testing.T) {
	c := New(20 * time.Millisecond)
	c.Set("k", "v")
	time.Sleep(40 * time.Millisecond)
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected miss after expiry")
	}
}

func TestClear(t *testing.T) {
	c := New(time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected a cleared")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("expected b cleared")
	}
}

func TestConcurrentUse(t *testing.T) {
	c := New(time.Minute)
	done := make(chan struct{})
	for g := 0; g < 8; g++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for i := 0; i < 200; i++ {
				c.Set("k", i)
				c.Get("k")
				c.Clear()
			}
		}()
	}
	for g := 0; g < 8; g++ {
		<-done
	}
}

func TestExpiredEntryDeleted(t *testing.T) {
	c := New(10 * time.Millisecond)
	c.Set("k", "v")
	time.Sleep(30 * time.Millisecond)
	_, _ = c.Get("k")
	c.mu.Lock()
	if _, exists := c.items["k"]; exists {
		t.Error("expired entry should be removed from the map on Get")
	}
	c.mu.Unlock()
}
