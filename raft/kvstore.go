package raft

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type KVStore struct {
	mu sync.Mutex
	db map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{
		db: make(map[string]string),
	}
}

func (kv *KVStore) Apply(command string) string {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "ERR_EMPTY_COMMAND"
	}

	action := strings.ToUpper(parts[0])
	switch action {
	case "SET":
		if len(parts) < 3 {
			return "ERR_INVALID_SET_FORMAT"
		}
		key := parts[1]
		valIndex := 2
		if parts[2] == "=" {
			if len(parts) < 4 {
				return "ERR_INVALID_SET_FORMAT"
			}
			valIndex = 3
		}
		value := strings.Join(parts[valIndex:], " ")
		kv.db[key] = value
		return "OK"

	case "GET":
		if len(parts) < 2 {
			return "ERR_INVALID_GET_FORMAT"
		}
		key := parts[1]
		val, exists := kv.db[key]
		if !exists {
			return "KEY_NOT_FOUND"
		}
		return val

	case "DEL":
		if len(parts) < 2 {
			return "ERR_INVALID_DEL_FORMAT"
		}
		key := parts[1]
		delete(kv.db, key)
		return "OK"

	default:
		return fmt.Sprintf("ERR_UNKNOWN_ACTION_%s", action)
	}
}

func (kv *KVStore) Snapshot() ([]byte, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	return json.Marshal(kv.db)
}

func (kv *KVStore) Restore(snapshot []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	var db map[string]string
	if err := json.Unmarshal(snapshot, &db); err != nil {
		return err
	}
	kv.db = db
	return nil
}

