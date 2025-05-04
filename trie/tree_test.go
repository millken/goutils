package trie

import (
	"testing"
)

func TestTree(t *testing.T) {
	// 创建域名树
	tree := NewDomainTree[string]()

	// 添加域名配置
	tree.Add("example.com", "主域名")
	tree.Add("www.example.com", "网站")
	tree.Add("api.example.com", "API服务器")
	tree.Add("cdn.example.com", "CDN主站")
	tree.Add("*.cdn.example.com", "CDN服务器")
	tree.Add("a1.cdn.example.com", "CDN服务器-A1")
	tree.Add("a1.a3.a4.cdn.example.com", "CDN服务器-A1-A3-A4")
	tests := []struct {
		domain string
		found  bool
		val    string
	}{
		{
			domain: "example.com",
			found:  true,
			val:    "主域名",
		},
		{
			domain: "www.example.com",
			found:  true,
			val:    "网站",
		},
		{
			domain: "api.example.com",
			found:  true,
			val:    "API服务器",
		},
		{
			domain: "api-1.example.com",
			found:  false,
			val:    "",
		},
		{
			domain: "cdn.example.com",
			found:  true,
			val:    "CDN主站",
		},
		{
			domain: "1.cdn.example.com",
			found:  true,
			val:    "CDN服务器",
		},
		{
			domain: "2.1.cdn.example.com",
			found:  true,
			val:    "CDN服务器",
		},
		{
			domain: "a1.cdn.example.com",
			found:  true,
			val:    "CDN服务器-A1",
		},
		{
			domain: "a1.cdn1.example.com",
			found:  false,
			val:    "",
		},
		{
			domain: "a1.a3.a4.cdn.example.com",
			found:  true,
			val:    "CDN服务器-A1-A3-A4",
		},
	}
	for _, test := range tests {
		val, found := tree.Lookup(test.domain)
		if found == test.found && val == test.val {
			// fmt.Printf("域名 %s 查找成功，值为: %s\n", test.domain, val)
		} else {
			t.Fatalf("域名 %s 查找失败，期望值: %s, 实际值: %s\n", test.domain, test.val, val)
		}
	}
	tree.Print()
	// // 获取所有域名配置
	// all := tree.All()
	// fmt.Println("全部域名配置:")
	// for domain, config := range all {
	// 	fmt.Printf("%s: %s\n", domain, config)
	// }
}

func TestDomainTreeDelete(t *testing.T) {
	tree := NewDomainTree[int]()
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
			domain: "a1.a2.a3.example.com",
			val:    3,
		},
	}
	for _, test := range tests {
		tree.Add(test.domain, test.val)
	}

	tree.Delete("*.example.com")
	if _, found := tree.Lookup("asdf.example.com"); found {
		t.Fatalf("删除失败，域名 *.example.com 仍然存在")
	}
	// tree.Delete("example.com")
	// if _, found := tree.Lookup("example.com"); found {
	// 	t.Fatalf("删除失败，域名 www.example.com 不存在")
	// }
	tree.Delete("a1.a2.a3.example.com")
	if _, found := tree.Lookup("a1.a2.a3.example.com"); found {
		t.Fatalf("删除失败，域名 a1.a2.a3.example.com 仍然存在")
	}
}

func BenchmarkDomainTreeLookup(b *testing.B) {
	tree := NewDomainTree[int]()
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
			domain: "sub.a.example.com",
			val:    4,
		},
	}
	for _, test := range tests {
		tree.Add(test.domain, test.val)
	}

	for b.Loop() {
		tree.Lookup("sub.a.2.3.3.45.5.6.example.com")
	}
}
