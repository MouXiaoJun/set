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
