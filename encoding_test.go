package main

import (
	"reflect"
	"testing"
)

func TestDecodeCommit(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name    string
		args    args
		want    Commit
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeCommit(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeCommit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeCommit() got = %v, want %v", got, tt.want)
			}
		})
	}
}
