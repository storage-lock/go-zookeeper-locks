package zookeeper_locks

import (
	"context"
	"strings"

	"github.com/go-zookeeper/zk"
	zookeeper_storage "github.com/storage-lock/go-zookeeper-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
)

var hostsStorageLockFactoryBeanFactory *storage_lock_factory.StorageLockFactoryBeanFactory[string, *zk.Conn] = storage_lock_factory.NewStorageLockFactoryBeanFactory[string, *zk.Conn]()

func NewZookeeperLockByHosts(ctx context.Context, hosts []string, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetZookeeperLockFactoryByHosts(ctx, hosts)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

func NewZookeeperLockByHostsWithOptions(ctx context.Context, hosts []string, options *storage_lock.StorageLockOptions) (*storage_lock.StorageLock, error) {
	factory, err := GetZookeeperLockFactoryByHosts(ctx, hosts)
	if err != nil {
		return nil, err
	}
	return factory.CreateLockWithOptions(options)
}

// hostsKey 将服务器列表转为缓存 key
func hostsKey(hosts []string) string {
	return strings.Join(hosts, ",")
}

func GetZookeeperLockFactoryByHosts(ctx context.Context, hosts []string) (*storage_lock_factory.StorageLockFactory[*zk.Conn], error) {
	key := hostsKey(hosts)
	return hostsStorageLockFactoryBeanFactory.GetOrInit(ctx, key, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*zk.Conn], error) {
		connectionManager := zookeeper_storage.NewZookeeperConnectionManagerFromHosts(hosts)
		options := zookeeper_storage.NewZookeeperStorageOptions().SetConnectionManager(connectionManager)
		// newZookeeperStorage 定义在 global_client.go，双入口共用以便单测统一注入
		storage, err := newZookeeperStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		factory := storage_lock_factory.NewStorageLockFactory(storage, options.ConnectionManager)
		return factory, nil
	})
}
