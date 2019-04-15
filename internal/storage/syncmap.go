package storage

import (
	"fmt"
	"sync"
)

// syncMapValue container for actual value with possibilities to implement any additional fields
type syncMapValue struct {
	val string
}

// Value returns value
// read lock is used
func (v *syncMapValue) Value() string {
	return v.val
}

// Set sets new value
// write lock is used
func (v *syncMapValue) Set(val string) {
	v.val = val
}

// syncSBucket contains sync map for storing values
// it is concurrently safe
type syncSBucket struct {
	data *sync.Map
}

// add adds new key => val to the map using Lock. Checks if key exists
func (s *syncSBucket) add(key, val string) error {
	if _, ok := s.data.Load(key); ok {
		return fmt.Errorf("key %s exists", key)
	}
	s.data.Store(key, syncMapValue{val: val})

	return nil
}

// update updates val by map key using Lock. Checks if key exists
func (s *syncSBucket) update(key, val string) error {
	if _, ok := s.data.Load(key); !ok {
		return fmt.Errorf("key %s not found", key)
	}
	s.data.Store(key, syncMapValue{val: val})

	return nil
}

// get returns val by map key using RLock. Checks if key exists
func (s *syncSBucket) get(key string) (SBucketValue, error) {
	v, ok := s.data.Load(key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	realVal, ok := v.(*syncMapValue)
	if !ok {
		return nil, fmt.Errorf("key %s returned invalid value", key)
	}

	return realVal, nil
}

// delete deletes map record
func (s *syncSBucket) delete(key string) error {
	s.data.Delete(key)

	return nil
}

// sBucketSyncMapStorage container for buckets map wit hrw mutex
type sBucketSyncMapStorage struct {
	mu      sync.RWMutex
	buckets map[string]*syncSBucket
}

// newSBucketSyncMapStorage
func newSBucketSyncMapStorage() SBucketStorage {
	return &sBucketSyncMapStorage{buckets: map[string]*syncSBucket{}}
}

// NewBucket sets new key in buckets map with sBucket as val
// a check for existence is made and error returned if bucket exists
// read lock for checking and write lock for adding
func (s *sBucketSyncMapStorage) NewBucket(name string) error {
	s.mu.RLock()
	if _, ok := s.buckets[name]; ok {
		s.mu.RUnlock()
		return fmt.Errorf("bucket %s exists", name)
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets[name] = &syncSBucket{data: &sync.Map{}}
	return nil
}

// Add creates new key in specified buckets map
// a check for existence is made and error returned if bucket exists
// read lock for checking and write lock for adding
func (s *sBucketSyncMapStorage) Add(bucket, key, val string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return b.add(key, val)
}

// Get returns value wrapped in &syncMapValue
// a check for existence is made and error returned if bucket or key don't exists
// read lock for checking buckets map key
func (s *sBucketSyncMapStorage) Get(bucket, key string) (SBucketValue, error) {
	b, err := s.getBucket(bucket)
	if err != nil {
		return nil, err
	}

	return b.get(key)
}

// Del deletes map value by key in specified bucket
// a check for existence is made and error returned if bucket doesn't exists
// read lock for checking buckets map key is used
func (s *sBucketSyncMapStorage) Del(bucket, key string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return b.delete(key)
}

// DelBucket deletes map value by key
// write lock for del is used
func (s *sBucketSyncMapStorage) DelBucket(bucket string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.buckets, bucket)
	return nil
}

// Update updates map value by key in specified bucket
// a check for existence is made and error returned if bucket or key doesn't exists
// read lock for checking and write lock for updating
func (s *sBucketSyncMapStorage) Update(bucket, key, val string) error {
	b, err := s.getBucket(bucket)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return b.update(key, val)
}

// getBucket returns bucket by name. Checks if key exists
func (s *sBucketSyncMapStorage) getBucket(name string) (*syncSBucket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.buckets[name]
	if !ok {
		return nil, fmt.Errorf("bucket %s not found", name)
	}

	return b, nil
}

// Stats returns storage usage statistics
func (s sBucketSyncMapStorage) Stats() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := fmt.Sprint("Buckets:", len(s.buckets), "\n")

	for name, data := range s.buckets {
		length := 0

		data.data.Range(func(_, _ interface{}) bool {
			length++

			return true
		})
		stats = stats + fmt.Sprint(name, ":", length, "\n")
	}

	return stats
}
