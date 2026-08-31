# go-set

Maintenance scope: preserve the published API; focus on bug fixes, security and Go compatibility, with no planned API expansion.

[中文](README_zh.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/MouXiaoJun/set.svg)](https://pkg.go.dev/github.com/MouXiaoJun/set)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=flat-square)](LICENSE)

A **zero-dependency** collection library built on modern Go generics: four set types (`HashSet`, `SortedSet`, `SafeSet`, `ImmutableSet`) plus `Deque` and `Heap`, behind one consistent API. Every type implements `All() iter.Seq[T]` for `range-over-func` iteration (Go 1.23+).

## Features

- ✅ **`HashSet[T comparable]`** — unordered set on `map[T]struct{}` (zero-byte values), **zero-value usable**, chainable `Add`/`Remove`
- ✅ **`SortedSet[T any]`** — ordered set: O(log n) binary-search `Contains`, range queries (`Lower`/`Floor`/`Ceiling`/`Higher`), supports **non-comparable types** (structs with slice/map fields)
- ✅ **`SafeSet[T comparable]`** — thread-safe: `RWMutex` wrapper, concurrent reads, snapshot iteration, verified with `go test -race`
- ✅ **`ImmutableSet[T comparable]`** — copy-on-write with structural sharing; returns the same instance when nothing changes; inherently concurrency-safe
- ✅ **`Deque[T any]`** — ring-buffer double-ended queue, O(1) amortized push/pop at both ends, negative indexing (`At(-1)`), zero-value usable
- ✅ **`Heap[T any]`** — generic binary heap: `NewMinHeap` / `NewMaxHeap` / custom comparator, O(n) batch heapify, O(log n) push/pop
- ✅ **Set operations** — `Union` / `Intersection` / `Difference` / `SymmetricDifference` / `IsSubset` / `IsSuperset` / `Equal` on all four set types; `IsDisjoint` on HashSet, SafeSet and ImmutableSet
- ✅ **In-place variants** — `AddAll` / `RemoveAll` / `RetainAll` on `HashSet`
- ✅ **Go 1.23+ iterators** — `All() iter.Seq[T]` everywhere, plus package-level `CollectSeq` for any `iter.Seq`
- ✅ **Zero dependencies** — standard library only (`iter` / `cmp` / `slices` / `sort` / `sync`)
- ✅ **Quality** — model-based fuzzing (`FuzzHashSetModel`, `FuzzSortedSetOrder`, `FuzzDequeModel`, `FuzzHeapOrder`) and race tests

## Install

```bash
go get github.com/MouXiaoJun/set
```

Requires **Go 1.23+** (for `iter.Seq` / range-over-func).

## Quick start

```go
package main

import (
	"fmt"
	"slices"

	"github.com/MouXiaoJun/set"
)

func main() {
	s := set.New(3, 1, 2, 3) // dedupes
	s.Add(4).Remove(1)       // chainable

	fmt.Println(s.Contains(2)) // true
	fmt.Println(s.Len())       // 3
	fmt.Println(s.String())    // {2, 3, 4} (sorted, for debugging)

	for v := range s.All() { // range-over-func (Go 1.23+)
		fmt.Println(v) // 2, 3, 4 (unordered)
	}

	// Set operations are pure: inputs are never mutated
	a := set.New(1, 2, 3)
	b := set.New(3, 4)
	fmt.Println(a.Union(b))                // {1, 2, 3, 4}
	fmt.Println(a.Intersection(b))         // {3}
	fmt.Println(a.Difference(b))           // {1, 2}
	fmt.Println(a.SymmetricDifference(b))  // {1, 2, 4}
	fmt.Println(a.Equal(set.New(3, 2, 1))) // true (order-independent)

	// Collect from any iter.Seq
	dedup := set.CollectSeq(slices.Values([]int{1, 2, 2, 3}))
	fmt.Println(dedup) // {1, 2, 3}
}
```

## Choosing a type

| Type | Backing | Contains | Add/Remove | Sorted | Thread-safe | Use when |
| --- | --- | --- | --- | --- | --- | --- |
| `HashSet` | map | O(1) | O(1) amortized | ✗ | ✗ | **Default**: dedupe, membership, set math |
| `SortedSet` | sorted slice | O(log n) | O(n); batch merge | ✓ | ✗ | Ordered traversal, range queries, Top-K |
| `SafeSet` | map + RWMutex | O(1) | O(1) amortized | ✗ | ✓ | Shared across goroutines |
| `ImmutableSet` | shared map | O(1) | O(n) copy on change | ✗ | ✓ (inherent) | Config snapshots, functional chains |
| `Deque` | ring buffer | — | O(1) both ends | ✗ | ✗ | Queue / stack / sliding window |
| `Heap` | binary heap | — | O(log n) | ✗ (heap order) | ✗ | Priority queue, Top-K, scheduling |

## HashSet

Backed by `map[T]struct{}`; `Contains` costs about the same as a raw map (see [Performance](#performance)). `Elements()` and `All()` have **no stable order**; use `SortedSet` when you need ordering.

```go
s := set.New(1, 2, 3)
s.Add(4).Remove(1) // chainable, idempotent

s.Union(set.New(3, 4))               // {1, 2, 3, 4}
s.Intersection(set.New(3, 4))        // {3}
s.Difference(set.New(3, 4))          // {1, 2}
s.SymmetricDifference(set.New(3, 4)) // {1, 2, 4}
s.IsSubset(set.New(1, 2, 3, 4))      // true
s.Equal(set.New(3, 2, 1))            // true

// In-place variants (return s for chaining)
s.AddAll(set.New(5))     // s = s ∪ {5}
s.RemoveAll(set.New(2))  // s = s \ {2}
s.RetainAll(set.New(3))  // s = s ∩ {3}

// Iteration (unordered) and early stop
for v := range s.All() { /* ... */ }
s.ForEach(func(v int) bool { return v != 2 }) // stop at 2
```

Package-level helpers: `New(elems...)` and `CollectSeq(seq)` for collecting any `iter.Seq` into a set.

## SortedSet

Elements stay sorted by your comparator (ascending). `T` does **not** need `comparable` — equality comes from the comparator (`cmp(a, b) == 0`), so structs with slice/map fields work.

Binary set operations require both comparators to define the same ordering and equality. `Remove` clears unused slots; `Clear` releases element references in O(n) while retaining capacity.

```go
s := set.NewOrderedSet(5, 1, 3, 2, 4)
s.Elements() // [1 2 3 4 5]
s.Min()      // 1, true
s.Max()      // 5, true

// Range queries, all O(log n) (TreeSet semantics)
s.Lower(3)    // 2   (largest < 3)
s.Floor(3)    // 3   (largest <= 3)
s.Ceiling(3)  // 3   (smallest >= 3)
s.Higher(3)   // 4   (smallest > 3)

// Custom comparator + non-comparable types
type User struct {
	ID   int
	Tags []string // not a valid map key
}
byID := set.NewSortedSet(func(a, b User) int { return a.ID - b.ID },
	User{2, nil}, User{1, nil}, User{3, nil})
for u := range byID.All() { /* ascending by ID */ }

// Set operations merge two sorted slices: O(n+m), result stays sorted
byID.Union(otherSortedSet)
```

**Batch, don't loop.** A single `Add`/`Remove` is O(n) (slice shift); a batch `Add(elems...)` filters → sorts → merges, and batch `Remove` rebuilds in one scan — both avoid O(n²). Batch-inserting 1000 elements is ~50% faster than element-by-element.

## SafeSet

`RWMutex` around `HashSet`: reads take `RLock` (concurrent), writes take `Lock`. Race-free under `go test -race`.

```go
s := set.NewSafeSet(1, 2, 3)
go func() { s.Add(4) }()
go func() { _ = s.Contains(1) }() // safe

// Snapshot semantics: the iterator captures one moment in time
iter := s.All()
s.Add(99) // not visible to the iterator already obtained
for v := range iter { /* 1, 2, 3 — no 99 */ }
```

`All()` / `Elements()` / `ForEach` copy under one lock and release it — the returned iterator can never race with concurrent writes, at the cost of not seeing writes made *after* you obtained it. Set operations snapshot both sets once and then run on plain maps: no per-element locking, no multi-lock deadlock.

`String()` also formats a snapshot after releasing the lock, so element Stringers can modify the set. Snapshots are shallow: pointed-to values still need their own synchronization. Binary operations take the two snapshots separately, not atomically across both sets.

## ImmutableSet

Once created, the underlying map never changes. Every "mutation" returns a new set with copy-on-write structural sharing — and returns the **same instance** when nothing actually changes.

```go
s := set.NewImmutable(1, 2, 3)
s2 := s.Add(4)   // new set {1, 2, 3, 4}; s is untouched
s3 := s.Remove(1) // new set {2, 3}

s.Add(1, 2) == s // true — already present, shared (zero copy)
s.Remove(9) == s // true — nothing to remove, shared

// Functional chaining
final := set.NewImmutable(1).Add(2, 3).Remove(1).Union(set.NewImmutable(4))
fmt.Println(final) // {2, 3, 4}

// Freeze / Thaw convert in both directions (decoupled copies)
im := set.New(1, 2).Freeze() // HashSet → ImmutableSet
hs := im.Thaw()              // ImmutableSet → HashSet
```

Inherently concurrency-safe (read-only map), ideal for config snapshots and cross-goroutine sharing. Not for hot write paths — every real change copies the whole table (see the benchmark below).

## Deque

Ring-buffer double-ended queue. Push/pop at both ends are O(1) amortized; it grows by doubling. **Zero-value usable.**

```go
d := set.NewDeque(1, 2, 3) // head → tail: 1, 2, 3
d.PushFront(0)             // 0, 1, 2, 3
d.PushBack(4)              // 0, 1, 2, 3, 4
d.PopFront()               // 0, true
d.PopBack()                // 4, true
d.PeekFront()              // 1, true
d.At(-1)                   // 3, true (negative index: -1 = tail)
for v := range d.All() { /* head → tail */ }

var dq set.Deque[string] // zero value is fine
dq.PushBack("a")
```

All pop/peek/`At` calls return the zero value plus `false` on empty/out-of-range — no panics. Typical uses: FIFO queue, LIFO stack, sliding windows (LRU), palindrome checks.

## Heap

Generic binary heap / priority queue. The comparator decides priority: `cmp(a, b) < 0` means a has priority (min-heap by default). `T` needs no `comparable` constraint.

```go
h := set.NewMinHeap(5, 1, 3, 2, 4)
v, _ := h.Peek() // 1
h.Push(0)
v, _ = h.Pop() // 0 — always the minimum

max := set.NewMaxHeap(1, 2, 3) // max-heap
v, _ = max.Pop()               // 3

byLen := set.NewHeap(func(a, b string) int { return len(a) - len(b) }, "ccc", "a", "bb")
v, _ = byLen.Pop() // "a"

// Batch construction heapifies in O(n); Push/Pop are O(log n)
```

`Elements()` / `All()` are in **heap order** (parent-before-child), not globally sorted — repeatedly `Pop` for sorted output. Use for Top-K, scheduling, merging sorted streams.

## Boundaries & limitations

- **Zero values**: `HashSet` and `Deque` are documented zero-value usable. `SortedSet` and `Heap` **must** be constructed (`NewSortedSet` / `NewOrderedSet` / `NewHeap` / `NewMinHeap` / `NewMaxHeap`) — a zero value has a nil comparator and panics. `SafeSet` / `ImmutableSet` should be created with `NewSafeSet` / `NewImmutable`.
- **SafeSet snapshots** reflect the moment you called `All()` / `Elements()` / `ForEach` — writes after that are invisible to the iterator (the set itself loses nothing; re-fetch for a fresh view).
- **ImmutableSet** copies the whole map on every real change — fine for config snapshots and functional chains, wasteful for hot write paths.
- **SortedSet**: single-element `Add`/`Remove` is O(n); always batch when inserting many elements to avoid O(n²). Equality is comparator-defined — make your comparator a strict weak order.
- **Unordered**: `HashSet`/`ImmutableSet`/`SafeSet` iteration order is not stable; `String()` sorts for deterministic debug output.
- **Concurrency**: `HashSet`, `SortedSet`, `Deque`, `Heap` are not thread-safe — share them via `SafeSet` / `ImmutableSet` or your own lock.
- **Empty containers**: every `Pop`/`Peek`/`Min`/`Max`/`At` returns the zero value plus `false`; check the bool.
- **Deduping a slice**: `set.New(slice...)` or `set.CollectSeq(slices.Values(slice))`.

## Performance

Apple M5, Go 1.27, `go test -bench . -benchmem` (fluctuates ±10%):

| Benchmark | Time | Allocs |
| --- | --- | --- |
| `HashSet.Contains` (1000 elems) | 3.6 ns/op | 0 |
| Raw `map` `Contains` (baseline) | 3.9 ns/op | 0 |
| Iterate 1000 elems (`All`) | 5.1 µs/op | 0 |
| `HashSet.Add` 1000 elems | 28.8 µs/op | 20 |
| `HashSet.Union` (1000 + 500) | 42.9 µs/op | 20 |
| `SortedSet.Contains` (1000 elems) | 20 ns/op | 0 |
| `SortedSet.Contains` (10000 elems) | 39 ns/op | 0 |
| `SortedSet.Elements` (1000 elems) | 0.8 µs/op | 1 |
| `SortedSet.Add` 1000 elems, one by one | 15 µs/op | 14 |
| `SortedSet.Add` 1000 elems, batched | 8 µs/op | 6 |
| `SafeSet.Add` 1000 elems | 35 µs/op | 23 |
| `ImmutableSet` 100 successive `Add`s | 82 µs/op | 484 |
| `Deque` alternating `PushBack`/`PopFront` | 7.2 ns/op | 0 |
| `Heap` alternating `Push`/`Pop` | 4.0 ns/op | 0 |

**`HashSet` is a bare map**: 3.6 ns/op, zero allocs — the generic wrapper adds nothing over `map[T]struct{}` (3.9 ns baseline). **`SortedSet.Contains` is O(log n)**: 10× the data, ~2× the time (20 → 39 ns), not 10×. **Batch paths pay off**: batched `SortedSet.Add` (8 µs) is ~50% faster than per-element (15 µs) and avoids the O(n²) worst case. **`Deque`/`Heap` constants are tiny**: alternating push/pop at 7.2 ns / 4.0 ns with zero allocs. **`ImmutableSet` cost is measurable**: 100 successive changes = 82 µs and 484 allocs — the quantifiable reason to avoid it on hot write paths (reads stay O(1), zero alloc).

Reproduce:

```bash
cd set
go test -bench . -benchmem -benchtime=1s
```

## FAQ

**How does this compare to golang-set?** [golang-set](https://github.com/deckarep/golang-set) (`github.com/deckarep/golang-set/v2`, package `mapset`) also supports Go 1.23 iterators, added in v2.8.0. This library additionally bundles a slice-based `SortedSet` with range queries, `Deque`, and `Heap`. Choose based on the containers and semantics you need; iterator support alone is not a difference.

**Why isn't SortedSet a skip list / red-black tree?** A sorted slice gives the same O(log n) lookups with contiguous memory, cache-friendliness, and zero pointer overhead. The trade-off is O(n) single-element writes — hence the batch `Add`/`Remove` merge paths that eliminate O(n²). Prefer `SortedSet` for read-heavy, range-query workloads; use batch APIs (or a tree) for hot random writes.

**When should I use ImmutableSet?** Config snapshots, functional chains, read-only sharing across goroutines, and anywhere you want a value that provably never changes. Copy-on-write makes no-op operations zero-copy and reads free. Don't use it for high-frequency writes — copy the whole map instead (see benchmarks); use `HashSet` for writes, `SafeSet` for concurrent writes.

**How does it fit the MouXiaoJun family (validator / copier / mask / dict_trans)?** Same author, same style: `github.com/MouXiaoJun/validator` (struct validation), `copier` (field copy / DTO conversion, not deep copy), `mask` (data masking), `dict_trans` (dictionary translation). Common combos: dedupe candidate values with `set` before validation; keep a `set` of processed IDs to avoid duplicate work; use `SortedSet` for Top-K or timeline range queries; share config as an `ImmutableSet` snapshot across goroutines. All are zero-dependency and iterator/generics-first.

**Does a SafeSet snapshot lose data?** No. The snapshot is just a point-in-time view — writes made after you obtained an iterator aren't visible to it, but nothing is lost from the set itself. Re-fetch `All()` / `Elements()` for a fresh view. Set operations are snapshot-based too: deterministic results, no deadlock risk, at the cost of ignoring concurrent writes.

## Changelog

See [CHANGELOG.md](./CHANGELOG.md).

## License

MIT, see [LICENSE](./LICENSE).
