package set

import "testing"

var sink int // 防止编译器优化掉被测操作

func benchN() int { return 1000 }

func BenchmarkHashSetAdd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := New[int]()
		for j := 0; j < 1000; j++ {
			s.Add(j)
		}
		sink = s.Len()
	}
}

func BenchmarkHashSetContains(b *testing.B) {
	s := New[int]()
	for j := 0; j < 1000; j++ {
		s.Add(j)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s.Contains(i % 1000) {
			sink++
		}
	}
}

func BenchmarkHashSetUnion(b *testing.B) {
	a := New[int]()
	c := New[int]()
	for j := 0; j < 1000; j++ {
		a.Add(j)
		if j%2 == 0 {
			c.Add(j)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = a.Union(c).Len()
	}
}

func BenchmarkHashSetIntersection(b *testing.B) {
	a := New[int]()
	c := New[int]()
	for j := 0; j < 1000; j++ {
		a.Add(j)
		if j%2 == 0 {
			c.Add(j)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = a.Intersection(c).Len()
	}
}

func BenchmarkHashSetIterate(b *testing.B) {
	s := New[int]()
	for j := 0; j < 1000; j++ {
		s.Add(j)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := 0
		for range s.All() {
			n++
		}
		sink = n
	}
}

func BenchmarkSortedSetElements(b *testing.B) {
	s := NewOrderedSet[int]()
	for j := 0; j < 1000; j++ {
		s.Add(999 - j)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = len(s.Elements())
	}
}

func BenchmarkSortedSetAdd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := NewOrderedSet[int]()
		for j := 0; j < 1000; j++ {
			s.Add(j)
		}
		sink = s.Len()
	}
}

func BenchmarkSafeSetAdd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := NewSafeSet[int]()
		for j := 0; j < 1000; j++ {
			s.Add(j)
		}
		sink = s.Len()
	}
}

func BenchmarkImmutableSetAdd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := NewImmutable[int]()
		for j := 0; j < 100; j++ {
			s = s.Add(j) // 每次返回新集合
		}
		sink = s.Len()
	}
}

// --- 对照基准：原生 map[T]struct{} 基线，衡量 HashSet 的包装开销 ---

func BenchmarkRawMapContains(b *testing.B) {
	m := make(map[int]struct{}, 1000)
	for j := 0; j < 1000; j++ {
		m[j] = struct{}{}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := m[i%1000]; ok {
			sink++
		}
	}
}

func BenchmarkRawMapAdd(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := make(map[int]struct{})
		for j := 0; j < 1000; j++ {
			m[j] = struct{}{}
		}
		sink = len(m)
	}
}

// --- 批量路径对照：一次 Add/Remove 大量元素 vs 逐个 ---
// 逐个 Add 是 O(n²)（每次 O(n) 移动）；批量路径是 O(n + m log m) 归并。

func BenchmarkSortedSetAddBatch(b *testing.B) {
	elems := make([]int, 1000)
	for j := range elems {
		elems[j] = 999 - j
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := NewOrderedSet[int]()
		s.Add(elems...) // 一次传入 1000 个
		sink = s.Len()
	}
}

func BenchmarkSortedSetRemoveBatch(b *testing.B) {
	all := make([]int, 1000)
	for j := range all {
		all[j] = j
	}
	s := NewOrderedSet(all...)
	drop := make([]int, 500)
	for j := range drop {
		drop[j] = j * 2 // 删除一半
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss := s.Clone()
		ss.Remove(drop...)
		sink = ss.Len()
	}
}

func BenchmarkSafeSetUnion(b *testing.B) {
	a := NewSafeSet[int]()
	c := NewSafeSet[int]()
	for j := 0; j < 1000; j++ {
		a.Add(j)
		if j%2 == 0 {
			c.Add(j)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = a.Union(c).Len()
	}
}

// --- 复杂度验证：SortedSet.Contains 的二分对数增长 ---
// 10000 元素应约为 1000 元素的 log10 倍耗时（约 1.33x），而不是线性 10x。

func BenchmarkSortedSetContains1K(b *testing.B) {
	s := NewOrderedSet[int]()
	for j := 0; j < 1000; j++ {
		s.Add(j)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s.Contains(i % 1000) {
			sink++
		}
	}
}

func BenchmarkSortedSetContains10K(b *testing.B) {
	s := NewOrderedSet[int]()
	for j := 0; j < 10000; j++ {
		s.Add(j)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if s.Contains(i % 10000) {
			sink++
		}
	}
}
