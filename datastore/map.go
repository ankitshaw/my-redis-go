package storage

import (
	"sync"
	"time"
)

type ConcurrentMap struct{
	sync.RWMutex
	datastore map[string]string
}

func NewKeyValueStore() *ConcurrentMap {
	kvStore := ConcurrentMap{
		datastore: make(map[string]string)
	}
	return &kvStore
}

func (kvStore *ConcurrentMap) Load(key string) (value string, ok bool) {
	kvStore.RLock()
	defer kvStore.RUnlock()
	result, ok := gcm.internal[key]
	return result, ok
}

func (kvStore *ConcurrentMap) Store(key string, value string) {
	kvStore.RLock()
	defer kvStore.RUnlock()
	kvStore.datastore[key] = value

}

func (kvStore *ConcurrentMap) Delete(key string) bool {
	kvStore.RLock()
	defer kvStore.RUnlock()

	_ , ok := kvStore.datastore[key]
	if ok == false{
		return false
	}

	delete(kvStore.datastore, key)

	return true

}