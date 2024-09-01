package trie

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSegmenter(t *testing.T) {
	tests := []struct {
		domain string
		start  int
		seg    string
		next   int
		expect bool
	}{
		{
			domain: "example.com",
			start:  len("example.com"),
			seg:    "com",
			next:   7,
			expect: true,
		},
		{
			domain: "example.com",
			start:  0,
			seg:    "",
			next:   -1,
			expect: true,
		},
		{
			domain: "example.com",
			start:  20,
			seg:    "",
			next:   -1,
			expect: true,
		},
		{
			domain: "example.com",
			start:  7,
			seg:    "example",
			next:   -1,
			expect: true,
		},
		{
			domain: "",
			start:  7,
			seg:    "",
			next:   -1,
			expect: true,
		},
	}

	for _, test := range tests {
		seg, next := segmenter(test.domain, test.start)
		assert.Equal(t, test.seg, seg)
		assert.Equal(t, test.next, next)
	}
	key := "+.example.com"
	for part, i := segmenter(key, len(key)); part != ""; part, i = segmenter(key, i) {
		t.Log(part)
	}
}

func TestDomainTrie(t *testing.T) {
	tree := NewDomainTrie[int]()
	var np = func(n int) *int {
		return &n
	}

	tests := []struct {
		domain string
		val    *int
	}{
		{
			domain: "example.com",
			val:    np(1),
		},
		{
			domain: "*.c.example.com",
			val:    np(2),
		},
		{
			domain: "www.example.com",
			val:    np(3),
		},
		{
			domain: "sub.*.example.com",
			val:    np(4),
		},
		{
			domain: "*.sub.c.example.com",
			val:    np(5),
		},
	}
	for _, test := range tests {
		tree.Insert(test.domain, *test.val)
	}
	tree.Print()

	tests = []struct {
		domain string
		val    *int
	}{
		{
			domain: "example.com",
			val:    np(1),
		},
		{
			domain: "a.example.com",
			val:    nil,
		},
		{
			domain: "b.c.example.com",
			val:    np(2),
		},
		{
			domain: "b.b.example.com",
			val:    nil,
		},
		{
			domain: "www.example.com",
			val:    np(3),
		},
		{
			domain: "sub.a.example.com",
			val:    np(4),
		},
		{
			domain: "sub.a.example.com",
			val:    np(4),
		},
		{
			domain: "sub.a.b.a.example.com",
			val:    np(4),
		},
		{
			domain: "bb.sub.c.example.com",
			val:    np(5),
		},
	}
	for _, test := range tests {
		node := tree.Search(test.domain)
		if test.val == nil {
			assert.Zero(t, node, test.domain)
			continue
		}
		assert.NotZero(t, node, test.domain)
		assert.Equal(t, test.val, &node.data, test.domain)
	}
}

func BenchmarkDomainTrieInsert(b *testing.B) {
	tree := NewDomainTrie[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Insert("example.com", i)
	}
}

func BenchmarkDomainTrieSearch(b *testing.B) {
	tree := NewDomainTrie[int]()
	tests := []struct {
		domain string
		val    int
	}{
		{
			domain: "example.com",
			val:    1,
		},
		{
			domain: "*.example.com",
			val:    2,
		},
		{
			domain: "www.example.com",
			val:    3,
		},
		{
			domain: "sub.*.example.com",
			val:    4,
		},
	}
	for _, test := range tests {
		tree.Insert(test.domain, test.val)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Search("sub.a.example.com")
	}
}
