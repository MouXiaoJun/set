package set

import (
	"slices"
	"testing"
)

func TestNewAndLen(t *testing.T) {
	s := New[int]()
	if s.Len() != 0 || !s.IsEmpty() {
		t.Fatalf("empty set: Len=%d IsEmpty=%v", s.Len(), s.IsEmpty())
	}

	s = New(1, 2, 3)
	if s.Len() != 3 {
		t.Fatalf("Len = %d, want 3", s.Len())
	}
}

func TestZeroValueUsable(t *testing.T) {
	var s HashSet[int] // 零值可用
	s.Add(1, 2)
	if !s.Contains(1) || !s.Contains(2) || s.Len() != 2 {
		t.Fatalf("zero-value set broken: %s", s.String())
	}
	s.Remove(1)
	if s.Contains(1) || s.Len() != 1 {
		t.Fatalf("remove from zero-value set broken: %s", s.String())
	}
}

func TestAddRemoveContains(t *testing.T) {
	s := New[string]()
	s.Add("a", "b", "a") // 幂等
	if s.Len() != 2 {
		t.Fatalf("Len = %d, want 2", s.Len())
	}
	if !s.Contains("a") || !s.Contains("b") || s.Contains("c") {
		t.Fatalf("Contains wrong: %s", s.String())
	}
	s.Remove("b", "not-exist") // 删除不存在是 no-op
	if s.Contains("b") || s.Len() != 1 {
		t.Fatalf("Remove wrong: %s", s.String())
	}
	s.Clear()
	if s.Len() != 0 || !s.IsEmpty() {
		t.Fatalf("Clear failed: %s", s.String())
	}
}

func TestCloneIsolation(t *testing.T) {
	a := New(1, 2, 3)
	b := a.Clone()
	b.Add(4)
	a.Remove(1)
	if a.Contains(4) || !b.Contains(1) {
		t.Fatalf("Clone not isolated: a=%s b=%s", a, b)
	}
	if a.Len() != 2 || b.Len() != 4 {
		t.Fatalf("Len mismatch: a=%d b=%d", a.Len(), b.Len())
	}
}

func TestElements(t *testing.T) {
	s := New(3, 1, 2)
	elems := s.Elements()
	if len(elems) != 3 {
		t.Fatalf("len = %d, want 3", len(elems))
	}
	for _, e := range []int{1, 2, 3} {
		if !slices.Contains(elems, e) {
			t.Fatalf("Elements missing %d: %v", e, elems)
		}
	}
}

func TestForEach(t *testing.T) {
	s := New(1, 2, 3, 4)
	var got []int
	s.ForEach(func(e int) bool {
		got = append(got, e)
		return e != 2 // 遇到 2 提前终止
	})
	// map 遍历顺序随机，但早停语义可验证：2 必须是最后一个被访问的元素，且只出现一次。
	if !slices.Contains(got, 2) || got[len(got)-1] != 2 {
		t.Fatalf("ForEach early-stop wrong: %v", got)
	}
}

func TestAllIter(t *testing.T) {
	s := New("a", "b", "c")
	var got []string
	for v := range s.All() {
		got = append(got, v)
	}
	if len(got) != 3 {
		t.Fatalf("iter len = %d, want 3", len(got))
	}
	for _, e := range []string{"a", "b", "c"} {
		if !slices.Contains(got, e) {
			t.Fatalf("iter missing %q: %v", e, got)
		}
	}
}

func TestStringStable(t *testing.T) {
	s := New("b", "a", "c")
	first := s.String()
	for i := 0; i < 10; i++ {
		if s.String() != first {
			t.Fatalf("String unstable: %q vs %q", first, s.String())
		}
	}
	if first != "{a, b, c}" {
		t.Fatalf("String = %q, want {a, b, c}", first)
	}
}

func TestUnion(t *testing.T) {
	a := New(1, 2, 3)
	b := New(3, 4)
	u := a.Union(b)
	if !u.Equal(New(1, 2, 3, 4)) || u.Len() != 4 {
		t.Fatalf("Union wrong: %s", u)
	}
	// 原集合未被修改
	if a.Len() != 3 || b.Len() != 2 {
		t.Fatalf("Union mutated input: a=%d b=%d", a.Len(), b.Len())
	}
}

func TestIntersection(t *testing.T) {
	a := New(1, 2, 3, 4)
	b := New(3, 4, 5)
	i := a.Intersection(b)
	if !i.Equal(New(3, 4)) {
		t.Fatalf("Intersection wrong: %s", i)
	}
	if !a.Intersection(New(9)).IsEmpty() {
		t.Fatalf("disjoint intersection should be empty")
	}
}

func TestDifference(t *testing.T) {
	a := New(1, 2, 3, 4)
	b := New(2, 4, 6)
	d := a.Difference(b)
	if !d.Equal(New(1, 3)) {
		t.Fatalf("Difference wrong: %s", d)
	}
}

func TestSymmetricDifference(t *testing.T) {
	a := New(1, 2, 3)
	b := New(3, 4, 5)
	sd := a.SymmetricDifference(b)
	if !sd.Equal(New(1, 2, 4, 5)) {
		t.Fatalf("SymmetricDifference wrong: %s", sd)
	}
}

func TestSubsetSuperset(t *testing.T) {
	a := New(1, 2)
	b := New(1, 2, 3)
	empty := New[int]()
	if !a.IsSubset(b) || !b.IsSuperset(a) {
		t.Fatalf("subset/superset wrong")
	}
	if b.IsSubset(a) || a.IsSuperset(b) {
		t.Fatalf("reverse subset/superset wrong")
	}
	if !empty.IsSubset(a) {
		t.Fatalf("empty set should be subset of everything")
	}
	if !a.IsSubset(a) {
		t.Fatalf("set should be subset of itself")
	}
}

func TestEqual(t *testing.T) {
	if !New(1, 2, 3).Equal(New(3, 2, 1)) {
		t.Fatalf("order-independent Equal failed")
	}
	if New(1, 2).Equal(New(1, 2, 3)) {
		t.Fatalf("different-size Equal failed")
	}
}

func TestIsDisjoint(t *testing.T) {
	if !New(1, 2).IsDisjoint(New(3, 4)) {
		t.Fatalf("disjoint sets reported overlapping")
	}
	if New(1, 2).IsDisjoint(New(2, 3)) {
		t.Fatalf("overlapping sets reported disjoint")
	}
}

func TestAddAllRemoveAllRetainAll(t *testing.T) {
	a := New(1, 2)
	b := New(2, 3, 4)

	a.AddAll(b)
	if !a.Equal(New(1, 2, 3, 4)) {
		t.Fatalf("AddAll wrong: %s", a)
	}

	a.RemoveAll(b)
	if !a.Equal(New(1)) {
		t.Fatalf("RemoveAll wrong: %s", a)
	}

	a = New(1, 2, 3, 4)
	a.RetainAll(New(2, 4, 6))
	if !a.Equal(New(2, 4)) {
		t.Fatalf("RetainAll wrong: %s", a)
	}
}

func TestChainCalls(t *testing.T) {
	s := New(1)
	s.Add(2).Remove(3) // 幂等链式
	if !s.Equal(New(1, 2)) {
		t.Fatalf("chain wrong: %s", s)
	}
}

func TestCollectSeq(t *testing.T) {
	s := CollectSeq(slices.Values([]int{1, 2, 2, 3}))
	if !s.Equal(New(1, 2, 3)) {
		t.Fatalf("CollectSeq wrong: %s", s)
	}
}

func TestStructElements(t *testing.T) {
	type point struct{ x, y int }
	s := New(point{1, 2}, point{3, 4})
	if !s.Contains(point{1, 2}) || s.Len() != 2 {
		t.Fatalf("struct element set broken: %s", s)
	}
}
