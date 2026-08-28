# go-set

[English](README.md)

[![Go Reference](https://pkg.go.dev/badge/github.com/MouXiaoJun/set.svg)](https://pkg.go.dev/github.com/MouXiaoJun/set)
[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MulanPSL--2.0-green.svg?style=flat-square)](LICENSE)

一个**零依赖、基于最新 Go 泛型**的集合库，四种集合一套 API，全部支持 `range-over-func` 迭代。

## 特性

- ✅ **`HashSet[T comparable]`**：无序集合，`map[T]struct{}` 底层零字节开销，**零值可用**
- ✅ **`SortedSet[T any]`**：有序集合，二分查找 O(log n)，区间查询（`Lower`/`Floor`/`Ceiling`/`Higher`），支持非 comparable 类型
- ✅ **`SafeSet[T comparable]`**：线程安全集合，RWMutex + 快照迭代，`go test -race` 验证
- ✅ **`ImmutableSet[T comparable]`**：不可变集合，copy-on-write 结构共享，天然并发安全
- ✅ **集合运算**：`Union` / `Intersection` / `Difference` / `SymmetricDifference` / `IsSubset` / `IsSuperset` / `Equal` / `IsDisjoint`
- ✅ **Go 1.23+ 迭代器**：所有类型实现 `All() iter.Seq[T]`，直接 `for v := range s.All()`
- ✅ **零依赖**：仅使用 Go 标准库（`iter` / `cmp` / `slices` / `sort`）

## 四种集合怎么选

| 集合 | 底层 | Contains | Add/Remove | 排序 | 并发安全 | 适用场景 |
| --- | --- | --- | --- | --- | --- | --- |
| `HashSet` | map | O(1) | O(1) | ✗ | ✗ | 默认选择，绝大多数场景 |
| `SortedSet` | 有序切片 | O(log n) | O(n) | ✓ | ✗ | 需要有序遍历、Top-K、区间查询 |
| `SafeSet` | map + RWMutex | O(1) | O(1) | ✗ | ✓ | 多 goroutine 共享读写 |
| `ImmutableSet` | map（结构共享） | O(1) | O(n) 复制 | ✗ | ✓（天然） | 配置快照、函数式链式、跨 goroutine 只读共享 |

## 快速开始

```go
package main

import (
	"fmt"
	"github.com/MouXiaoJun/set"
)

func main() {
	s := set.New(3, 1, 2, 3) // 自动去重
	s.Add(4).Remove(1)       // 链式调用

	fmt.Println(s.Contains(2)) // true
	fmt.Println(s.Len())       // 3

	// range-over-func 迭代（Go 1.23+）
	for v := range s.All() {
		fmt.Println(v)
	}
}
```

## 集合运算

```go
a := set.New(1, 2, 3)
b := set.New(3, 4)

a.Union(b)                // {1, 2, 3, 4}
a.Intersection(b)         // {3}
a.Difference(b)           // {1, 2}
a.SymmetricDifference(b)  // {1, 2, 4}
a.IsSubset(b)             // false
a.IsDisjoint(b)           // false
a.Equal(set.New(3, 2, 1)) // true（顺序无关）
```

集合运算都是**纯函数**：不修改任何输入，返回新集合。想要就地合并用可变操作：

```go
a.AddAll(b)     // a = a ∪ b
a.RemoveAll(b)  // a = a \ b
a.RetainAll(b)  // a = a ∩ b
```

## 迭代器与收集

```go
// 惰性收集任意 iter.Seq
s := set.CollectSeq(slices.Values([]int{1, 2, 2, 3}))

// 配合标准库转换
elems := slices.Collect(s.All()) // []int
```

## 有序集合 SortedSet

### 自然顺序

```go
s := set.NewOrderedSet(5, 1, 3, 2, 4)
s.Elements() // [1 2 3 4 5]，升序
s.Min()      // 1, true
s.Max()      // 5, true

// 区间查询（TreeSet 语义）
s.Lower(3)    // 2   （严格小于 3 的最大值）
s.Floor(3)    // 3   （小于等于 3 的最大值）
s.Ceiling(3)  // 3   （大于等于 3 的最小值）
s.Higher(3)   // 4   （严格大于 3 的最小值）
```

### 自定义比较器 + 非 comparable 类型

`SortedSet` 不要求元素可比较（相等性由比较器决定），因此支持含 slice/map 字段的结构体：

```go
type User struct {
	ID   int
	Tags []string // 无法作为 map key
}

byID := set.NewSortedSet(func(a, b User) int { return a.ID - b.ID }, users...)
for u := range byID.All() { // 按 ID 升序
	// ...
}
```

集合运算走归并算法：O(n+m)，且结果天然有序。

## 线程安全 SafeSet

```go
s := set.NewSafeSet(1, 2, 3)

// 任意 goroutine 并发调用，go test -race 验证无竞态
go func() { s.Add(4) }()
go func() { _ = s.Contains(1) }()
```

`All()` / `Elements()` 返回**快照**：读取瞬间加锁复制，拿到迭代器后与写操作无竞态。代价是迭代器不反映拿到之后的新写入。

## 不可变 ImmutableSet

```go
s := set.NewImmutable(1, 2, 3)

// 所有"修改"都返回新集合，原集合永不变
s2 := s.Add(4)
s3 := s.Remove(1)

// copy-on-write：无变化时共享同一份内存
s.Add(1, 2) == s // true，元素已存在，直接返回原集合
```

可变/不可变互转：

```go
im := set.New(1, 2, 3).Freeze() // HashSet → ImmutableSet
hs := im.Thaw()                  // ImmutableSet → HashSet
```

## 性能

Apple M5，`go test -bench . -benchmem`：

| 基准 | 耗时 | 分配 |
| --- | --- | --- |
| Contains（1000 元素） | 4.3 ns/op | 0 |
| 迭代 1000 元素 | 5.3 µs/op | 0 |
| Add 1000 元素 | 35 µs/op | 20 次 |
| SortedSet Elements（1000 元素） | 770 ns/op | 1 次 |

`HashSet.Contains` 零分配零耗时；`SafeSet` 读路径仅一次 RLock；`SortedSet` 的集合运算是归并，比逐个 `Contains` 快一个量级。

## 安装

```bash
go get github.com/MouXiaoJun/set
```

要求 Go 1.23+（`iter.Seq` 支持 range-over-func）。

## License

MulanPSL-2.0
