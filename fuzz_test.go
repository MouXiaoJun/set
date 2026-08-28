package set

import "testing"

// FuzzHashSetModel 用参考 map 对照校验 HashSet 的 Add/Remove/Contains/Len。
// 模糊输入是任意字节串：每个字节作为操作数，第 i 个字节决定第 i 步操作。
func FuzzHashSetModel(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{1, 2, 3})
	f.Add([]byte{0xff, 0, 1, 2})
	f.Add([]byte{7, 7, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		s := New[int]()
		ref := make(map[int]struct{})
		for i, op := range data {
			v := int(op)
			switch i % 4 {
			case 0:
				s.Add(v)
				ref[v] = struct{}{}
			case 1:
				s.Remove(v)
				delete(ref, v)
			case 2:
				_, want := ref[v]
				if s.Contains(v) != want {
					t.Fatalf("Contains(%d) = %v, want %v", v, s.Contains(v), want)
				}
			case 3:
				if s.Len() != len(ref) {
					t.Fatalf("Len = %d, want %d", s.Len(), len(ref))
				}
			}
		}
	})
}

// FuzzSortedSetOrder 校验 SortedSet 的排序结果：任意元素序列读出的有序切片必须严格升序。
func FuzzSortedSetOrder(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{5, 3, 1, 4, 2})
	f.Add([]byte{1, 1, 1})
	f.Add([]byte{0xff, 0, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		elems := make([]int, len(data))
		for i, b := range data {
			elems[i] = int(b)
		}
		s := NewOrderedSet(elems...)
		got := s.Elements()
		for i := 1; i < len(got); i++ {
			if got[i-1] >= got[i] {
				t.Fatalf("not strictly ascending: %v", got)
			}
		}
		// 元素完整性：去重后与 got 等长
		seen := New[int]()
		for _, e := range elems {
			seen.Add(e)
		}
		if len(got) != seen.Len() {
			t.Fatalf("len mismatch: got %d want %d", len(got), seen.Len())
		}
	})
}
