package client

type Client interface {
	CreateBucket(name string) error
	DeleteBucket(name string) error
	Add(bucket, key, val string) error
	Get(bucket, key string) (string, error)
	Update(bucket, key, val string) error
	Delete(bucket, key string) error

	Ping() error
	Wait()
	Close() error
}
