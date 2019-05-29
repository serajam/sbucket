package client

// Client wraps logic for working with sbucket server
type Client interface {
	bucketClient
	bucketValueClient

	Ping() error
	Wait()
	Close() error
}

// BucketClient wraps logic for working with bucket
type bucketClient interface {
	CreateBucket(name string) error
	DeleteBucket(name string) error
}

// bucketValueClient wraps logic for working with bucket values
type bucketValueClient interface {
	Add(bucket, key, val string) error
	Get(bucket, key string) (string, error)
	Update(bucket, key, val string) error
	Delete(bucket, key string) error
}
