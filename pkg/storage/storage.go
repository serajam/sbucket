package storage

import "errors"

const (
	MapType     = "map"
	SyncMapType = "sync_map"
)

// New creates new storage
func New(storageType string) (SBucketStorage, error) {
	switch storageType {
	case MapType:
		return newSBucketMapStorage(), nil
	case SyncMapType:
		return newSBucketSyncMapStorage(), nil
	default:
		return nil, errors.New("invalid storage type specified")
	}
}
