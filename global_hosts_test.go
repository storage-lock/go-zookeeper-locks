package zookeeper_locks

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewZookeeperLockByHosts(t *testing.T) {
	envName := "STORAGE_LOCK_ZOOKEEPER_HOSTS"
	hosts := os.Getenv(envName)
	if hosts == "" {
		t.Skip("STORAGE_LOCK_ZOOKEEPER_HOSTS not set, skipping ZooKeeper test")
	}

	ctx := context.Background()
	lockId := "test-zookeeper-lock"

	factory, err := GetZookeeperLockFactoryByHosts(ctx, []string{hosts})
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	lock, err := NewZookeeperLockByHosts(ctx, []string{hosts}, lockId)
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	// 加锁
	ownerId := "test-owner"
	err = lock.Lock(ctx, ownerId)
	assert.Nil(t, err)

	// 解锁
	err = lock.UnLock(ctx, ownerId)
	assert.Nil(t, err)
}
