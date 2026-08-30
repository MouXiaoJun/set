# go-set

[English](README.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/MouXiaoJun/set.svg)](https://pkg.go.dev/github.com/MouXiaoJun/set)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=flat-square)](LICENSE)

一个**零依赖、基于最新 Go 泛型**的集合库：四种集合（`HashSet` / `SortedSet` / `SafeSet` / `ImmutableSet`）+ 两种数据结构（`Deque` / `Heap`），一套统一 API，全部支持 Go 1.23+ 的 `iter.Seq` 迭代。

> 🧰 **六种容器**：无序集合、有序集合、线程安全集合、不可变集合、双端队列、优先队列
>
> ⚡ **零开销抽象**：`HashSet` 就是裸 `map[T]struct{}`，`Contains` 与原生 map 同为个位数纳秒、零分配
>
> 🔁 **Go 1.23+ 迭代器**：所有类型实现 `All() iter.Seq[T]`，直接 `for v := range s.All()`

## 特性

- ✅ **`HashSet[T comparable]`**：无序集合，`map[T]struct{}` 底层零字节开销，**零值可用**，`Add`/`Remove` 支持链式调用
- ✅ **`SortedSet[T any]`**：有序集合，二分查找 `Contains` O(log n)，区间查询（`Lower`/`Floor`/`Ceiling`/`Higher`），**支持非 comparable 类型**（含 slice/map 字段的结构体）
- ✅ **`SafeSet[T comparable]`**：线程安全集合，`RWMutex` 包装，读可并发、写互斥，`go test -race` 验证无竞态
- ✅ **`ImmutableSet[T comparable]`**：不可变集合，copy-on-write **结构共享**——无变化时返回原实例，天然并发安全
- ✅ **`Deque[T any]`**：泛型双端队列，环形缓冲，两端 Push/Pop O(1) 摊还，支持负索引 `At(-1)`，**零值可用**
- ✅ **`Heap[T any]`**：泛型二叉堆（优先队列），`NewMinHeap` / `NewMaxHeap` / 自定义比较器，批量构造堆化 O(n)，Push/Pop O(log n)
- ✅ **集合运算**：`Union` / `Intersection` / `Difference` / `SymmetricDifference` / `IsSubset` / `IsSuperset` / `Equal`，四种集合统一提供；`IsDisjoint` 由 HashSet、SafeSet、ImmutableSet 提供
- ✅ **就地运算**：`AddAll` / `RemoveAll` / `RetainAll`（`HashSet`），不新建集合直接改
- ✅ **Go 1.23+ 迭代器**：所有类型实现 `All() iter.Seq[T]`；`CollectSeq` 从任意 `iter.Seq` 收集
- ✅ **零依赖**：仅使用 Go 标准库（`iter` / `cmp` / `slices` / `sort` / `sync`）
- ✅ **工程质量**：模型对照模糊测试（`FuzzHashSetModel` / `FuzzSortedSetOrder` / `FuzzDequeModel` / `FuzzHeapOrder`）+ 并发竞态测试

## 安装

```bash
go get github.com/MouXiaoJun/set
```

要求 **Go 1.23+**（`iter.Seq` 与 range-over-func 需要）。

## 快速开始

```go
package main

import (
	"fmt"
	"slices"

	"github.com/MouXiaoJun/set"
)

func main() {
	// 创建集合：重复元素自动去重
	s := set.New(3, 1, 2, 3)
	s.Add(4).Remove(1) // Add/Remove 返回 s，可链式调用

	fmt.Println(s.Contains(2))   // true
	fmt.Println(s.Len())         // 3
	fmt.Println(s.String())      // {2, 3, 4}（String 有序输出，便于调试）

	// range-over-func 迭代（Go 1.23+）
	for v := range s.All() {
		fmt.Println(v) // 2, 3, 4（顺序不保证）
	}

	// 集合运算：纯函数，不修改任何输入
	a := set.New(1, 2, 3)
	b := set.New(3, 4)
	fmt.Println(a.Union(b))               // {1, 2, 3, 4}
	fmt.Println(a.Intersection(b))        // {3}
	fmt.Println(a.Difference(b))          // {1, 2}
	fmt.Println(a.SymmetricDifference(b)) // {1, 2, 4}
	fmt.Println(a.Equal(set.New(3, 2, 1))) // true（顺序无关）

	// 从 iter.Seq 收集（配合标准库 slices/maps）
	dedup := set.CollectSeq(slices.Values([]int{1, 2, 2, 3}))
	fmt.Println(dedup) // {1, 2, 3}
}
```

## 选型表：六种容器怎么选

| 类型 | 底层结构 | Contains | Add/Remove | 排序 | 并发安全 | 适用场景 |
| --- | --- | --- | --- | --- | --- | --- |
| `HashSet` | `map[T]struct{}` | O(1) | O(1) 均摊 | ✗ | ✗ | **默认选择**：去重、成员判断、集合运算 |
| `SortedSet` | 有序切片 + 二分 | O(log n) | O(n)，批量归并 | ✓ 升序 | ✗ | 有序遍历、区间查询、Top-K、最值 |
| `SafeSet` | map + RWMutex | O(1) | O(1) 均摊 | ✗ | ✓ | 多 goroutine 共享读写 |
| `ImmutableSet` | map（结构共享） | O(1) | O(n) 复制（仅变化时） | ✗ | ✓ 天然 | 配置快照、函数式链式、跨 goroutine 只读共享 |
| `Deque` | 环形缓冲 | —（无 Contains） | 两端 O(1) 摊还 | ✗ | ✗ | 队列 / 栈 / 滑动窗口 |
| `Heap` | 二叉堆 | —（无 Contains） | Push/Pop O(log n) | ✗（仅堆序） | ✗ | 优先队列、Top-K、任务调度 |

---

## HashSet——默认选择

最常用的集合：基于 `map[T]struct{}`，`struct{}` 零字节不占额外空间；`Contains` 与原生 map 读写速度几乎一致（见[性能](#性能)）。

### API 一览

| 方法 | 说明 | 复杂度 |
| --- | --- | --- |
| `New(elems ...T)`（包级） | 创建集合，自动去重 | O(n) |
| `Add(elems ...T) *HashSet[T]` | 添加一个或多个元素（幂等），可链式 | O(k) 均摊 |
| `Remove(elems ...T) *HashSet[T]` | 删除元素，不存在是 no-op，可链式 | O(k) 均摊 |
| `Contains(elem T) bool` | 判断是否存在 | O(1) |
| `Len() int` / `IsEmpty() bool` | 元素个数 / 是否为空 | O(1) |
| `Clear()` | 清空集合 | O(n) |
| `Clone() *HashSet[T]` | 浅拷贝新集合 | O(n) |
| `Elements() []T` | 元素切片（**顺序不保证**） | O(n) |
| `ForEach(fn func(T) bool)` | 遍历，fn 返回 false 提前终止 | O(n) |
| `String() string` | 有序字符串输出（fmt.Stringer） | O(n log n) |
| `All() iter.Seq[T]` | 迭代器（顺序不保证） | O(n) |
| `Union` / `Intersection` / `Difference` / `SymmetricDifference` | 集合运算，返回新集合，不改输入 | O(n+m) |
| `IsSubset` / `IsSuperset` / `Equal` / `IsDisjoint` | 关系判断 | O(n+m) |
| `AddAll` / `RemoveAll` / `RetainAll` | 就地运算，修改 s 并返回 s | O(n+m) |
| `Freeze() *ImmutableSet[T]` | 转为不可变集合（复制解耦） | O(n) |
| `CollectSeq(seq iter.Seq[T])`（包级） | 从任意迭代器收集为集合 | O(n) |

### 集合运算（纯函数）

所有运算都是**纯函数**：不修改任何输入集合，返回新集合。`Intersection` 和 `IsDisjoint` 会遍历较小的集合，减少比较次数。

```go
a := set.New(1, 2, 3)
b := set.New(3, 4)

a.Union(b)               // {1, 2, 3, 4}  并集
a.Intersection(b)        // {3}           交集
a.Difference(b)          // {1, 2}        差集：属于 a 不属于 b
a.SymmetricDifference(b) // {1, 2, 4}     对称差：恰好只属于一方
a.IsSubset(b)            // false         a ⊆ b
b.IsSuperset(a)          // false         b ⊇ a
a.Equal(set.New(3, 2, 1)) // true        元素相同（顺序无关）
a.IsDisjoint(set.New(5, 6)) // true      无公共元素
// a 与 b 均未被修改
```

### 就地运算：AddAll / RemoveAll / RetainAll

想复用已有集合、避免分配新 map，用可变版本（返回 s 支持链式）：

```go
a := set.New(1, 2)
b := set.New(2, 3, 4)

a.AddAll(b)    // a = a ∪ b → {1, 2, 3, 4}
a.RemoveAll(b) // a = a \ b → {1}
a = set.New(1, 2, 3, 4)
a.RetainAll(set.New(2, 4, 6)) // a = a ∩ b → {2, 4}
```

### 迭代与收集（Go 1.23+）

```go
// 迭代：range-over-func，顺序不保证稳定
for v := range s.All() {
	// ...
}

// 早停：ForEach 的 fn 返回 false 立即终止（和 All 的 yield 语义一致）
s.ForEach(func(v int) bool {
	fmt.Println(v)
	return v != 2 // 访问到 2 后停止
})

// 收集：从任意 iter.Seq 构造集合（配合标准库）
fromSlice := set.CollectSeq(slices.Values([]int{1, 2, 2, 3})) // {1, 2, 3}
fromFilter := set.CollectSeq(maps.Keys(someMap))              // 只取 key
fromRange := set.CollectSeq(slices.Values([]int{1, 2, 3, 4, 5}))
```

---

## SortedSet——有序集合

元素按比较器**升序**排列，底层是"始终有序的切片 + 二分查找"：

- `Contains` / `Lower` / `Floor` / `Ceiling` / `Higher` 都是 O(log n) 二分；
- 单元素 `Add` / `Remove` 需要移动切片元素，O(n)；
- 批量 `Add` 走"过滤 + 排序 + 归并"，批量 `Remove` 走"删除集 + 一次扫描重建"，避免逐个操作的 O(n²)；
- 集合运算用归并算法，O(n+m) 且结果天然有序。

### API 一览

| 方法 | 说明 | 复杂度 |
| --- | --- | --- |
| `NewSortedSet(cmpFn func(a, b T) int, elems ...T)`（包级） | 按自定义比较器构造，一次传入批量排序去重 | O(n log n) |
| `NewOrderedSet(elems ...T)`（包级，`T cmp.Ordered`） | 按自然顺序构造（int/string/…） | O(n log n) |
| `Add(elems ...T)` | 添加；单个 O(n) 插入，批量走归并 | O(n) / O(m log n + m log m + n + m) |
| `Remove(elems ...T)` | 删除；单个 O(n)，批量构建删除集后一次扫描重建 | O(n) / O((n+m) log m) |
| `Contains(elem T) bool` | 二分查找 | O(log n) |
| `Len() int` / `IsEmpty() bool` | 元素个数 / 是否为空 | O(1) |
| `Clear()` | 清空并释放元素引用（保留容量） | O(n) |
| `Clone() *SortedSet[T]` | 浅拷贝（共享比较器） | O(n) |
| `Elements() []T` | 升序切片副本 | O(n) |
| `All() iter.Seq[T]` | 升序迭代器 | O(n) |
| `Min() (T, bool)` / `Max() (T, bool)` | 最小 / 最大元素，空集合返回零值与 false | O(1) |
| `Lower` / `Floor` / `Ceiling` / `Higher` | 区间查询 | O(log n) |
| `Union` / `Intersection` / `Difference` / `SymmetricDifference` | 归并运算，结果保持升序 | O(n+m) |
| `IsSubset` / `IsSuperset` / `Equal` | 关系判断 | O(n+m) |

### 自然顺序

```go
s := set.NewOrderedSet(5, 1, 3, 2, 4)
fmt.Println(s.Elements()) // [1 2 3 4 5]，升序
min, _ := s.Min()         // 1
max, _ := s.Max()         // 5

// 升序迭代
for v := range s.All() {
	fmt.Println(v) // 1 2 3 4 5
}
```

### 区间查询（Lower / Floor / Ceiling / Higher）

TreeSet 语义的四个查询，全部 O(log n)，未命中返回零值与 `false`：

```go
s := set.NewOrderedSet(1, 3, 5, 7, 9)

s.Lower(5)    // 3   （严格小于 5 的最大值）
s.Floor(5)    // 5   （小于等于 5 的最大值）
s.Ceiling(5)  // 5   （大于等于 5 的最小值）
s.Higher(5)   // 7   （严格大于 5 的最小值）

s.Lower(1)    // 0, false（没有更小的了）
s.Higher(9)   // 0, false（没有更大的了）
```

典型用途：取第 K 小/第 K 大、找"恰好不小于 X 的元素"、实现时间线游标。

### 自定义比较器与不可比较类型

`T` **不需要 comparable**——相等性由比较器决定（`cmp(a, b) == 0` 视为同一元素）。因此支持含 slice / map 字段、无法做 map key 的结构体：

```go
type User struct {
	ID   int
	Tags []string // 无法作为 map key
}

byID := set.NewSortedSet(func(a, b User) int { return a.ID - b.ID },
	User{2, []string{"x"}}, User{1, []string{"y"}}, User{3, []string{"z"}})
for u := range byID.All() { // 按 ID 升序
	fmt.Println(u.ID) // 1, 2, 3
}

// 相等性由比较器判定：ID 相同即视为同一元素（即使 Tags 不同）
byID.Add(User{1, []string{"new"}}) // no-op，ID=1 已存在
fmt.Println(byID.Len())            // 3
```

比较器语义与标准库 `cmp.Compare` 一致：`cmp(a, b) < 0` 表示 a 排在 b 前。自定义比较器建议保证全序（如测试中的按长度 + 字典序兜底）。

### 集合运算：归并，结果天然有序

```go
a := set.NewOrderedSet(1, 2, 3)
b := set.NewOrderedSet(2, 3, 4)

a.Union(b).Elements()              // [1 2 3 4]
a.Intersection(b).Elements()       // [2 3]
a.Difference(b).Elements()         // [1]
a.SymmetricDifference(b).Elements() // [1 4]
a.IsSubset(set.NewOrderedSet(1, 2, 3, 4)) // true
a.Equal(set.NewOrderedSet(3, 2, 1))       // true
```

两个有序切片双指针归并：O(n+m)，结果保持升序，且不修改任何输入。

前提：两个集合的比较器必须定义相同的排序与相等性。`Remove` 会清除底层数组中不再使用的槽位，避免保留已删除元素的引用。

### 批量操作：一次传多个，而不是逐个调

逐个 `Add` 每次都要 O(n) 移动切片元素，n 个就是 O(n²)；**批量 `Add`**（一次传入多个）走"过滤已存在 → 排序去重 → 归并"，实测 1000 元素批量比逐个**快约 50%**：

```go
// 推荐：一次批量插入（走归并路径）
s := set.NewOrderedSet[int]()
s.Add(5, 1, 3, 1, 2, 5, 3, 4) // 自动去重 → [1 2 3 4 5]

// 不推荐：逐个插入（每次 O(n) 移动，最坏 O(n²)）
for _, e := range bigSlice {
	s.Add(e)
}
```

批量 `Remove` 同理：构建去重删除集后一次扫描重建，避免逐个删除的 O(k·n)。

---

## SafeSet——线程安全集合

在 `HashSet` 外包一层 `RWMutex`：读操作走 `RLock`（可并发），写操作走 `Lock`（互斥）。专为多 goroutine 共享读写设计，`go test -race` 验证无竞态。

### API 一览

| 方法 | 说明 |
| --- | --- |
| `NewSafeSet(elems ...T)`（包级） | 创建线程安全集合 |
| `Add(elems ...T)` / `Remove(elems ...T)` | 加锁写 |
| `Contains(elem T) bool` | 加锁读（RLock） |
| `Len() int` / `IsEmpty() bool` / `Clear()` | 加锁读 / 写 |
| `Clone() *SafeSet[T]` | 快照拷贝 |
| `Elements() []T` | **快照**（锁内复制，顺序不保证） |
| `ForEach(fn func(T) bool)` | 在快照上遍历，支持早停 |
| `All() iter.Seq[T]` | **基于快照的迭代器** |
| `String() string` | 有序输出（快照上排序） |
| `Union` / `Intersection` / `Difference` / `SymmetricDifference` | 各快照一次后在底层做 map 运算 |
| `IsSubset` / `IsSuperset` / `Equal` / `IsDisjoint` | 关系判断（快照模式） |

复杂度与 `HashSet` 相同（O(1) 读写、O(n+m) 运算），外加一次锁开销；`All()` / `Elements()` / `ForEach` 是 O(n) 的快照复制。

### 并发使用

```go
s := set.NewSafeSet(1, 2, 3)

// 任意 goroutine 并发调用，无竞态（go test -race 验证）
var wg sync.WaitGroup
for w := 0; w < 8; w++ {
	wg.Add(1)
	go func(seed int) {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			v := seed*100 + i
			s.Add(v)
			_ = s.Contains(v)
		}
	}(w)
}
wg.Wait()
fmt.Println(s.Len()) // 3 + 800（每个 worker 100 个新元素）
```

### 快照语义

`All()` / `Elements()` / `ForEach` 返回**某一时刻的快照**：读取瞬间加锁复制，返回后不再持锁。这样迭代器拿到手之后即可安全使用，不会与后续写操作产生 map 并发读写竞态；代价是**快照不反映"拿到迭代器之后"的新写入**：

```go
s := set.NewSafeSet(1, 2, 3)
iter := s.All() // 快照：此刻只有 1,2,3
s.Add(4, 5)     // 之后的写入不影响已拿到的迭代器
for v := range iter {
	fmt.Println(v) // 1, 2, 3（看不到 4, 5）
}
```

集合运算也走快照模式：两个集合**各自** RLock 快照一次，再在底层 `HashSet` 上做 map 运算——不会对每个元素反复加锁，也无多锁持有，无死锁风险。

---

## ImmutableSet——不可变集合

一旦创建，底层 map 永不修改。所有"修改"操作都返回新集合，并通过 **copy-on-write 结构共享**避免整表复制：真正变化时才复制旧 map 并增删元素。因此 `ImmutableSet` **天然并发安全**。

### API 一览

| 方法 | 说明 | 复杂度 |
| --- | --- | --- |
| `NewImmutable(elems ...T)`（包级） | 创建不可变集合 | O(n) |
| `Add(elems ...T) *ImmutableSet[T]` | 返回新集合；元素已全部存在时**返回原实例** | O(k) 检查 / O(n+k) 复制 |
| `Remove(elems ...T) *ImmutableSet[T]` | 返回新集合；元素都不存在时**返回原实例** | O(k) 检查 / O(n) 复制 |
| `Contains(elem T) bool` | 判断是否存在 | O(1) |
| `Len() int` / `IsEmpty() bool` | 元素个数 / 是否为空 | O(1) |
| `Clear() *ImmutableSet[T]` | 返回空集合（不修改原集合） | O(1) |
| `Elements() []T` / `ForEach` / `All()` | 读取（顺序不保证） | O(n) |
| `String() string` | 有序输出 | O(n log n) |
| `Union` / `Intersection` / `Difference` / `SymmetricDifference` | 集合运算，返回新集合 | O(n+m) |
| `IsSubset` / `IsSuperset` / `Equal` / `IsDisjoint` | 关系判断 | O(n+m) |
| `Thaw() *HashSet[T]` | 转回可变集合（复制解耦） | O(n) |
| （`HashSet` 上）`Freeze() *ImmutableSet[T]` | 可变 → 不可变（复制解耦） | O(n) |

### 结构共享（copy-on-write）

```go
s := set.NewImmutable(1, 2, 3)

s2 := s.Add(4)    // 返回新集合 {1, 2, 3, 4}，s 仍是 {1, 2, 3}
s3 := s.Remove(1) // 返回新集合 {2, 3}

// 无变化 → 返回原实例（共享同一份 map，零复制）
fmt.Println(s.Add(1, 2) == s) // true（元素已全部存在）
fmt.Println(s.Remove(9) == s) // true（元素都不存在）
```

### 函数式链式

每次"变式"都产生新值，适合无副作用地表达数据流：

```go
final := set.NewImmutable(1).
	Add(2, 3).
	Remove(1).
	Union(set.NewImmutable(4))
fmt.Println(final) // {2, 3, 4}
// 每一步的中间集合都可以继续独立使用
```

### Freeze / Thaw 互转

```go
im := set.New(1, 2, 3).Freeze() // HashSet → ImmutableSet（复制，与原集合解耦）
hs := im.Thaw()                 // ImmutableSet → HashSet（复制）

hs.Add(4)        // 修改可变副本不影响 im
fmt.Println(im)  // {1, 2, 3}
```

---

## Deque——双端队列

基于**环形缓冲区**的泛型双端队列：两端 `Push`/`Pop` 都是 O(1) 摊还，自动扩容（容量翻倍），支持从头到尾的迭代。**零值可用**。

### API 一览

| 方法 | 说明 | 复杂度 |
| --- | --- | --- |
| `NewDeque(elems ...T)`（包级） | 创建队列（顺序：头 → 尾），初始容量 8 | O(n) |
| `PushFront(v T)` / `PushBack(v T)` | 头部 / 尾部插入 | O(1) 摊还 |
| `PopFront() (T, bool)` / `PopBack() (T, bool)` | 头部 / 尾部弹出，空队列返回零值与 false | O(1) 摊还 |
| `PeekFront() (T, bool)` / `PeekBack() (T, bool)` | 查看不移除，空队列返回零值与 false | O(1) |
| `Len() int` / `IsEmpty() bool` | 元素个数 / 是否为空 | O(1) |
| `Clear()` | 清空（保留容量，释放引用） | O(n) |
| `At(i int) (T, bool)` | 按下标访问，**支持负索引**（-1 = 尾部），越界返回零值与 false | O(1) |
| `Elements() []T` | 从头到尾的切片副本 | O(n) |
| `All() iter.Seq[T]` | 从头到尾的迭代器 | O(n) |

### 基本用法

```go
d := set.NewDeque(1, 2, 3) // 头 → 尾：1, 2, 3
d.PushFront(0)             // 0, 1, 2, 3
d.PushBack(4)              // 0, 1, 2, 3, 4

v, _ := d.PopFront()  // 0
v, _ = d.PopBack()    // 4
v, _ = d.PeekFront()  // 1
v, _ = d.PeekBack()   // 3
v, _ = d.At(0)        // 1（头部）
v, _ = d.At(-1)       // 3（负索引：-1 = 尾部）
v, _ = d.At(-3)       // 1

for v := range d.All() { // 从头到尾：1, 2, 3
	fmt.Println(v)
}

// 零值可用
var zero set.Deque[string]
zero.PushBack("a")
zero.PushFront("b")
fmt.Println(zero.Elements()) // [b a]
```

典型用途：FIFO 队列、LIFO 栈、滑动窗口（LRU 淘汰）、回文检查。

---

## Heap——优先队列

泛型二叉堆。比较器决定优先级：`cmp(a, b) < 0` 表示 a 优先于 b（默认小顶堆，堆顶最小）。`T` 不要求 comparable，相等性由比较器决定。

### API 一览

| 方法 | 说明 | 复杂度 |
| --- | --- | --- |
| `NewHeap(cmpFn func(a, b T) int, elems ...T)`（包级） | 按自定义比较器构造，一次性传入批量堆化 | O(n) |
| `NewMinHeap(elems ...T)`（包级，`T cmp.Ordered`） | 小顶堆（堆顶最小） | O(n) |
| `NewMaxHeap(elems ...T)`（包级，`T cmp.Ordered`） | 大顶堆（堆顶最大） | O(n) |
| `Push(v T)` | 插入 | O(log n) |
| `Pop() (T, bool)` | 弹出堆顶，空堆返回零值与 false | O(log n) |
| `Peek() (T, bool)` | 查看堆顶，空堆返回零值与 false | O(1) |
| `Len() int` / `IsEmpty() bool` | 元素个数 / 是否为空 | O(1) |
| `Clear()` | 清空（保留容量） | O(n) |
| `Elements() []T` | **堆序**副本（非全局有序） | O(n) |
| `All() iter.Seq[T]` | **堆序**迭代器（非全局有序） | O(n) |

### 小顶堆 / 大顶堆 / 自定义比较器

```go
// 小顶堆：堆顶始终最小
h := set.NewMinHeap(5, 1, 3, 2, 4)
v, _ := h.Peek() // 1
h.Push(0)
v, _ = h.Pop() // 0

// 升序弹出（堆排序）
var sorted []int
for v, ok := h.Pop(); ok; v, ok = h.Pop() {
	sorted = append(sorted, v)
}
fmt.Println(sorted) // [1 2 3 4 5]

// 大顶堆
max := set.NewMaxHeap(1, 2, 3)
v, _ = max.Pop() // 3

// 自定义比较器：短字符串优先
byLen := set.NewHeap(func(a, b string) int { return len(a) - len(b) }, "ccc", "a", "bb")
v, _ = byLen.Pop() // "a"

// 批量构造自动堆化 O(n)；注意 Elements()/All() 是堆序而非全局有序，
// 要全局有序请反复 Pop（Heap 本身就是原地堆排序）。
```

典型用途：Top-K、任务调度、合并有序流、Dijkstra 优先队列。

---

## 边界与限制

**零值可用性**

- ✅ `HashSet`：文档承诺零值可用，`var s set.HashSet[int]; s.Add(1)` 首次写入时惰性初始化底层 map。
- ✅ `Deque`：文档承诺零值可用，`var d set.Deque[int]` 可直接 `Push`（首次扩容到容量 8）。
- ❌ `SortedSet` / `Heap`：**必须通过构造函数创建**（`NewSortedSet` / `NewOrderedSet` / `NewHeap` / `NewMinHeap` / `NewMaxHeap`）。零值缺少比较器，调用会 panic（nil 函数调用）。
- ⚠️ `SafeSet` / `ImmutableSet`：文档未承诺零值可用，请用 `NewSafeSet` / `NewImmutable` 构造。

**SafeSet 快照不反映之后写入**：`All()` / `Elements()` / `ForEach` 拿到的是"读取瞬间"的快照，之后的 `Add`/`Remove` 不会被看到（集合本身的数据不丢）。需要最新视图就再取一次。这是有意的设计——迭代器与并发写零竞态。

**ImmutableSet 高频写不划算**：每次真正的变式都要复制整张 map（实测 100 次变式约 484 次分配，见[性能](#性能)）。高频写入请用 `HashSet`，多 goroutine 高频写请用 `SafeSet`。

**SortedSet 批量 vs 逐个**：单个 `Add`/`Remove` 是 O(n)（有序切片移动）；批量（一次传多个）走"过滤 + 排序 + 归并"或"删除集 + 一次扫描重建"，避免 O(n²)。批量插入 1000 元素比逐个快约 50%。

**顶层切片去重**：一行完成：

```go
dedup := set.New(slice...)                       // 去重（顺序不保证）
dedup2 := set.CollectSeq(slices.Values(slice))   // 从迭代器去重
```

**并行注意事项**（摘自源码注释与实现）：

- `HashSet` / `SortedSet` / `Deque` / `Heap` **不是线程安全**的，多 goroutine 共享请用 `SafeSet` 或 `ImmutableSet`，或自行加锁。
- `SafeSet` 读走 `RLock` 可并发；集合运算走快照模式（各自 RLock 一次，无多锁持有），无死锁顺序问题。
- `ImmutableSet` 底层 map 永不修改，跨 goroutine 只读共享零风险。

**其他**

- 所有 `Pop`/`Peek`/`Min`/`Max`/`At` 在空容器 / 越界时都返回**零值与 false**，不会 panic——注意检查第二个返回值。
- `HashSet` / `ImmutableSet` / `SafeSet` 的 `Elements()` 与 `All()` **顺序不保证稳定**；需要稳定顺序用 `SortedSet`，或对 `Elements()` 自行排序。
- `HashSet.String()` / `ImmutableSet.String()` 按有序形式输出（内部排序），保证结果稳定、便于调试与断言。
- `Heap.Elements()` / `All()` 是**堆序**（父子关系有序，兄弟无序），不是全局有序；要全局有序请反复 `Pop`。
- `SortedSet` 的相等性由比较器判定：`cmp(a, b) == 0` 即视为同一元素。自定义比较器请保证全序（严格弱序）。

## 性能

Apple M5，Go 1.27，`go test -bench . -benchmem`（每次波动 ±10%）：

| 基准 | 耗时 | 分配 |
| --- | --- | --- |
| `HashSet.Contains`（1000 元素） | 3.6 ns/op | 0 |
| 原生 `map` 对照 `Contains` | 3.9 ns/op | 0 |
| `HashSet` 迭代 1000 元素（`All`） | 5.1 µs/op | 0 |
| `HashSet.Add` 1000 元素 | 28.8 µs/op | 20 次 |
| `HashSet.Union`（1000 + 500 元素） | 42.9 µs/op | 20 次 |
| `SortedSet.Contains`（1000 元素） | 20 ns/op | 0 |
| `SortedSet.Contains`（10000 元素） | 39 ns/op | 0 |
| `SortedSet.Elements`（1000 元素） | 0.8 µs/op | 1 次 |
| `SortedSet` 逐个 `Add` 1000 元素 | 15 µs/op | 14 次 |
| `SortedSet` 批量 `Add` 1000 元素 | 8 µs/op | 6 次 |
| `SafeSet.Add` 1000 元素 | 35 µs/op | 23 次 |
| `ImmutableSet` 逐个 `Add`（100 次变式） | 82 µs/op | 484 次 |
| `Deque` 交替 `PushBack`/`PopFront` | 7.2 ns/op | 0 |
| `Heap` 交替 `Push`/`Pop` | 4.0 ns/op | 0 |

**HashSet 就是裸 map**：`Contains` 3.6 ns/op、零分配，与原生 `map[T]struct{}` 对照（3.9 ns/op、零分配）几乎一致——泛型包装开销为零。

**SortedSet.Contains 是 O(log n)**：数据量从 1K 涨到 10K（10 倍），耗时只从 20 ns 涨到 39 ns（约 2 倍），而不是线性 10 倍——二分查找生效。

**批量路径有专门优化**：`SortedSet` 批量 `Add` 1000 元素 8 µs，逐个 `Add` 15 µs，快约 50%（且避免 O(n²) 最坏情况）；批量 `Remove` 构建删除集后一次扫描重建。

**Deque / Heap 常数极小**：交替 Push/Pop 都是个位数纳秒、零分配——环形缓冲与二叉堆的常数开销可以忽略。

**ImmutableSet 的成本可量化**：100 次变式（每次复制整表）82 µs、484 次分配——这就是"高频写不划算"的数据依据；读路径（`Contains`）仍是 O(1)、零分配。

自己复现：

```bash
cd set
go test -bench . -benchmem -benchtime=1s
```

## FAQ

**1. 和 golang-set 比怎么样？**

[golang-set](https://github.com/deckarep/golang-set)（`github.com/deckarep/golang-set/v2`，包名 `mapset`）自 v2.8.0 起也支持 Go 1.23 迭代器。本库另外提供切片型 `SortedSet`（含区间查询）、`Deque` 和 `Heap`。应按所需容器和语义选择，不能把迭代器支持当作独有差异。

**2. 为什么 SortedSet 不用跳表 / 红黑树？**

切片 + 二分的组合在查询上与平衡树同阶（O(log n)），但内存连续、缓存友好、零指针开销、零额外依赖，代码也更简单。代价是单元素写入 O(n) 的移动——所以本库提供了批量 `Add`/`Remove` 的归并路径来消除 O(n²)。**读多写少、需要区间查询的场景选 `SortedSet`**；高频随机单点写请改用批量 API 或树形结构。

**3. 什么时候用 ImmutableSet？**

配置快照、函数式链式运算、跨 goroutine 只读共享、需要"值永不变化"的稳定视图时。copy-on-write 让无变化操作零复制（直接返回原实例），读路径与 `Contains` 完全免费。**高频写入不要用**：每次变式复制整表（见性能表），写频繁用 `HashSet`，多 goroutine 写用 `SafeSet`。

**4. 如何与 validator / copier / mask / dict_trans 配合（MouXiaoJun 家族）？**

同作者家族库：`github.com/MouXiaoJun/validator`（结构体校验）、`copier`（深拷贝）、`mask`（数据脱敏）、`dict_trans`（字典翻译）。常见组合：用 `set` 收集候选值去重后再喂给 validator 校验、用 `set` 存"已处理 ID"防止重复处理、用 `SortedSet` 做 Top-K 或时间线区间查询、用 `ImmutableSet` 做配置快照跨 goroutine 共享给 copier/mask 的产物。四个库风格一致：零依赖、泛型/迭代器优先、中文注释。

**5. SafeSet 的快照会"丢数据"吗？**

不会。快照只是"读取那一刻的视图"：`All()` 拿到迭代器之后的 `Add`/`Remove` 不会被看到，但集合本身的数据一条不少。需要最新视图就再调一次 `All()` / `Elements()`。二元集合运算分别获取两个快照，不保证跨集合原子性。`String()` 也在释放锁后格式化快照，元素 Stringer 可以修改集合。快照是浅拷贝，指针指向的可变对象仍需调用方自行同步。

## 更新日志

详见 [CHANGELOG.md](./CHANGELOG.md)。

## License

MIT，见 [LICENSE](./LICENSE)。
