package storage

import (
	"reflect"
	"sync"
	"testing"
)

func TestMapValue_Set(t *testing.T) {
	d := "some val"
	v := &mapValue{}
	v.Set(d)

	if v.val != d {
		t.Errorf("%s not equal %s", v.val, "some val")
	}
}

func TestMapValue_Value(t *testing.T) {
	d := "some val"
	v := &mapValue{}
	v.Set(d)

	if v.Value() != d {
		t.Errorf("%s not equal %s", v.val, "some val")
	}
}

func Test_sBucket_add(t *testing.T) {
	type fields struct {
		mu   sync.RWMutex
		Data map[string]SBucketValue
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
		{"sBucket key non exists",
			fields{Data: map[string]SBucketValue{}},
			args{"key", "val"},
			false},
		{"sBucket key exists",
			fields{Data: map[string]SBucketValue{"key": &mapValue{val: "val"}}},
			args{"key", "val"},
			true},
		{"sBucket empty key",
			fields{Data: map[string]SBucketValue{"key": &mapValue{val: "val"}}},
			args{"", ""},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucket{
				mu:   tt.fields.mu,
				data: tt.fields.Data,
			}

			if err := s.add(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("sBucket.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucket_update(t *testing.T) {
	type fields struct {
		mu   sync.RWMutex
		Data map[string]SBucketValue
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
		{"sBucket key non exists",
			fields{Data: map[string]SBucketValue{}},
			args{"key", "val"},
			true},
		{"sBucket key exists",
			fields{Data: map[string]SBucketValue{"key": &mapValue{val: "val"}}},
			args{"key", "val"},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucket{
				mu:   tt.fields.mu,
				data: tt.fields.Data,
			}
			if err := s.update(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("sBucket.update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucket_get(t *testing.T) {
	type fields struct {
		mu   sync.RWMutex
		Data map[string]SBucketValue
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
		{"sBucket key non exists",
			fields{Data: map[string]SBucketValue{}},
			args{"key"},
			nil,
			true},
		{"sBucket key exists",
			fields{Data: map[string]SBucketValue{"key": &mapValue{val: "val"}}},
			args{"key"},
			&mapValue{val: "val"},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucket{
				mu:   tt.fields.mu,
				data: tt.fields.Data,
			}
			got, err := s.get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("sBucket.get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sBucket.get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucket_delete(t *testing.T) {
	type fields struct {
		mu   sync.RWMutex
		Data map[string]SBucketValue
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
		{"sBucket delete",
			fields{Data: map[string]SBucketValue{"key": &mapValue{val: "val"}}},
			args{"key"},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &sBucket{
				mu:   tt.fields.mu,
				data: tt.fields.Data,
			}
			if err := s.delete(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("sBucket.delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			_, err := s.get(tt.args.key)
			if err == nil {
				t.Errorf("sBucket.get() key not deleted")
			}
		})
	}
}

func TestNewSBucketMapStorage(t *testing.T) {
	tests := []struct {
		name string
		want SBucketStorage
	}{
		{
			"SBucketMapStorage create",
			&sBucketMapStorage{buckets: map[string]*sBucket{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newSBucketMapStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newSBucketMapStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucketMapStorage_NewBucket(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
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
			"SBucketMapStorage should not create bucket if exists",
			fields{buckets: map[string]*sBucket{"test_bucket": {}}},
			args{"test_bucket"},
			true,
		},
		{
			"SBucketMapStorage create",
			fields{buckets: map[string]*sBucket{}},
			args{"test_bucket"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			if err := s.NewBucket(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.NewBucket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketMapStorage_Add(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
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
			"SBucketMapStorage should add key to bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{}}}},
			args{"test_bucket", "key", "val"},
			false,
		},
		{
			"SBucketMapStorage should return error when no bucket",
			fields{buckets: map[string]*sBucket{}},
			args{"test_bucket", "key", "val"},
			true,
		},
		{
			"SBucketMapStorage should not add key to bucket when exists",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{}}}}},
			args{"test_bucket", "key", "val"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			if err := s.Add(tt.args.bucket, tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketMapStorage_Get(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
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
			"SBucketMapStorage should get SBucketValue by key from bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{val: "val"}}}}},
			args{"test_bucket", "key"},
			&mapValue{val: "val"},
			false,
		},
		{
			"SBucketMapStorage should return error whn no bucket",
			fields{buckets: map[string]*sBucket{}},
			args{"test_bucket", "key"},
			nil,
			true,
		},
		{
			"SBucketMapStorage should return error when no key in bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{}}}},
			args{"test_bucket", "key"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			got, err := s.Get(tt.args.bucket, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sBucketMapStorage.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucketMapStorage_Del(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
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
			"SBucketMapStorage should delete key from bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{val: "val"}}}}},
			args{"test_bucket", "key"},
			false,
		},
		{
			"SBucketMapStorage should return error when bucket key dont exists",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{}}}},
			args{"t", "key"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			if err := s.Del(tt.args.bucket, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.Del() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, err := s.Get(tt.args.bucket, tt.args.key); err == nil {
				t.Errorf("sBucketMapStorage.Get() exptected error got val")
			}
		})
	}
}

func Test_sBucketMapStorage_DelBucket(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
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
			"SBucketMapStorage should delete bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{val: "val"}}}}},
			args{"test_bucket"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			if err := s.DelBucket(tt.args.bucket); (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.DelBucket() error = %v, wantErr %v", err, tt.wantErr)
			}

			if _, err := s.getBucket(tt.args.bucket); err == nil {
				t.Errorf("sBucketMapStorage.getBucket() exptected error got val")
			}
		})
	}
}

func Test_sBucketMapStorage_Update(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
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
			"SBucketMapStorage should update SBucketValue by key from bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{val: "val"}}}}},
			args{"test_bucket", "key", "val2"},
			false,
		},
		{
			"SBucketMapStorage should return error when no bucket",
			fields{buckets: map[string]*sBucket{}},
			args{"test_bucket", "key", ""},
			true,
		},
		{
			"SBucketMapStorage should return error when no key in bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{}}}},
			args{"test_bucket", "key", ""},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			if err := s.Update(tt.args.bucket, tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sBucketMapStorage_getBucket(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *sBucket
		wantErr bool
	}{
		{
			"SBucketMapStorage should return bucket",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{val: "val"}}}}},
			args{"test_bucket"},
			&sBucket{data: map[string]SBucketValue{"key": &mapValue{val: "val"}}},
			false,
		},
		{
			"SBucketMapStorage should return error when no bucket",
			fields{buckets: map[string]*sBucket{}},
			args{"test_bucket"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			got, err := s.getBucket(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("sBucketMapStorage.getBucket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sBucketMapStorage.getBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sBucketMapStorage_Stats(t *testing.T) {
	type fields struct {
		buckets map[string]*sBucket
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"SBucketMapStorage should return stats",
			fields{buckets: map[string]*sBucket{"test_bucket": {data: map[string]SBucketValue{"key": &mapValue{val: "val"}}}}},
			"Buckets:1\ntest_bucket:1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sBucketMapStorage{
				buckets: tt.fields.buckets,
			}
			if got := s.Stats(); got != tt.want {
				t.Errorf("sBucketMapStorage.Stats() = %v, want %v", got, tt.want)
			}
		})
	}
}
