package set

import (
	"fmt"
	"sync"
	"testing"
)

// TestSafeSetConcurrent 用 -race 验证并发读写无竞态。
// 多 goroutine 同时 Add/Remove/Contains/Len/迭代，跑完校验最终元素集合。
func TestSafeSetConcurrent(t *testing.T) {
	const workers = 16
	const perWorker = 2000

	s := NewSafeSet[int]()

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				v := seed*perWorker + i
				s.Add(v)
				s.Contains(v)
				if i%3 == 0 {
					s.Remove(v)
					s.Remove(v) // 重复删除 no-op
				}
			}
		}(w)
	}
	wg.Wait()

	// 每个 worker 添加 perWorker 个、删除 i%3==0 的 (perWorker+2)/3 个（0,3,...,1998 → 667）。
	removedPerWorker := (perWorker + 2) / 3
	want := workers * (perWorker - removedPerWorker)
	if s.Len() != want {
		t.Fatalf("Len = %d, want %d", s.Len(), want)
	}
}

func TestSafeSetConcurrentReadDuringWrite(t *testing.T) {
	s := NewSafeSet(1, 2, 3)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				_ = s.Len()
				_ = s.Contains(1)
				for range s.All() { // 迭代快照，不能与写并发
					break
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 1000; j++ {
			s.Add(j)
			s.Remove(j)
		}
	}()
	wg.Wait()
}

func TestSafeSetOps(t *testing.T) {
	a := NewSafeSet(1, 2, 3)
	b := NewSafeSet(3, 4)

	if u := a.Union(b); !u.Equal(NewSafeSet(1, 2, 3, 4)) {
		t.Fatalf("Union wrong: %s", u)
	}
	if i := a.Intersection(b); !i.Equal(NewSafeSet(3)) {
		t.Fatalf("Intersection wrong: %s", i)
	}
	if d := a.Difference(b); !d.Equal(NewSafeSet(1, 2)) {
		t.Fatalf("Difference wrong: %s", d)
	}
	if sd := a.SymmetricDifference(b); !sd.Equal(NewSafeSet(1, 2, 4)) {
		t.Fatalf("SymmetricDifference wrong: %s", sd)
	}
	if !a.IsSubset(NewSafeSet(1, 2, 3, 4)) || !NewSafeSet(1, 2, 3, 4).IsSuperset(a) {
		t.Fatalf("subset/superset wrong")
	}
	if !NewSafeSet(1, 2, 3).Equal(NewSafeSet(3, 2, 1)) {
		t.Fatalf("Equal wrong")
	}
	if !NewSafeSet(1, 2).IsDisjoint(NewSafeSet(3, 4)) {
		t.Fatalf("IsDisjoint wrong")
	}
}

func TestSafeSetSnapshot(t *testing.T) {
	s := NewSafeSet(1, 2, 3)
	iter := s.All() // 快照：此刻有 1,2,3
	s.Add(4, 5)     // 之后写入不影响已拿到的迭代器
	var got []int
	for v := range iter {
		got = append(got, v)
	}
	if len(got) != 3 {
		t.Fatalf("snapshot iter len = %d, want 3 (got %v)", len(got), got)
	}
}

func TestSafeSetClone(t *testing.T) {
	a := NewSafeSet(1, 2)
	b := a.Clone()
	b.Add(3)
	if a.Len() != 2 || b.Len() != 3 {
		t.Fatalf("clone isolation broken: a=%d b=%d", a.Len(), b.Len())
	}
}

func TestSafeSetString(t *testing.T) {
	s := NewSafeSet("a", "b")
	got := fmt.Sprint(s)
	if got != "{a, b}" {
		t.Fatalf("String = %q", got)
	}
}
