package envconfig

import (
	"reflect"
	"testing"
)

func TestParseTag_Migration(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want TagMetadata
	}{
		{
			name: "name and new json option",
			tag:  "CONFIG,json",
			want: TagMetadata{
				Name: "CONFIG",
				JSON: true, // This is what we want 'json' to map to eventually
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTag(tt.tag)
			if err != nil {
				t.Fatalf("ParseTag() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTag() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
