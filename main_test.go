package main

import (
	"reflect"
	"testing"
)

func Test_newStore(t *testing.T) {
	tests := []struct {
		name string
		want *store
	}{
		{
			name: "success",
			want: &store{},
		},
		{
			name: "sad path",
			want: &store{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newStore() = %v, want %v", got, tt.want)
			}
		})
	}
}
