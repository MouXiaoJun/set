package set

import "iter"

// Deque 是泛型双端队列：基于环形缓冲区，两端 Push/Pop 都是 O(1) 摊还。
// 自动扩容（容量翻倍），支持从头到尾的迭代器（range-over-func）。
//
// 零值可用：var d Deque[int] 可以直接 Push。
type Deque[T any] struct {
	buf  []T
	head int // 第一个元素的下标
	tail int // 下一个 PushBack 位置（环形）
	size int
}

// NewDeque 创建包含 elems 的双端队列（顺序：头 → 尾）。
func NewDeque[T any](elems ...T) *Deque[T] {
	d := &Deque[T]{buf: make([]T, max(8, len(elems)))}
	for _, e := range elems {
		d.PushBack(e)
	}
	return d
}

// ensure 扩容：容量翻倍并重新排列，head 对齐下标 0。
func (d *Deque[T]) ensure() {
	if len(d.buf) == 0 {
		d.buf = make([]T, 8)
		return
	}
	if d.size < len(d.buf) {
		return
	}
	nb := make([]T, len(d.buf)*2)
	for i := 0; i < d.size; i++ {
		nb[i] = d.buf[(d.head+i)%len(d.buf)]
	}
	d.buf = nb
	d.head = 0
	d.tail = d.size
}

// PushFront 从头部插入（O(1) 摊还）。
func (d *Deque[T]) PushFront(v T) {
	d.ensure()
	d.head = (d.head - 1 + len(d.buf)) % len(d.buf)
	d.buf[d.head] = v
	d.size++
}

// PushBack 从尾部插入（O(1) 摊还）。
func (d *Deque[T]) PushBack(v T) {
	d.ensure()
	d.buf[d.tail] = v
	d.tail = (d.tail + 1) % len(d.buf)
	d.size++
}

// PopFront 弹出头部元素；空队列返回零值与 false。
func (d *Deque[T]) PopFront() (T, bool) {
	var zero T
	if d.size == 0 {
		return zero, false
	}
	v := d.buf[d.head]
	var clear T
	d.buf[d.head] = clear // 释放引用，帮助 GC
	d.head = (d.head + 1) % len(d.buf)
	d.size--
	return v, true
}

// PopBack 弹出尾部元素；空队列返回零值与 false。
func (d *Deque[T]) PopBack() (T, bool) {
	var zero T
	if d.size == 0 {
		return zero, false
	}
	d.tail = (d.tail - 1 + len(d.buf)) % len(d.buf)
	v := d.buf[d.tail]
	var clear T
	d.buf[d.tail] = clear
	d.size--
	return v, true
}

// PeekFront 查看头部元素，不移除。空队列返回零值与 false。
func (d *Deque[T]) PeekFront() (T, bool) {
	var zero T
	if d.size == 0 {
		return zero, false
	}
	return d.buf[d.head], true
}

// PeekBack 查看尾部元素，不移除。空队列返回零值与 false。
func (d *Deque[T]) PeekBack() (T, bool) {
	var zero T
	if d.size == 0 {
		return zero, false
	}
	return d.buf[(d.tail-1+len(d.buf))%len(d.buf)], true
}

// Len 返回元素个数。
func (d *Deque[T]) Len() int {
	return d.size
}

// IsEmpty 队列是否为空。
func (d *Deque[T]) IsEmpty() bool {
	return d.size == 0
}

// Clear 清空队列（保留容量）。
func (d *Deque[T]) Clear() {
	var zero T
	for i := range d.buf {
		d.buf[i] = zero
	}
	d.head, d.tail, d.size = 0, 0, 0
}

// At 按下标访问（0 = 头部，支持负索引：-1 = 尾部）。越界返回零值与 false。
func (d *Deque[T]) At(i int) (T, bool) {
	var zero T
	if i < 0 {
		i += d.size
	}
	if i < 0 || i >= d.size {
		return zero, false
	}
	return d.buf[(d.head+i)%len(d.buf)], true
}

// Elements 返回从头到尾的元素切片副本。
func (d *Deque[T]) Elements() []T {
	out := make([]T, d.size)
	for i := 0; i < d.size; i++ {
		out[i] = d.buf[(d.head+i)%len(d.buf)]
	}
	return out
}

// All 返回从头到尾迭代的迭代器（range-over-func）。
func (d *Deque[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := 0; i < d.size; i++ {
			if !yield(d.buf[(d.head+i)%len(d.buf)]) {
				return
			}
		}
	}
}
