// Map with sync.RWMutex is used for implementing table-like storage
// sync.RWMutex ensures safe concurrent r/w operations
//
// Implementations are that of described in type.go
// Implementation consists of
// 1. actual value mapValue stored as map value.
//   Can be added any additional fields such as timestamp or ttl
// 2. bucket itself sBucket with map and RW mutext
// 3. wrapper sBucketMapStorage with all access methods
//

package storage

import (
	"errors"
	"fmt"
	"sync"
)

// mapValue container for actual value with possibilities to implement any additional fields
type mapValue struct {
	mu  sync.RWMutex
	val string
}

// Value returns value
// read lock is used
func (v *mapValue) Value() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.val
}

// Set sets new value
// write lock is used
func (v *mapValue) Set(val string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.val = val
}

// sBucket contains bucket map and mutex for safe concurrent access
type sBucket struct {
	mu   sync.RWMutex
	data map[string]SBucketValue
}

func (s *sBucket) check(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.data[key]; ok {

		return fmt.Errorf("key %s exists", key)
	}

	return nil
}

// add adds new key => val to the map using Lock. Checks if key exists
func (s *sBucket) add(key, val string) error {
	if len(key) == 0 {
		return errors.New("key must not be empty")
	}

	if err := s.check(key); err != nil {
		return err
	}

	s.mu.Lock()
	s.data[key] = &mapValue{val: val}
	s.mu.Unlock()
	return nil
}

// update updates val by map key using Lock. Checks if key exists
func (s *sBucket) update(key, val string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mapVal, ok := s.data[key]
	if !ok {
		return fmt.Errorf("key %s not foung", key)
	}
	mapVal.Set(val)

	return nil
}

// get returns val by map key using RLock. Checks if key exists
func (s *sBucket) get(key string) (SBucketValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mapVal, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return mapVal, nil
}

// delete deletes map record. read lock is used
func (s *sBucket) delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)

	return nil
}

// sBucketMapStorage container for buckets map with rw mutex
type sBucketMapStorage struct {
	mu      sync.RWMutex
	buckets map[string]*sBucket
}

// newSBucketMapStorage returns new storage based on map and RW mutex
func newSBucketMapStorage() SBucketStorage {
	return &sBucketMapStorage{buckets: map[string]*sBucket{}}
}

// NewBucket sets new key in buckets map with sBucket as val
// a check for existence is made and error returned if bucket exists
// read lock for checking and write lock for adding
func (s *sBucketMapStorage) NewBucket(name string) error {
	s.mu.RLock()
	if _, ok := s.buckets[name]; ok {
		s.mu.RUnlock()

		return fmt.Errorf("bucket %s exists", name)
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets[name] = &sBucket{data: map[string]SBucketValue{}}
	return nil
}

// Add creates new key in specified bucket map
// a check for existence is made and error returned if bucket exists
// read lock for checking and write lock for adding
func (s *sBucketMapStorage) Add(bucket, key, val string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}
	return b.add(key, val)
}

// Get returns value wrapped in &mapValue
// a check for existence is made and error returned if bucket or key don't exists
// read lock for checking
func (s *sBucketMapStorage) Get(bucket, key string) (SBucketValue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, err := s.getBucket(bucket)
	if err != nil {
		return nil, err
	}

	v, err := b.get(key)
	if err != nil {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return v, nil
}

// Del deletes map value by key in specified bucket
// a check for existence is made and error returned if bucket doesn't exists
// read lock for checking and write lock for adding
func (s *sBucketMapStorage) Del(bucket, key string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}

	return b.delete(key)
}

// DelBucket deletes map value by key
// write lock for del is used
func (s *sBucketMapStorage) DelBucket(bucket string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.buckets, bucket)
	return nil
}

// Update updates map value by key in specified bucket
// a check for existence is made and error returned if bucket or key doesn't exists
// read lock for checking and write lock for updating
func (s *sBucketMapStorage) Update(bucket, key, val string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}

	return b.update(key, val)
}

// getBucket returns bucket by name. Checks if key exists
func (s *sBucketMapStorage) getBucket(name string) (*sBucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.buckets[name]
	if !ok {
		return nil, fmt.Errorf("bucket %s not found", name)
	}

	return b, nil
}

// Stats returns storage usage statistics
func (s *sBucketMapStorage) Stats() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := fmt.Sprint("Buckets:", len(s.buckets), "\n")

	for name, data := range s.buckets {
		data.mu.RLock()
		stats = stats + fmt.Sprint(name, ":", len(data.data), "\n")
		data.mu.RUnlock()
	}

	return stats
}
