package zookeeper_locks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	zookeeper_storage "github.com/storage-lock/go-zookeeper-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	"github.com/stretchr/testify/assert"
)

// testZkHosts 本地 sl-zk 容器地址，直连可用。
func testZkHosts() []string {
	return []string{"127.0.0.1:2181"}
}

// 连接真实的 ZooKeeper；不可达时跳过整条测试。
//
// zk.Connect 是异步的，连不上不会立刻报错，要等第一次操作才失败并返回
// 一个不可用的连接——此前据此继续调用工厂方法，最终 panic。
// 这里主动探一次，把"环境不具备"提前变成 skip。
func testZkConn(t *testing.T) *zk.Conn {
	t.Helper()
	conn, _, err := zk.Connect(testZkHosts(), 5*time.Second)
	if err != nil {
		t.Skipf("ZooKeeper 不可用，跳过集成测试: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if _, _, err := conn.Get("/"); err != nil {
		t.Skipf("ZooKeeper 不可用，跳过集成测试: %v", err)
	}
	return conn
}

func TestHostsKey(t *testing.T) {
	assert.Equal(t, "a,b", hostsKey([]string{"a", "b"}))
	assert.Equal(t, "single", hostsKey([]string{"single"}))
}

func TestGetZookeeperLockFactoryByClient(t *testing.T) {
	ctx := context.Background()
	conn := testZkConn(t)

	factory, err := GetZookeeperLockFactoryByClient(ctx, conn)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetZookeeperLockFactoryByClient(ctx, conn)
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	// 端到端加解锁（真 ZK 后端）
	lock, err := NewZookeeperLockByClient(ctx, conn, "test-zk-client-lock")
	assert.Nil(t, err)
	assert.NotNil(t, lock)
	assert.Nil(t, lock.Lock(ctx, "owner-1"))
	assert.Nil(t, lock.UnLock(ctx, "owner-1"))

	opts := storage_lock.NewStorageLockOptionsWithLockId("test-zk-client-lock-opt")
	lock2, err := NewZookeeperLockByClientWithOptions(ctx, conn, opts)
	assert.Nil(t, err)
	assert.NotNil(t, lock2)
}

func TestGetZookeeperLockFactoryByHosts(t *testing.T) {
	ctx := context.Background()
	hosts := testZkHosts()

	factory, err := GetZookeeperLockFactoryByHosts(ctx, hosts)
	if err != nil {
		t.Skipf("ZooKeeper 不可用，跳过集成测试: %v", err)
	}
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetZookeeperLockFactoryByHosts(ctx, hosts)
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	// 端到端加解锁（真 ZK 后端）
	lock, err := NewZookeeperLockByHosts(ctx, hosts, "test-zk-hosts-lock")
	assert.Nil(t, err)
	assert.NotNil(t, lock)
	assert.Nil(t, lock.Lock(ctx, "owner-1"))
	assert.Nil(t, lock.UnLock(ctx, "owner-1"))

	opts := storage_lock.NewStorageLockOptionsWithLockId("test-zk-hosts-lock-opt")
	lock2, err := NewZookeeperLockByHostsWithOptions(ctx, hosts, opts)
	assert.Nil(t, err)
	assert.NotNil(t, lock2)
}

func TestZookeeperLock_StorageError(t *testing.T) {
	old := newZookeeperStorage
	newZookeeperStorage = func(ctx context.Context, options *zookeeper_storage.ZookeeperStorageOptions) (storage.Storage, error) {
		return nil, errors.New("boom")
	}
	defer func() { newZookeeperStorage = old }()

	ctx := context.Background()

	// 独立 conn 指针，避免命中成功用例的缓存（bean factory 按指针缓存）
	errConn := testZkConn(t)
	_, err := GetZookeeperLockFactoryByClient(ctx, errConn)
	assert.NotNil(t, err)
	_, err = NewZookeeperLockByClient(ctx, errConn, "x")
	assert.NotNil(t, err)
	_, err = NewZookeeperLockByClientWithOptions(ctx, errConn, storage_lock.NewStorageLockOptionsWithLockId("x"))
	assert.NotNil(t, err)

	// 独立 hosts key，避免命中成功用例的缓存
	errHosts := []string{"127.0.0.1:9999"}
	_, err = GetZookeeperLockFactoryByHosts(ctx, errHosts)
	assert.NotNil(t, err)
	_, err = NewZookeeperLockByHosts(ctx, errHosts, "x")
	assert.NotNil(t, err)
	_, err = NewZookeeperLockByHostsWithOptions(ctx, errHosts, storage_lock.NewStorageLockOptionsWithLockId("x"))
	assert.NotNil(t, err)
}
