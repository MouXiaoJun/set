package set

import (
	"cmp"
	"iter"
)

// Heap 是泛型二叉堆（优先队列）。比较器决定优先级：
// cmp(a,b) < 0 表示 a 优先于 b（默认小顶堆，堆顶最小）。
// Push / Pop 都是 O(log n)，Peek O(1)，批量构造用堆化 O(n)。
//
// T 不要求 comparable；相等性由比较器决定（与 SortedSet 一致）。
// 零值不可用（缺少比较器），必须通过 NewHeap / NewMinHeap / NewMaxHeap 构造。
type Heap[T any] struct {
	cmp func(a, b T) int
	s   []T
}

// NewHeap 创建按 cmpFn 排序的优先队列。elems 一次性传入（内部堆化 O(n)）。
func NewHeap[T any](cmpFn func(a, b T) int, elems ...T) *Heap[T] {
	h := &Heap[T]{cmp: cmpFn, s: append([]T(nil), elems...)}
	// 自底向上堆化：O(n)
	for i := len(h.s)/2 - 1; i >= 0; i-- {
		h.siftDown(i)
	}
	return h
}

// NewMinHeap 创建小顶堆（堆顶最小，T 需满足 cmp.Ordered）。
func NewMinHeap[T cmp.Ordered](elems ...T) *Heap[T] {
	return NewHeap(cmp.Compare[T], elems...)
}

// NewMaxHeap 创建大顶堆（堆顶最大，T 需满足 cmp.Ordered）。
func NewMaxHeap[T cmp.Ordered](elems ...T) *Heap[T] {
	return NewHeap(func(a, b T) int { return -cmp.Compare(a, b) }, elems...)
}

// siftUp 上浮：新插入元素向根方向移动。
func (h *Heap[T]) siftUp(i int) {
	for i > 0 {
		p := (i - 1) / 2
		if h.cmp(h.s[i], h.s[p]) >= 0 {
			break // 已满足堆序
		}
		h.s[i], h.s[p] = h.s[p], h.s[i]
		i = p
	}
}

// siftDown 下沉：根方向元素向叶子移动。
func (h *Heap[T]) siftDown(i int) {
	n := len(h.s)
	for {
		l, r := 2*i+1, 2*i+2
		smallest := i
		if l < n && h.cmp(h.s[l], h.s[smallest]) < 0 {
			smallest = l
		}
		if r < n && h.cmp(h.s[r], h.s[smallest]) < 0 {
			smallest = r
		}
		if smallest == i {
			return
		}
		h.s[i], h.s[smallest] = h.s[smallest], h.s[i]
		i = smallest
	}
}

// Push 插入元素（O(log n)）。
func (h *Heap[T]) Push(v T) {
	h.s = append(h.s, v)
	h.siftUp(len(h.s) - 1)
}

// Pop 弹出优先级最高的元素（O(log n)）；空堆返回零值与 false。
func (h *Heap[T]) Pop() (T, bool) {
	var zero T
	if len(h.s) == 0 {
		return zero, false
	}
	top := h.s[0]
	last := len(h.s) - 1
	h.s[0] = h.s[last]
	h.s[last] = zero // 释放引用
	h.s = h.s[:last]
	if len(h.s) > 1 {
		h.siftDown(0)
	}
	return top, true
}

// Peek 查看堆顶元素，不移除。空堆返回零值与 false。
func (h *Heap[T]) Peek() (T, bool) {
	var zero T
	if len(h.s) == 0 {
		return zero, false
	}
	return h.s[0], true
}

// Len 返回元素个数。
func (h *Heap[T]) Len() int {
	return len(h.s)
}

// IsEmpty 堆是否为空。
func (h *Heap[T]) IsEmpty() bool {
	return len(h.s) == 0
}

// Clear 清空堆（保留容量）。
func (h *Heap[T]) Clear() {
	var zero T
	for i := range h.s {
		h.s[i] = zero
	}
	h.s = h.s[:0]
}

// Elements 返回堆序副本（不是全局有序；要有序结果请反复 Pop）。
func (h *Heap[T]) Elements() []T {
	return append([]T(nil), h.s...)
}

// All 返回按堆序遍历的迭代器（range-over-func；不保证全局有序）。
func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, e := range h.s {
			if !yield(e) {
				return
			}
		}
	}
}
