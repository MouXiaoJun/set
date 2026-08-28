// Package set 提供基于 Go 泛型的高性能集合类型。
//
// 本包零依赖，仅使用 Go 标准库，并面向 Go 1.23+ 的 iter.Seq 设计：
//   - HashSet[T comparable]     无序集合（默认，map 实现，零值可用）
//   - SortedSet[T any]          有序集合（自定义比较器，惰性排序缓存）
//   - SafeSet[T comparable]     线程安全集合（RWMutex 包装 HashSet）
//   - ImmutableSet[T comparable] 不可变集合（结构共享，每次运算返回新集合）
//
// 所有类型都实现了 All() iter.Seq[T]，支持 range-over-func：
//
//	for v := range s.All() { ... }
package set

import (
	"fmt"
	"iter"
	"sort"
	"strings"
)

// HashSet 是基于 map 的无序集合，T 必须是可比较类型。
// 底层用 map[T]struct{} 存元素：struct{} 零字节，不占用额外空间。
//
// 零值可用：var s HashSet[int] 可以直接 Add，首次写入时惰性初始化。
type HashSet[T comparable] struct {
	m map[T]struct{}
}

// New 创建包含 elems 的集合。
func New[T comparable](elems ...T) *HashSet[T] {
	s := &HashSet[T]{m: make(map[T]struct{}, len(elems))}
	for _, e := range elems {
		s.m[e] = struct{}{}
	}
	return s
}

// ensure 惰性初始化底层 map，保证零值可用。
func (s *HashSet[T]) ensure() {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
}

// Add 添加一个或多个元素（幂等）。返回 s 以支持链式调用。
func (s *HashSet[T]) Add(elems ...T) *HashSet[T] {
	s.ensure()
	for _, e := range elems {
		s.m[e] = struct{}{}
	}
	return s
}

// Remove 删除一个或多个元素；删除不存在的元素是安全的 no-op。返回 s 以支持链式调用。
func (s *HashSet[T]) Remove(elems ...T) *HashSet[T] {
	for _, e := range elems {
		delete(s.m, e)
	}
	return s
}

// Contains 判断集合是否包含 elem。
func (s *HashSet[T]) Contains(elem T) bool {
	_, ok := s.m[elem]
	return ok
}

// Len 返回元素个数。
func (s *HashSet[T]) Len() int {
	return len(s.m)
}

// IsEmpty 集合是否为空。
func (s *HashSet[T]) IsEmpty() bool {
	return len(s.m) == 0
}

// Clear 清空集合。
func (s *HashSet[T]) Clear() {
	clear(s.m)
}

// Clone 返回浅拷贝的新集合。
func (s *HashSet[T]) Clone() *HashSet[T] {
	clone := &HashSet[T]{m: make(map[T]struct{}, len(s.m))}
	for e := range s.m {
		clone.m[e] = struct{}{}
	}
	return clone
}

// Elements 返回所有元素组成的切片，顺序不保证稳定。
// 需要有序结果请用 SortedSet，或先 ToSlice 再自行排序。
func (s *HashSet[T]) Elements() []T {
	out := make([]T, 0, len(s.m))
	for e := range s.m {
		out = append(out, e)
	}
	return out
}

// ForEach 遍历集合，fn 返回 false 可提前终止。
func (s *HashSet[T]) ForEach(fn func(T) bool) {
	for e := range s.m {
		if !fn(e) {
			return
		}
	}
}

// String 实现 fmt.Stringer，按有序形式输出，保证结果稳定、便于调试。
func (s *HashSet[T]) String() string {
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

// All 返回一个迭代器，支持 range-over-func（遍历顺序不保证稳定）。
//
//	for v := range s.All() { ... }
func (s *HashSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range s.m {
			if !yield(e) {
				return
			}
		}
	}
}

// Union 返回并集：包含 s 与 other 的所有元素（两个集合均不被修改）。
func (s *HashSet[T]) Union(other *HashSet[T]) *HashSet[T] {
	out := New[T]()
	for e := range s.m {
		out.m[e] = struct{}{}
	}
	for e := range other.m {
		out.m[e] = struct{}{}
	}
	return out
}

// Intersection 返回交集：同时属于 s 与 other 的元素。
func (s *HashSet[T]) Intersection(other *HashSet[T]) *HashSet[T] {
	// 遍历较小的集合，减少比较次数。
	a, b := s, other
	if len(a.m) > len(b.m) {
		a, b = b, a
	}
	out := New[T]()
	for e := range a.m {
		if _, ok := b.m[e]; ok {
			out.m[e] = struct{}{}
		}
	}
	return out
}

// Difference 返回差集：属于 s 但不属于 other 的元素。
func (s *HashSet[T]) Difference(other *HashSet[T]) *HashSet[T] {
	out := New[T]()
	for e := range s.m {
		if _, ok := other.m[e]; !ok {
			out.m[e] = struct{}{}
		}
	}
	return out
}

// SymmetricDifference 返回对称差集：恰好只属于其中一方的元素。
func (s *HashSet[T]) SymmetricDifference(other *HashSet[T]) *HashSet[T] {
	out := New[T]()
	for e := range s.m {
		if _, ok := other.m[e]; !ok {
			out.m[e] = struct{}{}
		}
	}
	for e := range other.m {
		if _, ok := s.m[e]; !ok {
			out.m[e] = struct{}{}
		}
	}
	return out
}

// IsSubset 判断 s 是否是 other 的子集（s ⊆ other）。
func (s *HashSet[T]) IsSubset(other *HashSet[T]) bool {
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

// IsSuperset 判断 s 是否是 other 的超集（s ⊇ other）。
func (s *HashSet[T]) IsSuperset(other *HashSet[T]) bool {
	return other.IsSubset(s)
}

// Equal 判断两个集合元素是否完全相同。
func (s *HashSet[T]) Equal(other *HashSet[T]) bool {
	if len(s.m) != len(other.m) {
		return false
	}
	return s.IsSubset(other)
}

// IsDisjoint 判断两个集合是否不相交（无公共元素）。
func (s *HashSet[T]) IsDisjoint(other *HashSet[T]) bool {
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

// AddAll 把 other 的所有元素并入 s（可变操作，返回 s 以支持链式调用）。
func (s *HashSet[T]) AddAll(other *HashSet[T]) *HashSet[T] {
	s.ensure()
	for e := range other.m {
		s.m[e] = struct{}{}
	}
	return s
}

// RemoveAll 从 s 中删除 other 含有的所有元素（可变操作，返回 s）。
func (s *HashSet[T]) RemoveAll(other *HashSet[T]) *HashSet[T] {
	for e := range other.m {
		delete(s.m, e)
	}
	return s
}

// RetainAll 保留 s 与 other 的交集（可变操作，返回 s）。
func (s *HashSet[T]) RetainAll(other *HashSet[T]) *HashSet[T] {
	for e := range s.m {
		if _, ok := other.m[e]; !ok {
			delete(s.m, e)
		}
	}
	return s
}

// CollectSeq 把 iter.Seq 中的元素收集为集合，配合 slices/maps 及惰性迭代器使用：
//
//	s := set.CollectSeq(slices.Values([]int{1, 2, 3}))
func CollectSeq[T comparable](seq iter.Seq[T]) *HashSet[T] {
	s := New[T]()
	for e := range seq {
		s.m[e] = struct{}{}
	}
	return s
}
