package set

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

// ImmutableSet 是不可变集合：一旦创建，底层 map 永不修改。
// 所有"修改"操作都返回新集合，且通过结构共享避免整表复制：
//   - Add 时若元素已全部存在，直接返回原集合（共享同一份 map）；
//   - Remove 时若元素都不存在，同样返回原集合；
//   - 真正变化时才复制旧 map 并增删元素（copy-on-write）。
//
// 因此 ImmutableSet 天然并发安全，适合做配置快照、函数式链式运算、
// 跨 goroutine 共享的只读集合。代价是频繁"变式"会产生复制，不适合高频写入。
type ImmutableSet[T comparable] struct {
	m map[T]struct{}
}

// NewImmutable 创建不可变集合。
func NewImmutable[T comparable](elems ...T) *ImmutableSet[T] {
	s := &ImmutableSet[T]{m: make(map[T]struct{}, len(elems))}
	for _, e := range elems {
		s.m[e] = struct{}{}
	}
	return s
}

// Freeze 把可变 HashSet 转成不可变集合（复制底层 map，与原集合解耦）。
func (s *HashSet[T]) Freeze() *ImmutableSet[T] {
	return &ImmutableSet[T]{m: s.Clone().m}
}

// Thaw 把不可变集合转回可变 HashSet（复制底层 map，与原集合解耦）。
func (s *ImmutableSet[T]) Thaw() *HashSet[T] {
	hs := New[T]()
	for e := range s.m {
		hs.m[e] = struct{}{}
	}
	return hs
}

// Add 返回一个包含 elems 的新集合；元素已全部存在时返回原集合（结构共享）。
func (s *ImmutableSet[T]) Add(elems ...T) *ImmutableSet[T] {
	for _, e := range elems {
		if _, ok := s.m[e]; !ok {
			return s.addChanged(elems)
		}
	}
	return s
}

// addChanged 仅在确有新元素时调用：复制旧 map 并写入。
func (s *ImmutableSet[T]) addChanged(elems []T) *ImmutableSet[T] {
	m := make(map[T]struct{}, len(s.m)+len(elems))
	for k := range s.m {
		m[k] = struct{}{}
	}
	for _, e := range elems {
		m[e] = struct{}{}
	}
	return &ImmutableSet[T]{m: m}
}

// Remove 返回删掉 elems 的新集合；元素都不存在时返回原集合（结构共享）。
func (s *ImmutableSet[T]) Remove(elems ...T) *ImmutableSet[T] {
	for _, e := range elems {
		if _, ok := s.m[e]; ok {
			return s.removeChanged(elems)
		}
	}
	return s
}

// removeChanged 仅在确有元素被删时调用。
func (s *ImmutableSet[T]) removeChanged(elems []T) *ImmutableSet[T] {
	m := make(map[T]struct{}, len(s.m))
	for k := range s.m {
		m[k] = struct{}{}
	}
	for _, e := range elems {
		delete(m, e)
	}
	return &ImmutableSet[T]{m: m}
}

// Contains 判断是否包含 elem。
func (s *ImmutableSet[T]) Contains(elem T) bool {
	_, ok := s.m[elem]
	return ok
}

// Len 返回元素个数。
func (s *ImmutableSet[T]) Len() int {
	return len(s.m)
}

// IsEmpty 集合是否为空。
func (s *ImmutableSet[T]) IsEmpty() bool {
	return len(s.m) == 0
}

// Clear 返回空集合（不可变，不会修改原集合）。
func (s *ImmutableSet[T]) Clear() *ImmutableSet[T] {
	return NewImmutable[T]()
}

// Elements 返回所有元素切片（顺序不保证稳定）。
func (s *ImmutableSet[T]) Elements() []T {
	out := make([]T, 0, len(s.m))
	for e := range s.m {
		out = append(out, e)
	}
	return out
}

// ForEach 遍历集合，fn 返回 false 可提前终止。
func (s *ImmutableSet[T]) ForEach(fn func(T) bool) {
	for e := range s.m {
		if !fn(e) {
			return
		}
	}
}

// All 返回迭代器（range-over-func）。
func (s *ImmutableSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range s.m {
			if !yield(e) {
				return
			}
		}
	}
}

// String 实现 fmt.Stringer，有序输出，便于调试与断言。
func (s *ImmutableSet[T]) String() string {
	elems := s.Elements()
	sort.Slice(elems, func(i, j int) bool {
		return fmt.Sprint(elems[i]) < fmt.Sprint(elems[j])
	})
	var b strings.Builder
	b.WriteByte('{')
	for i, e := range elems {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, e)
	}
	b.WriteByte('}')
	return b.String()
}

// Union 返回并集。
func (s *ImmutableSet[T]) Union(other *ImmutableSet[T]) *ImmutableSet[T] {
	m := make(map[T]struct{}, len(s.m)+len(other.m))
	for e := range s.m {
		m[e] = struct{}{}
	}
	for e := range other.m {
		m[e] = struct{}{}
	}
	return &ImmutableSet[T]{m: m}
}

// Intersection 返回交集。
func (s *ImmutableSet[T]) Intersection(other *ImmutableSet[T]) *ImmutableSet[T] {
	a, b := s, other
	if len(a.m) > len(b.m) {
		a, b = b, a
	}
	m := make(map[T]struct{})
	for e := range a.m {
		if _, ok := b.m[e]; ok {
			m[e] = struct{}{}
		}
	}
	return &ImmutableSet[T]{m: m}
}

// Difference 返回差集：属于 s 但不属于 other。
func (s *ImmutableSet[T]) Difference(other *ImmutableSet[T]) *ImmutableSet[T] {
	m := make(map[T]struct{})
	for e := range s.m {
		if _, ok := other.m[e]; !ok {
			m[e] = struct{}{}
		}
	}
	return &ImmutableSet[T]{m: m}
}

// SymmetricDifference 返回对称差集。
func (s *ImmutableSet[T]) SymmetricDifference(other *ImmutableSet[T]) *ImmutableSet[T] {
	m := make(map[T]struct{})
	for e := range s.m {
		if _, ok := other.m[e]; !ok {
			m[e] = struct{}{}
		}
	}
	for e := range other.m {
		if _, ok := s.m[e]; !ok {
			m[e] = struct{}{}
		}
	}
	return &ImmutableSet[T]{m: m}
}

// IsSubset 判断 s ⊆ other。
func (s *ImmutableSet[T]) IsSubset(other *ImmutableSet[T]) bool {
	if len(s.m) > len(other.m) {
		return false
	}
	for e := range s.m {
		if _, ok := other.m[e]; !ok {
			return false
		}
	}
	return true
}

// IsSuperset 判断 s ⊇ other。
func (s *ImmutableSet[T]) IsSuperset(other *ImmutableSet[T]) bool {
	return other.IsSubset(s)
}

// Equal 判断两个集合元素是否完全相同。
func (s *ImmutableSet[T]) Equal(other *ImmutableSet[T]) bool {
	if len(s.m) != len(other.m) {
		return false
	}
	return s.IsSubset(other)
}

// IsDisjoint 判断两个集合是否不相交。
func (s *ImmutableSet[T]) IsDisjoint(other *ImmutableSet[T]) bool {
	a, b := s, other
	if len(a.m) > len(b.m) {
		a, b = b, a
	}
	for e := range a.m {
		if _, ok := b.m[e]; ok {
			return false
		}
	}
	return true
}
