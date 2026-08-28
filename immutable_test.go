package set

import (
	"testing"
)

func TestImmutableBasics(t *testing.T) {
	s := NewImmutable(1, 2, 3)
	if s.Len() != 3 || !s.Contains(1) || !s.Contains(3) || s.Contains(9) {
		t.Fatalf("basics wrong: %s", s)
	}
	if !s.IsEmpty() == (s.Len() == 0) {
		t.Fatalf("IsEmpty inconsistent")
	}
}

// TestImmutableStructuralSharing 验证结构共享：无变化时返回原集合，原集合永不被修改。
func TestImmutableStructuralSharing(t *testing.T) {
	s := NewImmutable(1, 2, 3)

	// 元素已存在：Add 应返回原集合（共享）
	if s.Add(1, 2) != s {
		t.Fatalf("Add of existing elems should share the same instance")
	}
	// 元素都不存在：Remove 应返回原集合
	if s.Remove(9, 10) != s {
		t.Fatalf("Remove of missing elems should share the same instance")
	}

	// 确实变化时返回新集合，原集合不变
	added := s.Add(4)
	if added == s || s.Contains(4) || !added.Contains(4) {
		t.Fatalf("Add(4) should return a new set, keep old untouched: s=%s added=%s", s, added)
	}
	removed := s.Remove(1)
	if removed == s || !s.Contains(1) || removed.Contains(1) {
		t.Fatalf("Remove(1) should return a new set, keep old untouched")
	}
}

func TestImmutableOps(t *testing.T) {
	a := NewImmutable(1, 2, 3)
	b := NewImmutable(3, 4)

	if u := a.Union(b); !u.Equal(NewImmutable(1, 2, 3, 4)) {
		t.Fatalf("Union wrong: %s", u)
	}
	if i := a.Intersection(b); !i.Equal(NewImmutable(3)) {
		t.Fatalf("Intersection wrong: %s", i)
	}
	if d := a.Difference(b); !d.Equal(NewImmutable(1, 2)) {
		t.Fatalf("Difference wrong: %s", d)
	}
	if sd := a.SymmetricDifference(b); !sd.Equal(NewImmutable(1, 2, 4)) {
		t.Fatalf("SymmetricDifference wrong: %s", sd)
	}
	if !a.IsSubset(NewImmutable(1, 2, 3, 4)) || !NewImmutable(1, 2, 3, 4).IsSuperset(a) {
		t.Fatalf("subset/superset wrong")
	}
	if !NewImmutable(1, 2, 3).Equal(NewImmutable(3, 2, 1)) {
		t.Fatalf("Equal wrong")
	}
	if !NewImmutable(1, 2).IsDisjoint(NewImmutable(3, 4)) {
		t.Fatalf("IsDisjoint wrong")
	}
}

// TestImmutableFreezeThaw 验证可变/不可变互转且解耦。
func TestImmutableFreezeThaw(t *testing.T) {
	hs := New(1, 2, 3)
	im := hs.Freeze()
	hs.Add(4) // 改原可变集合不影响不可变集合
	if im.Contains(4) {
		t.Fatalf("Freeze should isolate: %s", im)
	}

	back := im.Thaw()
	back.Add(5)
	if im.Contains(5) {
		t.Fatalf("Thaw should isolate: %s", im)
	}
	if !back.Equal(New(1, 2, 3, 5)) {
		t.Fatalf("Thaw content wrong: %s", back)
	}
}

func TestImmutableClear(t *testing.T) {
	s := NewImmutable(1, 2)
	c := s.Clear()
	if !c.IsEmpty() || !s.Equal(NewImmutable(1, 2)) {
		t.Fatalf("Clear should return empty without mutating: s=%s c=%s", s, c)
	}
}

func TestImmutableChain(t *testing.T) {
	// 函数式链式：每步都是新集合
	s := NewImmutable(1).
		Add(2, 3).
		Remove(1).
		Union(NewImmutable(4))
	if !s.Equal(NewImmutable(2, 3, 4)) {
		t.Fatalf("chain result wrong: %s", s)
	}
}

func TestImmutableIterAndForEach(t *testing.T) {
	s := NewImmutable("a", "b", "c")
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Fatalf("iter count = %d", count)
	}

	var visited []string
	s.ForEach(func(e string) bool {
		visited = append(visited, e)
		return false // 提前终止
	})
	if len(visited) != 1 {
		t.Fatalf("ForEach early-stop broken: %v", visited)
	}
}

func TestImmutableString(t *testing.T) {
	if got := NewImmutable("c", "a", "b").String(); got != "{a, b, c}" {
		t.Fatalf("String = %q", got)
	}
}
