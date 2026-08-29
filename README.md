# go-set

[中文](README_zh.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/MouXiaoJun/set.svg)](https://pkg.go.dev/github.com/MouXiaoJun/set)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MulanPSL--2.0-green.svg?style=flat-square)](LICENSE)

A **zero-dependency** collection library built on modern Go generics — four set types behind one consistent API, all supporting `range-over-func` iteration (Go 1.23+ `iter.Seq`).

## Types

| Type | Backing | Contains | Sorted | Thread-safe | Use when |
| --- | --- | --- | --- | --- | --- |
| `HashSet[T comparable]` | map | O(1) | ✗ | ✗ | Default choice; zero-value usable |
| `SortedSet[T any]` | sorted slice | O(log n) | ✓ | ✗ | Ordered traversal, range queries |
| `SafeSet[T comparable]` | map + RWMutex | O(1) | ✗ | ✓ | Shared across goroutines |
| `ImmutableSet[T comparable]` | shared map | O(1) | ✗ | ✓ (inherent) | Config snapshots, functional chains |

## Quick start

```go
s := set.New(3, 1, 2, 3) // dedupes
s.Add(4).Remove(1)       // chainable

s.Contains(2) // true
s.Len()       // 3

for v := range s.All() { // range-over-func
	fmt.Println(v)
}
```

## Set operations (pure, inputs never mutated)

```go
a := set.New(1, 2, 3)
b := set.New(3, 4)

a.Union(b)               // {1,2,3,4}
a.Intersection(b)        // {3}
a.Difference(b)          // {1,2}
a.SymmetricDifference(b) // {1,2,4}
a.IsSubset(b)            // false
a.IsDisjoint(b)          // false
```

In-place variants: `AddAll`, `RemoveAll`, `RetainAll`.

## Deque & Heap

```go
d := set.NewDeque(1, 2, 3)
d.PushFront(0)
d.PopBack() // 3
d.At(-1)    // 2

h := set.NewMinHeap(5, 1, 3)
h.Push(0)
v, _ := h.Pop() // 0 — always the minimum

max := set.NewMaxHeap(1, 2, 3) // max-heap
byLen := set.NewHeap(func(a, b string) int { return len(a) - len(b) }, "ccc", "a") // custom
```

`Deque` is a ring buffer — O(1) push/pop at both ends, negative indexing (`At(-1)`). `Heap` is a generic binary heap — O(log n) push/pop, O(n) batch heapify, min/max/custom comparators.

## SortedSet

```go
s := set.NewOrderedSet(5, 1, 3, 2, 4)
s.Elements() // [1 2 3 4 5]
s.Min()      // 1, true
s.Floor(3)   // 3    (largest <= 3)
s.Higher(3)  // 4    (smallest > 3)
s.Lower(3)   // 2    (largest < 3)
s.Ceiling(3) // 3    (smallest >= 3)
```

`T` needs no `comparable` constraint — equality comes from your comparator, so structs with slice/map fields work:

```go
byID := set.NewSortedSet(func(a, b User) int { return a.ID - b.ID }, users...)
```

Union/Intersection/etc. use merge over two sorted slices: O(n+m), result stays sorted.

## SafeSet

```go
s := set.NewSafeSet(1, 2, 3)
go func() { s.Add(4) }()
go func() { _ = s.Contains(1) }() // race-free, verified with -race
```

`All()`/`Elements()` return a **snapshot**: read is locked, copy is made, the iterator never races with later writes (but won't see them).

## ImmutableSet

```go
s := set.NewImmutable(1, 2, 3)
s2 := s.Add(4)   // returns a new set; s is untouched
s.Add(1, 2) == s // true — no change, shared instance (copy-on-write)

im := set.New(1, 2).Freeze() // HashSet → ImmutableSet
hs := im.Thaw()              // ImmutableSet → HashSet
```

## Performance

Apple M5, `go test -bench . -benchmem`:

| Benchmark | Time | Allocs |
| --- | --- | --- |
| HashSet Contains (1000 elems) | 3.8 ns/op | 0 |
| Iterate 1000 elems | 5.3 µs/op | 0 |
| HashSet Add 1000 elems | 28 µs/op | 20 |
| SortedSet Contains (1000 elems) | 19 ns/op | 0 |
| SortedSet Contains (10000 elems) | 37 ns/op | 0 |
| SortedSet Elements (1000 elems) | 800 ns/op | 1 |
| SortedSet batch Add 1000 elems | 8 µs/op | 6 |

**`HashSet` is a bare map.** `Contains` is ~3.9 ns / 0 allocs — identical to a raw `map[T]struct{}` benchmark. **`SortedSet.Contains` is O(log n)**: 10× the data, 1.9× the time.

**Batch operations are optimized.** `SortedSet.Add(elems...)` with many elements goes through "filter existing → sort & dedupe → merge" instead of one binary insertion per element — ~40% faster than element-by-element `Add` and avoids the O(n·m) worst case. Batch `Remove` builds a deduped drop-set and rebuilds the slice in one scan.

## Install

```bash
go get github.com/MouXiaoJun/set
```

Requires Go 1.23+.

## License

MulanPSL-2.0
