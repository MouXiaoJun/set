package set

import (
	"math/rand"
	"slices"
	"testing"
)

// FuzzDequeModel 用标准 slice 模拟对照验证 Deque 的 Push/Pop/Peek 行为。
func FuzzDequeModel(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{})
	f.Add([]byte{0xff, 0, 1})
	f.Fuzz(func(t *testing.T, ops []byte) {
		d := NewDeque[int]()
		var ref []int // 头部 = 下标 0
		for _, op := range ops {
			switch op % 6 {
			case 0: // PushBack
				v := int(op)
				d.PushBack(v)
				ref = append(ref, v)
			case 1: // PushFront
				v := int(op)
				d.PushFront(v)
				ref = append([]int{v}, ref...)
			case 2: // PopFront
				want, wantOK := popFront(ref)
				ref = trimFront(ref)
				got, gotOK := d.PopFront()
				if got != want || gotOK != wantOK {
					t.Fatalf("PopFront got (%d,%v) want (%d,%v)", got, gotOK, want, wantOK)
				}
			case 3: // PopBack
				want, wantOK := popBack(ref)
				ref = trimBack(ref)
				got, gotOK := d.PopBack()
				if got != want || gotOK != wantOK {
					t.Fatalf("PopBack got (%d,%v) want (%d,%v)", got, gotOK, want, wantOK)
				}
			case 4: // PeekFront
				want, wantOK := popFront(ref)
				got, gotOK := d.PeekFront()
				if got != want || gotOK != wantOK {
					t.Fatalf("PeekFront got (%d,%v) want (%d,%v)", got, gotOK, want, wantOK)
				}
			case 5: // Len / Elements
				if d.Len() != len(ref) {
					t.Fatalf("Len = %d, want %d", d.Len(), len(ref))
				}
				if !slices.Equal(d.Elements(), ref) {
					t.Fatalf("Elements = %v, want %v", d.Elements(), ref)
				}
			}
		}
	})
}

func popFront(s []int) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}
	return s[0], true
}
func trimFront(s []int) []int {
	if len(s) == 0 {
		return s
	}
	return s[1:]
}
func popBack(s []int) (int, bool) {
	if len(s) == 0 {
		return 0, false
	}
	return s[len(s)-1], true
}
func trimBack(s []int) []int {
	if len(s) == 0 {
		return s
	}
	return s[:len(s)-1]
}

// FuzzHeapOrder 随机 Push/Pop 后验证弹出顺序非递减（小顶堆）。
func FuzzHeapOrder(f *testing.F) {
	f.Add([]byte{5, 3, 1, 4, 2})
	f.Add([]byte{})
	f.Add([]byte{1, 1, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		h := NewMinHeap[int]()
		for _, b := range data {
			h.Push(int(b))
		}
		prev := -1
		for v, ok := h.Pop(); ok; v, ok = h.Pop() {
			if v < prev {
				t.Fatalf("pop order violated: %d after %d", v, prev)
			}
			prev = v
		}
	})
}

// Benchmark 验证 Deque/Heap 的常数。

func BenchmarkDequePushBack(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := NewDeque[int]()
		for j := 0; j < 1000; j++ {
			d.PushBack(j)
		}
		sink = d.Len()
	}
}

func BenchmarkDequePushPop(b *testing.B) {
	d := NewDeque[int]()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.PushBack(i)
		_, _ = d.PopFront()
	}
}

func BenchmarkHeapPush(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h := NewMinHeap[int]()
		for j := 0; j < 1000; j++ {
			h.Push(rand.Int())
		}
		sink = h.Len()
	}
}

func BenchmarkHeapPushPop(b *testing.B) {
	h := NewMinHeap[int]()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		_, _ = h.Pop()
	}
}
