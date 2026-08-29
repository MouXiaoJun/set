# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2026-02-XX

### Added
- Deque：环形缓冲双端队列，两端 O(1)，支持负索引 At(-1)，iter 迭代
- Heap：泛型二叉堆（NewMinHeap / NewMaxHeap / 自定义比较器），批量堆化 O(n)

## [Unreleased]

## [1.1.0] - 2026-02-XX

### Changed
- `SortedSet` 批量 `Add` 改为"过滤 + 排序去重 + 归并"：从逐个插入的 O(n·m) 降为 O(n + m log m)，1000 元素批量添加快约 40%
- `SortedSet` 批量 `Remove` 构建去重删除集后一次扫描重建，避免逐个删除的 O(k·n)
- 二分查找从 `sort.Search` 换成 `slices.BinarySearchFunc`：`Contains` 34ns → 19ns（1K 元素）
- `SafeSet` 集合运算改为快照模式：各自 RLock 快照后在底层 `HashSet` 上做 map 运算，去掉逐元素加锁，且无多锁持有（消除潜在死锁顺序）
- `HashSet.New` 直接写入 map，去掉冗余的惰性初始化检查

## [1.0.0] - 2026-02-XX

### Added
- 初始版本发布
- `HashSet[T comparable]`：无序集合，map 实现，零值可用，8 种集合运算
- `SortedSet[T any]`：有序集合，二分查找 O(log n)，区间查询（Lower/Floor/Ceiling/Higher），支持非 comparable 类型
- `SafeSet[T comparable]`：线程安全集合，RWMutex + 快照迭代
- `ImmutableSet[T comparable]`：不可变集合，copy-on-write 结构共享，Freeze/Thaw 互转
- 全部类型支持 Go 1.23+ `iter.Seq` 迭代（range-over-func）
- 零依赖，仅使用 Go 标准库
