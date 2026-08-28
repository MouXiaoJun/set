package set

import (
	"iter"
	"sync"
)

// SafeSet 是线程安全的集合：在 HashSet 外包一层 RWMutex。
// 读操作走 RLock（可并发），写操作走 Lock（互斥）。
//
// 注意：All() / Elements() 返回的是某一时刻的快照——读取瞬间加锁复制，返回后不再持锁。
// 这样迭代器拿到手之后即可安全使用，不会与后续写操作产生 map 并发读写竞态；
// 代价是快照不反映"拿到迭代器之后"的新写入。
type SafeSet[T comparable] struct {
	mu sync.RWMutex
	s  HashSet[T]
}

// NewSafeSet 创建线程安全集合。
func NewSafeSet[T comparable](elems ...T) *SafeSet[T] {
	ss := &SafeSet[T]{}
	ss.s.Add(elems...) // 构造期单线程，无需加锁
	return ss
}

// Add 添加元素。
func (s *SafeSet[T]) Add(elems ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Add(elems...)
}

// Remove 删除元素。
func (s *SafeSet[T]) Remove(elems ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Remove(elems...)
}

// Contains 判断是否包含 elem。
func (s *SafeSet[T]) Contains(elem T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.s.Contains(elem)
}

// Len 返回元素个数。
func (s *SafeSet[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.s.Len()
}

// IsEmpty 集合是否为空。
func (s *SafeSet[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.s.IsEmpty()
}

// Clear 清空集合。
func (s *SafeSet[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s.Clear()
}

// Clone 返回快照拷贝。
func (s *SafeSet[T]) Clone() *SafeSet[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	clone := &SafeSet[T]{}
	clone.s.Add(s.s.Elements()...)
	return clone
}

// Elements 返回某一时刻的元素快照（顺序不保证稳定）。
func (s *SafeSet[T]) Elements() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.s.Elements()
}

// ForEach 在快照上遍历，fn 返回 false 可提前终止。
func (s *SafeSet[T]) ForEach(fn func(T) bool) {
	for _, e := range s.Elements() {
		if !fn(e) {
			return
		}
	}
}

// All 返回基于快照的迭代器（range-over-func）。
func (s *SafeSet[T]) All() iter.Seq[T] {
	snapshot := s.Elements()
	return func(yield func(T) bool) {
		for _, e := range snapshot {
			if !yield(e) {
				return
			}
		}
	}
}

// String 实现 fmt.Stringer（在快照上排序输出，结果稳定）。
func (s *SafeSet[T]) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.s.String()
}

// Union 返回并集。
func (s *SafeSet[T]) Union(other *SafeSet[T]) *SafeSet[T] {
	out := NewSafeSet[T]()
	for e := range s.All() {
		out.Add(e)
	}
	for e := range other.All() {
		out.Add(e)
	}
	return out
}

// Intersection 返回交集。
func (s *SafeSet[T]) Intersection(other *SafeSet[T]) *SafeSet[T] {
	out := NewSafeSet[T]()
	for e := range s.All() {
		if other.Contains(e) {
			out.Add(e)
		}
	}
	return out
}

// Difference 返回差集：属于 s 但不属于 other。
func (s *SafeSet[T]) Difference(other *SafeSet[T]) *SafeSet[T] {
	out := NewSafeSet[T]()
	for e := range s.All() {
		if !other.Contains(e) {
			out.Add(e)
		}
	}
	return out
}

// SymmetricDifference 返回对称差集。
func (s *SafeSet[T]) SymmetricDifference(other *SafeSet[T]) *SafeSet[T] {
	out := NewSafeSet[T]()
	for e := range s.All() {
		if !other.Contains(e) {
			out.Add(e)
		}
	}
	for e := range other.All() {
		if !s.Contains(e) {
			out.Add(e)
		}
	}
	return out
}

// IsSubset 判断 s ⊆ other。
func (s *SafeSet[T]) IsSubset(other *SafeSet[T]) bool {
	if s.Len() > other.Len() {
		return false
	}
	for e := range s.All() {
		if !other.Contains(e) {
			return false
		}
	}
	return true
}

// IsSuperset 判断 s ⊇ other。
func (s *SafeSet[T]) IsSuperset(other *SafeSet[T]) bool {
	return other.IsSubset(s)
}

// Equal 判断两个集合元素是否完全相同。
func (s *SafeSet[T]) Equal(other *SafeSet[T]) bool {
	if s.Len() != other.Len() {
		return false
	}
	return s.IsSubset(other)
}

// IsDisjoint 判断两个集合是否不相交。
func (s *SafeSet[T]) IsDisjoint(other *SafeSet[T]) bool {
	a, b := s, other
	if a.Len() > b.Len() {
		a, b = b, a
	}
	for e := range a.All() {
		if b.Contains(e) {
			return false
		}
	}
	return true
}
