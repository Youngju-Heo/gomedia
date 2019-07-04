package common

import (
	"sync"
)

var (
	mutex          sync.RWMutex
	interfaceStore = []interface{}{}
	remainIndex    = make([]int, 0)
)

// SavePointer unsafe pointer
func SavePointer(v interface{}) (idx int) {
	if v == nil {
		return -1
	}

	mutex.Lock()
	if len(remainIndex) > 0 {
		idx = remainIndex[0]
		remainIndex = remainIndex[1:]
		interfaceStore[idx] = v
	} else {
		idx = len(interfaceStore)
		interfaceStore = append(interfaceStore, v)
	}
	mutex.Unlock()

	return
}

// RestorePointer restore unsafe pointer
func RestorePointer(idx int) (v interface{}) {
	if idx >= 0 {

		mutex.RLock()
		if idx < len(interfaceStore) {
			v = interfaceStore[idx]
		}
		mutex.RUnlock()
	}

	return
}

// UnrefPointer release saved unsafe pointer
func UnrefPointer(idx int) {
	if idx >= 0 {

		mutex.Lock()
		if idx < len(interfaceStore) {
			interfaceStore[idx] = nil
			remainIndex = append(remainIndex, idx)
		}
		mutex.Unlock()
	}
}
