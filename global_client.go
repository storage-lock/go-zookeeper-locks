package zookeeper_locks

import (
	"context"
	"github.com/go-zookeeper/zk"
	zookeeper_storage "github.com/storage-lock/go-zookeeper-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
)

var clientStorageLockFactoryBeanFactory *storage_lock_factory.StorageLockFactoryBeanFactory[*zk.Conn, *zk.Conn] = storage_lock_factory.NewStorageLockFactoryBeanFactory[*zk.Conn, *zk.Conn]()

// newZookeeperStorage 指向 zookeeper_storage.NewZookeeperStorage，包级变量以便单测注入构造失败，
// 覆盖 GetOrInit initFunc 的错误分支。client/hosts 双入口共用。
var newZookeeperStorage = func(ctx context.Context, options *zookeeper_storage.ZookeeperStorageOptions) (storage.Storage, error) {
	return zookeeper_storage.NewZookeeperStorage(ctx, options)
}

func NewZookeeperLockByClient(ctx context.Context, client *zk.Conn, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetZookeeperLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

func NewZookeeperLockByClientWithOptions(ctx context.Context, client *zk.Conn, options *storage_lock.StorageLockOptions) (*storage_lock.StorageLock, error) {
	factory, err := GetZookeeperLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLockWithOptions(options)
}

func GetZookeeperLockFactoryByClient(ctx context.Context, client *zk.Conn) (*storage_lock_factory.StorageLockFactory[*zk.Conn], error) {
	return clientStorageLockFactoryBeanFactory.GetOrInit(ctx, client, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*zk.Conn], error) {
		connectionManager := zookeeper_storage.NewZookeeperConnectionManagerFromClient(client)
		options := zookeeper_storage.NewZookeeperStorageOptions().SetConnectionManager(connectionManager)
		storage, err := newZookeeperStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		factory := storage_lock_factory.NewStorageLockFactory(storage, options.ConnectionManager)
		return factory, nil
	})
}
