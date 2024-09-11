package hashmap

import (
	"testing"
)

func TestHashMap(t *testing.T) {
	m := NewHashMap[int, string]()
	m.Set(1, "one")
	v, ok := m.Get(1)
	if !ok || v != "one" {
		t.Errorf("expected value to be 'one', got %s", v)
	}
	m.Delete(1)
	_, ok = m.Get(1)
	if ok {
		t.Errorf("expected value to be deleted")
	}
}

func BenchmarkHashMap(b *testing.B) {
	m := NewHashMap[int, string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(i, "value")
	}
}

func BenchmarkShardMap(b *testing.B) {
	m := NewShardMap[int, string]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(i, "value")
	}
}
