package set

import (
	"slices"
	"testing"
)

func TestDequeBasics(t *testing.T) {
	var d Deque[int] // 零值可用
	d.PushBack(1)
	d.PushBack(2)
	d.PushFront(0)
	if d.Len() != 3 {
		t.Fatalf("Len = %d", d.Len())
	}
	if v, _ := d.PeekFront(); v != 0 {
		t.Fatalf("PeekFront = %d", v)
	}
	if v, _ := d.PeekBack(); v != 2 {
		t.Fatalf("PeekBack = %d", v)
	}
	if got := d.Elements(); !slices.Equal(got, []int{0, 1, 2}) {
		t.Fatalf("Elements = %v", got)
	}
	if v, _ := d.PopFront(); v != 0 {
		t.Fatalf("PopFront = %d", v)
	}
	if v, _ := d.PopBack(); v != 2 {
		t.Fatalf("PopBack = %d", v)
	}
	if d.Len() != 1 {
		t.Fatalf("Len after pops = %d", d.Len())
	}
	if v, _ := d.PopFront(); v != 1 {
		t.Fatalf("last pop = %d", v)
	}
	if !d.IsEmpty() {
		t.Fatal("should be empty")
	}
	if _, ok := d.PopFront(); ok {
		t.Fatal("pop from empty should fail")
	}
	if _, ok := d.PopBack(); ok {
		t.Fatal("pop back from empty should fail")
	}
}

func TestDequeGrow(t *testing.T) {
	// 超过初始容量 8，触发扩容
	d := NewDeque[int]()
	for i := 0; i < 1000; i++ {
		d.PushBack(i)
	}
	if d.Len() != 1000 {
		t.Fatalf("Len = %d", d.Len())
	}
	for i := 0; i < 1000; i++ {
		if v, _ := d.PopFront(); v != i {
			t.Fatalf("PopFront = %d, want %d", v, i)
		}
	}
}

func TestDequeFrontGrowth(t *testing.T) {
	d := NewDeque[int]()
	for i := 0; i < 500; i++ {
		d.PushFront(i)
	}
	// 逆序弹出
	for i := 499; i >= 0; i-- {
		if v, _ := d.PopFront(); v != i {
			t.Fatalf("PopFront = %d, want %d", v, i)
		}
	}
}

func TestDequeAt(t *testing.T) {
	d := NewDeque(10, 20, 30)
	if v, _ := d.At(0); v != 10 {
		t.Fatalf("At(0) = %d", v)
	}
	if v, _ := d.At(-1); v != 30 {
		t.Fatalf("At(-1) = %d", v)
	}
	if v, _ := d.At(-3); v != 10 {
		t.Fatalf("At(-3) = %d", v)
	}
	if _, ok := d.At(3); ok {
		t.Fatal("At(3) should fail")
	}
	if _, ok := d.At(-4); ok {
		t.Fatal("At(-4) should fail")
	}
}

func TestDequeIter(t *testing.T) {
	d := NewDeque("a", "b", "c")
	var got []string
	for v := range d.All() {
		got = append(got, v)
	}
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("iter = %v", got)
	}
}

func TestDequeClear(t *testing.T) {
	d := NewDeque(1, 2, 3)
	d.Clear()
	if !d.IsEmpty() || d.Len() != 0 {
		t.Fatalf("Clear failed: len=%d", d.Len())
	}
	d.PushBack(9) // 清空后可继续用
	if v, _ := d.PeekFront(); v != 9 {
		t.Fatalf("after clear reuse = %d", v)
	}
}

// --- Heap ---

func TestMinHeap(t *testing.T) {
	h := NewMinHeap(5, 1, 3, 2, 4)
	var got []int
	for v, ok := h.Pop(); ok; v, ok = h.Pop() {
		got = append(got, v)
	}
	if !slices.Equal(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("min heap pop order = %v", got)
	}
}

func TestMaxHeap(t *testing.T) {
	h := NewMaxHeap(5, 1, 3, 2, 4)
	var got []int
	for v, ok := h.Pop(); ok; v, ok = h.Pop() {
		got = append(got, v)
	}
	if !slices.Equal(got, []int{5, 4, 3, 2, 1}) {
		t.Fatalf("max heap pop order = %v", got)
	}
}

func TestHeapPushPopInterleaved(t *testing.T) {
	h := NewMinHeap[int]()
	h.Push(3)
	h.Push(1)
	if v, _ := h.Peek(); v != 1 {
		t.Fatalf("Peek = %d", v)
	}
	if v, _ := h.Pop(); v != 1 {
		t.Fatalf("Pop = %d", v)
	}
	h.Push(0)
	h.Push(2)
	if v, _ := h.Pop(); v != 0 {
		t.Fatalf("Pop = %d", v)
	}
	if v, _ := h.Pop(); v != 2 {
		t.Fatalf("Pop = %d", v)
	}
	if v, _ := h.Pop(); v != 3 {
		t.Fatalf("Pop = %d", v)
	}
	if _, ok := h.Pop(); ok {
		t.Fatal("pop from empty should fail")
	}
}

func TestHeapCustomComparator(t *testing.T) {
	// 按长度优先（短字符串先出）
	byLen := NewHeap(func(a, b string) int {
		return len(a) - len(b)
	}, "ccc", "a", "bb")
	var got []string
	for v, ok := byLen.Pop(); ok; v, ok = byLen.Pop() {
		got = append(got, v)
	}
	if !slices.Equal(got, []string{"a", "bb", "ccc"}) {
		t.Fatalf("custom cmp order = %v", got)
	}
}

func TestHeapEmpty(t *testing.T) {
	h := NewMinHeap[int]()
	if !h.IsEmpty() || h.Len() != 0 {
		t.Fatal("new heap should be empty")
	}
	if _, ok := h.Peek(); ok {
		t.Fatal("peek empty should fail")
	}
	if _, ok := h.Pop(); ok {
		t.Fatal("pop empty should fail")
	}
}

func TestHeapElementsAndClear(t *testing.T) {
	h := NewMinHeap(3, 1, 2)
	if h.Len() != 3 {
		t.Fatalf("Len = %d", h.Len())
	}
	// Elements 是堆序（非全局有序），但元素完整
	seen := New[int]()
	for _, e := range h.Elements() {
		seen.Add(e)
	}
	if !seen.Equal(New(1, 2, 3)) {
		t.Fatalf("Elements content = %v", h.Elements())
	}
	// 迭代器同样
	n := 0
	for range h.All() {
		n++
	}
	if n != 3 {
		t.Fatalf("iter count = %d", n)
	}
	h.Clear()
	if !h.IsEmpty() {
		t.Fatal("clear failed")
	}
}
