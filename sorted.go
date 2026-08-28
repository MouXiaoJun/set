package set

import (
	"cmp"
	"iter"
	"slices"
	"sort"
)

// SortedSet 是有序集合：元素按比较器升序排列，支持有序遍历与区间查询。
//
// 底层是"始终有序的切片 + 二分查找"：
//   - Contains / Lower / Floor / Ceiling / Higher 都是 O(log n) 二分；
//   - 单元素 Add / Remove 需要移动切片元素，O(n)；
//   - 批量 Add 走"过滤 + 排序 + 归并"，批量 Remove 走"删除集 + 一次扫描重建"，
//     避免逐个插入/删除的 O(n²)；
//   - 集合运算（Union / Intersection / …）用归并算法，O(n+m) 且结果天然有序。
//
// T 不要求 comparable：相等性由比较器决定（cmp(a,b) == 0 视为同一元素），
// 因此支持任何类型，包括含 slice / map 字段、无法做 map key 的结构体。
// 比较器语义与标准库 cmp.Compare 一致：cmp(a,b) < 0 表示 a 在 b 前。
//
// 零值不可用（缺少比较器），必须通过 NewSortedSet / NewOrderedSet 构造。
type SortedSet[T any] struct {
	cmp func(a, b T) int
	srt []T // 始终保持升序
}

// NewSortedSet 创建按 cmpFn 排序的集合。elems 一次性传入（内部批量排序去重）。
func NewSortedSet[T any](cmpFn func(a, b T) int, elems ...T) *SortedSet[T] {
	s := &SortedSet[T]{cmp: cmpFn}
	if len(elems) == 0 {
		return s
	}
	s.srt = slices.Clone(elems)
	sort.Slice(s.srt, func(i, j int) bool { return s.cmp(s.srt[i], s.srt[j]) < 0 })
	s.srt = slices.CompactFunc(s.srt, func(a, b T) bool { return s.cmp(a, b) == 0 })
	return s
}

// NewOrderedSet 创建按自然顺序排序的集合（T 需满足 cmp.Ordered，如 int、string）。
func NewOrderedSet[T cmp.Ordered](elems ...T) *SortedSet[T] {
	return NewSortedSet(cmp.Compare[T], elems...)
}

// searchGE 返回第一个 >= elem 的下标（二分）。前置条件：srt 已有序。
func (s *SortedSet[T]) searchGE(elem T) int {
	i, _ := slices.BinarySearchFunc(s.srt, elem, s.cmp)
	return i
}

// searchGT 返回第一个 > elem 的下标（二分）。集合无重复元素，最多后移一次。
func (s *SortedSet[T]) searchGT(elem T) int {
	i := s.searchGE(elem)
	for i < len(s.srt) && s.cmp(s.srt[i], elem) == 0 {
		i++
	}
	return i
}

// insertOne 单元素有序插入（O(n) 移动）。
func (s *SortedSet[T]) insertOne(elem T) {
	i := s.searchGE(elem)
	if i < len(s.srt) && s.cmp(s.srt[i], elem) == 0 {
		return // 已存在
	}
	s.srt = append(s.srt, elem)
	copy(s.srt[i+1:], s.srt[i:])
	s.srt[i] = elem
}

// Add 添加元素（幂等）。
// 单元素走 O(n) 插入快路径；批量走"过滤已存在 + 排序去重 + 归并"，
// 复杂度 O(m log n + m log m + n + m)，避免逐个插入的 O(n·m)。
func (s *SortedSet[T]) Add(elems ...T) {
	if len(elems) <= 1 {
		for _, e := range elems {
			s.insertOne(e)
		}
		return
	}
	// 批量路径：过滤掉已存在的元素，剩下的排序去重后与现有有序切片归并。
	add := make([]T, 0, len(elems))
	for _, e := range elems {
		if !s.Contains(e) {
			add = append(add, e)
		}
	}
	if len(add) == 0 {
		return
	}
	sort.Slice(add, func(i, j int) bool { return s.cmp(add[i], add[j]) < 0 })
	add = slices.CompactFunc(add, func(a, b T) bool { return s.cmp(a, b) == 0 })

	merged := make([]T, 0, len(s.srt)+len(add))
	i, j := 0, 0
	for i < len(s.srt) && j < len(add) {
		if s.cmp(s.srt[i], add[j]) < 0 {
			merged = append(merged, s.srt[i])
			i++
		} else {
			merged = append(merged, add[j])
			j++
		}
	}
	merged = append(merged, s.srt[i:]...)
	merged = append(merged, add[j:]...)
	s.srt = merged
}

// removeOne 单元素删除（O(n) 移动）。
func (s *SortedSet[T]) removeOne(elem T) {
	i := s.searchGE(elem)
	if i < len(s.srt) && s.cmp(s.srt[i], elem) == 0 {
		s.srt = append(s.srt[:i], s.srt[i+1:]...)
	}
}

// Remove 删除元素；不存在的元素是 no-op。
// 单元素走 O(n) 快路径；批量构建去重删除集后一次扫描重建，O((n+m) log m)。
func (s *SortedSet[T]) Remove(elems ...T) {
	if len(elems) <= 1 {
		for _, e := range elems {
			s.removeOne(e)
		}
		return
	}
	drop := NewSortedSet(s.cmp, elems...) // 去重 + 排序
	out := s.srt[:0]                      // 复用底层数组
	for _, e := range s.srt {
		if !drop.Contains(e) {
			out = append(out, e)
		}
	}
	s.srt = out
}

// Contains 判断是否包含 elem（O(log n) 二分）。
func (s *SortedSet[T]) Contains(elem T) bool {
	_, found := slices.BinarySearchFunc(s.srt, elem, s.cmp)
	return found
}

// Len 返回元素个数。
func (s *SortedSet[T]) Len() int {
	return len(s.srt)
}

// IsEmpty 集合是否为空。
func (s *SortedSet[T]) IsEmpty() bool {
	return len(s.srt) == 0
}

// Clear 清空集合。
func (s *SortedSet[T]) Clear() {
	s.srt = s.srt[:0]
}

// Clone 返回浅拷贝（共享比较器函数）。
func (s *SortedSet[T]) Clone() *SortedSet[T] {
	return &SortedSet[T]{cmp: s.cmp, srt: slices.Clone(s.srt)}
}

// Elements 返回按比较器升序排列的元素切片副本。
func (s *SortedSet[T]) Elements() []T {
	return slices.Clone(s.srt)
}

// All 返回按升序迭代的迭代器（range-over-func）。
func (s *SortedSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, e := range s.srt {
			if !yield(e) {
				return
			}
		}
	}
}

// Min 返回最小元素。空集合返回零值与 false。
func (s *SortedSet[T]) Min() (T, bool) {
	var zero T
	if len(s.srt) == 0 {
		return zero, false
	}
	return s.srt[0], true
}

// Max 返回最大元素。空集合返回零值与 false。
func (s *SortedSet[T]) Max() (T, bool) {
	var zero T
	if len(s.srt) == 0 {
		return zero, false
	}
	return s.srt[len(s.srt)-1], true
}

// Lower 返回严格小于 elem 的最大元素（TreeSet 语义）。无则零值与 false。
func (s *SortedSet[T]) Lower(elem T) (T, bool) {
	var zero T
	i := s.searchGE(elem)
	if i == 0 {
		return zero, false
	}
	return s.srt[i-1], true
}

// Floor 返回小于等于 elem 的最大元素。无则零值与 false。
func (s *SortedSet[T]) Floor(elem T) (T, bool) {
	var zero T
	i := s.searchGT(elem)
	if i == 0 {
		return zero, false
	}
	return s.srt[i-1], true
}

// Ceiling 返回大于等于 elem 的最小元素。无则零值与 false。
func (s *SortedSet[T]) Ceiling(elem T) (T, bool) {
	var zero T
	i := s.searchGE(elem)
	if i == len(s.srt) {
		return zero, false
	}
	return s.srt[i], true
}

// Higher 返回严格大于 elem 的最小元素。无则零值与 false。
func (s *SortedSet[T]) Higher(elem T) (T, bool) {
	var zero T
	i := s.searchGT(elem)
	if i == len(s.srt) {
		return zero, false
	}
	return s.srt[i], true
}

// Union 返回并集。归并两个有序切片，O(n+m)，结果保持升序。
func (s *SortedSet[T]) Union(other *SortedSet[T]) *SortedSet[T] {
	out := &SortedSet[T]{cmp: s.cmp}
	i, j := 0, 0
	for i < len(s.srt) && j < len(other.srt) {
		switch c := s.cmp(s.srt[i], other.srt[j]); {
		case c < 0:
			out.srt = append(out.srt, s.srt[i])
			i++
		case c > 0:
			out.srt = append(out.srt, other.srt[j])
			j++
		default:
			out.srt = append(out.srt, s.srt[i])
			i++
			j++
		}
	}
	out.srt = append(out.srt, s.srt[i:]...)
	out.srt = append(out.srt, other.srt[j:]...)
	return out
}

// Intersection 返回交集。归并取相等元素，O(n+m)。
func (s *SortedSet[T]) Intersection(other *SortedSet[T]) *SortedSet[T] {
	out := &SortedSet[T]{cmp: s.cmp}
	i, j := 0, 0
	for i < len(s.srt) && j < len(other.srt) {
		switch c := s.cmp(s.srt[i], other.srt[j]); {
		case c < 0:
			i++
		case c > 0:
			j++
		default:
			out.srt = append(out.srt, s.srt[i])
			i++
			j++
		}
	}
	return out
}

// Difference 返回差集：属于 s 但不属于 other 的元素。归并，O(n+m)。
func (s *SortedSet[T]) Difference(other *SortedSet[T]) *SortedSet[T] {
	out := &SortedSet[T]{cmp: s.cmp}
	i, j := 0, 0
	for i < len(s.srt) && j < len(other.srt) {
		switch c := s.cmp(s.srt[i], other.srt[j]); {
		case c < 0:
			out.srt = append(out.srt, s.srt[i])
			i++
		case c > 0:
			j++
		default:
			i++
			j++
		}
	}
	out.srt = append(out.srt, s.srt[i:]...)
	return out
}

// SymmetricDifference 返回对称差集：恰好只属于其中一方的元素。归并，O(n+m)。
func (s *SortedSet[T]) SymmetricDifference(other *SortedSet[T]) *SortedSet[T] {
	out := &SortedSet[T]{cmp: s.cmp}
	i, j := 0, 0
	for i < len(s.srt) && j < len(other.srt) {
		switch c := s.cmp(s.srt[i], other.srt[j]); {
		case c < 0:
			out.srt = append(out.srt, s.srt[i])
			i++
		case c > 0:
			out.srt = append(out.srt, other.srt[j])
			j++
		default:
			i++
			j++
		}
	}
	out.srt = append(out.srt, s.srt[i:]...)
	out.srt = append(out.srt, other.srt[j:]...)
	return out
}

// IsSubset 判断 s ⊆ other。双指针，O(n+m)。
func (s *SortedSet[T]) IsSubset(other *SortedSet[T]) bool {
	if len(s.srt) > len(other.srt) {
		return false
	}
	i, j := 0, 0
	for i < len(s.srt) && j < len(other.srt) {
		switch c := s.cmp(s.srt[i], other.srt[j]); {
		case c < 0:
			return false // s 里有 other 没有的元素
		case c > 0:
			j++
		default:
			i++
			j++
		}
	}
	return i == len(s.srt)
}

// IsSuperset 判断 s ⊇ other。
func (s *SortedSet[T]) IsSuperset(other *SortedSet[T]) bool {
	return other.IsSubset(s)
}

// Equal 判断两个集合元素是否完全相同（比较器判定相等）。
func (s *SortedSet[T]) Equal(other *SortedSet[T]) bool {
	if len(s.srt) != len(other.srt) {
		return false
	}
	for i := range s.srt {
		if s.cmp(s.srt[i], other.srt[i]) != 0 {
			return false
		}
	}
	return true
}
