package set

import (
	"slices"
	"testing"
)

func TestOrderedSetSorted(t *testing.T) {
	s := NewOrderedSet(5, 1, 3, 2, 4)
	want := []int{1, 2, 3, 4, 5}
	if got := s.Elements(); !slices.Equal(got, want) {
		t.Fatalf("Elements = %v, want %v", got, want)
	}

	var iterated []int
	for v := range s.All() {
		iterated = append(iterated, v)
	}
	if !slices.Equal(iterated, want) {
		t.Fatalf("All iter = %v, want %v", iterated, want)
	}
}

func TestOrderedSetMutationStaysSorted(t *testing.T) {
	s := NewOrderedSet(2, 1)
	if got := s.Elements(); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("first read = %v", got)
	}
	s.Add(0)
	s.Add(9)
	s.Remove(9)
	s.Add(1) // 已存在，幂等
	if got := s.Elements(); !slices.Equal(got, []int{0, 1, 2}) {
		t.Fatalf("after mutation = %v", got)
	}
	if !s.Contains(0) || !s.Contains(2) || s.Contains(9) {
		t.Fatalf("Contains wrong after mutation")
	}
}

func TestOrderedSetStrings(t *testing.T) {
	s := NewOrderedSet("banana", "apple", "cherry")
	want := []string{"apple", "banana", "cherry"}
	if got := s.Elements(); !slices.Equal(got, want) {
		t.Fatalf("Elements = %v, want %v", got, want)
	}
}

func TestCustomComparator(t *testing.T) {
	// 按长度升序的自定义比较器；等长时按字典序（保证全序）
	byLen := NewSortedSet(func(a, b string) int {
		if c := len(a) - len(b); c != 0 {
			return c
		}
		return slices.Compare([]byte(a), []byte(b))
	}, "bb", "a", "ccc", "cc")
	if got := byLen.Elements(); !slices.Equal(got, []string{"a", "bb", "cc", "ccc"}) {
		t.Fatalf("custom cmp Elements = %v", got)
	}
}

// TestNonComparableElements 验证 SortedSet 支持无法作为 map key 的类型（含 slice 字段）。
func TestNonComparableElements(t *testing.T) {
	type rec struct {
		name string
		tags []string
	}
	byName := NewSortedSet(func(a, b rec) int {
		return slices.Compare([]byte(a.name), []byte(b.name))
	}, rec{"bob", []string{"x"}}, rec{"alice", []string{"y"}})
	if got := byName.Elements(); got[0].name != "alice" || got[1].name != "bob" {
		t.Fatalf("non-comparable order wrong: %+v", got)
	}
	if !byName.Contains(rec{"bob", []string{"z"}}) { // tags 不同但 name 相等 → 判定存在
		t.Fatalf("comparator-based Contains wrong")
	}
}

func TestMinMax(t *testing.T) {
	s := NewOrderedSet(3, 1, 2)
	if v, ok := s.Min(); !ok || v != 1 {
		t.Fatalf("Min = %d,%v", v, ok)
	}
	if v, ok := s.Max(); !ok || v != 3 {
		t.Fatalf("Max = %d,%v", v, ok)
	}
	var empty SortedSet[int]
	if _, ok := empty.Min(); ok {
		t.Fatalf("empty Min should be !ok")
	}
	if _, ok := empty.Max(); ok {
		t.Fatalf("empty Max should be !ok")
	}
}

func TestRangeQueries(t *testing.T) {
	s := NewOrderedSet(1, 3, 5, 7, 9)

	if v, ok := s.Lower(5); !ok || v != 3 {
		t.Fatalf("Lower(5) = %d,%v, want 3,true", v, ok)
	}
	if v, ok := s.Lower(1); ok {
		t.Fatalf("Lower(1) should be !ok, got %d", v)
	}
	if v, ok := s.Floor(5); !ok || v != 5 {
		t.Fatalf("Floor(5) = %d,%v, want 5,true", v, ok)
	}
	if v, ok := s.Floor(4); !ok || v != 3 {
		t.Fatalf("Floor(4) = %d,%v, want 3,true", v, ok)
	}
	if v, ok := s.Ceiling(6); !ok || v != 7 {
		t.Fatalf("Ceiling(6) = %d,%v, want 7,true", v, ok)
	}
	if v, ok := s.Ceiling(5); !ok || v != 5 {
		t.Fatalf("Ceiling(5) = %d,%v, want 5,true", v, ok)
	}
	if v, ok := s.Higher(7); !ok || v != 9 {
		t.Fatalf("Higher(7) = %d,%v, want 9,true", v, ok)
	}
	if v, ok := s.Higher(9); ok {
		t.Fatalf("Higher(9) should be !ok, got %d", v)
	}
}

func TestSortedSetOps(t *testing.T) {
	a := NewOrderedSet(1, 2, 3)
	b := NewOrderedSet(2, 3, 4)

	if got := a.Union(b).Elements(); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("Union = %v", got)
	}
	if got := a.Intersection(b).Elements(); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("Intersection = %v", got)
	}
	if got := a.Difference(b).Elements(); !slices.Equal(got, []int{1}) {
		t.Fatalf("Difference = %v", got)
	}
	if got := a.SymmetricDifference(b).Elements(); !slices.Equal(got, []int{1, 4}) {
		t.Fatalf("SymmetricDifference = %v", got)
	}
	if !a.IsSubset(NewOrderedSet(1, 2, 3, 4)) {
		t.Fatalf("a should be subset of {1,2,3,4}")
	}
	if !NewOrderedSet(1, 2, 3, 4).IsSuperset(a) {
		t.Fatalf("superset check failed")
	}
	if !a.Equal(NewOrderedSet(3, 2, 1)) {
		t.Fatalf("Equal failed")
	}
	// 运算不修改输入
	if got := a.Elements(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("ops mutated input: %v", got)
	}
}

func TestSortedSetCustomType(t *testing.T) {
	// 自定义类型：按 ID 排序
	type user struct {
		id   int
		name string
	}
	byID := NewSortedSet(func(a, b user) int { return a.id - b.id },
		user{2, "b"}, user{1, "a"}, user{3, "c"})
	got := byID.Elements()
	if got[0].id != 1 || got[1].id != 2 || got[2].id != 3 {
		t.Fatalf("custom type order wrong: %+v", got)
	}
	if v, ok := byID.Max(); !ok || v.id != 3 {
		t.Fatalf("Max = %+v,%v", v, ok)
	}
}
