package storage

import (
	"reflect"
	"sync"
	"testing"
)

func Test_syncMapValue_Value(t *testing.T) {
	type fields struct {
		val string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Should return value",
			fields{"val"},
			"val",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &syncMapValue{
				val: tt.fields.val,
			}
			if got := v.Value(); got != tt.want {
				t.Errorf("syncMapValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncMapValue_Set(t *testing.T) {
	type fields struct {
		val string
	}
	type args struct {
		val string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			"Should set value",
			fields{"val"},
			args{"val"},
			"val",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &syncMapValue{
				val: tt.fields.val,
			}
			v.Set(tt.args.val)

			if got := v.Value(); got != tt.want {
				t.Errorf("syncMapValue.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncSBucket_add(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", "test")

	type fields struct {
		data *sync.Map
	}
	type args struct {
		key string
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should add key-value",
			fields{data: &sync.Map{}},
			args{key: "test", val: "test"},
			false,
		},
		{
			"Should return error if key exists",
			fields{data: testData},
			args{key: "test", val: "test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//tt.fields.data.Store(tt.args.key, tt.args.val)
			s := &syncSBucket{
				data: tt.fields.data,
			}
			if err := s.add(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("syncSBucket.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncSBucket_update(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", "test")

	type fields struct {
		data *sync.Map
	}
	type args struct {
		key string
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should update value",
			fields{data: testData},
			args{key: "test", val: "test"},
			false,
		},
		{
			"Should return error if no key",
			fields{data: &sync.Map{}},
			args{key: "test", val: "test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncSBucket{
				data: tt.fields.data,
			}
			if err := s.update(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("syncSBucket.update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_syncSBucket_get(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", &syncMapValue{"test"})
	testData.Store("invalid", "invalid")

	type fields struct {
		data *sync.Map
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SBucketValue
		wantErr bool
	}{
		{
			"Should return value",
			fields{data: testData},
			args{key: "test"},
			&syncMapValue{"test"},
			false,
		},
		{
			"Should return error when no key",
			fields{data: &sync.Map{}},
			args{key: "test"},
			nil,
			true,
		},
		{
			"Should return error when unable to covert interface",
			fields{data: testData},
			args{key: "invalid"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncSBucket{
				data: tt.fields.data,
			}
			got, err := s.get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncSBucket.get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("syncSBucket.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syncSBucket_delete(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", &syncMapValue{"test"})

	type fields struct {
		data *sync.Map
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should delete key",
			fields{data: testData},
			args{key: "test"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncSBucket{
				data: tt.fields.data,
			}
			if err := s.delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("syncSBucket.delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSBucketSyncMapStorage(t *testing.T) {
	tests := []struct {
		name string
		want SBucketStorage
	}{
		{
			"should create symc map storage",
			&sBucketSyncMapStorage{buckets: map[string]*syncSBucket{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSBucketSyncMapStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSBucketSyncMapStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_NewBucket(t *testing.T) {
	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should create new bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{}},
			args{"test"},
			false,
		},
		{
			"Should return error when bucket key exists",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {}}},
			args{"test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			if err := s.NewBucket(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.NewBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_Add(t *testing.T) {
	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		bucket string
		key    string
		val    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should add new key-val to bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {&sync.Map{}}}},
			args{"test", "test", "test"},
			false,
		},
		{
			"Should return error when no bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{}},
			args{"test", "test", "test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			if err := s.Add(tt.args.bucket, tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_Get(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", &syncMapValue{"test"})

	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		bucket string
		key    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    SBucketValue
		wantErr bool
	}{
		{
			"Should get value by bucket and key",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {testData}}},
			args{"test", "test"},
			&syncMapValue{"test"},
			false,
		},
		{
			"Should return error when no bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{}},
			args{"test", "test"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			got, err := s.Get(tt.args.bucket, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sBucketSyncMapStorage.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_Del(t *testing.T) {
	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		bucket string
		key    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should delete value",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {&sync.Map{}}}},
			args{"test", "test"},
			false,
		},
		{
			"Should return error when no bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{}},
			args{"test", "test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			if err := s.Del(tt.args.bucket, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.Del() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_DelBucket(t *testing.T) {
	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		bucket string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should delete bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {&sync.Map{}}}},
			args{"test"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			if err := s.DelBucket(tt.args.bucket); (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.DelBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_Update(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", &syncMapValue{"test"})

	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		bucket string
		key    string
		val    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"Should update val by bucket and key",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {testData}}},
			args{"test", "test", "test"},
			false,
		},
		{
			"Should return error when no bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{}},
			args{"test", "test", "test"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			if err := s.Update(tt.args.bucket, tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_getBucket(t *testing.T) {
	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *syncSBucket
		wantErr bool
	}{
		{
			"Should return bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{"test": {&sync.Map{}}}},
			args{"test"},
			&syncSBucket{data: &sync.Map{}},
			false,
		},
		{
			"Should return error when no bucket",
			fields{mu: sync.RWMutex{}, buckets: map[string]*syncSBucket{}},
			args{"test"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			got, err := s.getBucket(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("sBucketSyncMapStorage.getBucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sBucketSyncMapStorage.getBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucketSyncMapStorage_Stats(t *testing.T) {
	testData := &sync.Map{}
	testData.Store("test", &syncMapValue{"test"})

	type fields struct {
		mu      sync.RWMutex
		buckets map[string]*syncSBucket
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"SBucketMapStorage should return stats",
			fields{buckets: map[string]*syncSBucket{"test": {data: testData}}},
			"Buckets:1\ntest:1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketSyncMapStorage{
				mu:      tt.fields.mu,
				buckets: tt.fields.buckets,
			}
			if got := s.Stats(); got != tt.want {
				t.Errorf("sBucketSyncMapStorage.Stats() = %v, want %v", got, tt.want)
			}
		})
	}
}
