# go-zookeeper-locks

这是便利封装：它不做锁逻辑，只负责构造 `go-zookeeper-storage的 Storage`
再交给核心库。锁语义（互斥、租约、fencing token）全部来自
[go-storage-lock](https://github.com/storage-lock/go-storage-lock)。

能否建锁取决于底层存储是否满足 CAS 与可靠时间源这两个前提——
这个封装不会额外增加能力，也不会替底层背书。

包名：`zookeeper_locks`

## 分布式锁的前提

锁的正确性依赖两件事：同一个 `lockId` 上有**线性一致的 CAS**，以及
**可靠、单调且跨节点可协调的时间源**。能力标志是接口自述，不是证明；
具体约束与故障模型见
[go-storage-lock 的 README](https://github.com/storage-lock/go-storage-lock#能力边界)。

## 用法

```go
store, err := zookeeper_locks.NewZookeeperLockByHosts(...)
lock, err := storage_lock.NewStorageLock(store, "my-lock-key")
owner := lock.NewOwnerId()
err = lock.Lock(ctx, owner)
```

同一套 `构造 Storage → NewStorageLock → Lock(ctx, ownerId)` 在
[go-storage-lock-example](https://github.com/storage-lock/go-storage-lock-example) 里有
各后端的完整示例。

## License

MIT
