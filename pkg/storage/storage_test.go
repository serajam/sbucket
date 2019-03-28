package storage

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		args    args
		want    SBucketStorage
		wantErr bool
	}{
		{
			"Should create map storage",
			args{"map"},
			&sBucketMapStorage{buckets: map[string]*sBucket{}},
			false,
		},
		{
			"Should create sync map storage",
			args{"sync_map"},
			&sBucketSyncMapStorage{buckets: map[string]*syncSBucket{}},
			false,
		},
		{
			"Should return error on invalid type",
			args{""},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
