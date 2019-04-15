package storage

// SBucketValue is the interface that wraps value that
// should be stored in storage bucket.
// It is obvious that there is better type for storing data
// but string is used as mostly universal type for storing
// tokens, json, or any other serializable data. Also strings are immutable
// and we are not planning to mutate data.
type SBucketValue interface {
	Value() string
	Set(string)
}

// SBucketStorage is the interface that provides access methods for storing data.
// Any possible implementations can be used for database such as map, sync.Map or other
// custom data structures allowing to store strings
//
// Add, Get, Del, Update methods should be made concurrently safe in oder to
// maintain consistency of data.
type SBucketStorage interface {
	NewBucket(name string) error
	DelBucket(name string) error

	Add(bucket, key, val string) error
	Get(bucket, key string) (SBucketValue, error)
	Del(bucket, key string) error
	Update(bucket, key, val string) error

	Stats() string
}
