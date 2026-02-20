package envconfig

import (
	"reflect"
	"testing"
)

func TestParseTag(t *testing.T) {
	testCases := []struct {
		name     string
		tag      string
		want     TagMetadata
		wantErr  bool
	}{
		{
			name: "just name",
			tag:  "PORT",
			want: TagMetadata{
				Name: "PORT",
			},
		},
		{
			name: "name and required",
			tag:  "PORT,required",
			want: TagMetadata{
				Name:     "PORT",
				Required: true,
			},
		},
		{
			name: "name and default",
			tag:  "PORT,default=8080",
			want: TagMetadata{
				Name:    "PORT",
				Default: "8080",
			},
		},
		{
			name: "name and prefix",
			tag:  "SERVER,prefix=API_",
			want: TagMetadata{
				Name:   "SERVER",
				Prefix: "API_",
			},
		},
		{
			name: "name and envjson",
			tag:  "CONFIG,envjson",
			want: TagMetadata{
				Name:    "CONFIG",
				EnvJSON: true,
			},
		},
		{
			name: "all options",
			tag:  "PORT,required,default=8080,prefix=API_,envjson",
			want: TagMetadata{
				Name:     "PORT",
				Required: true,
				Default:  "8080",
				Prefix:   "API_",
				EnvJSON:  true,
			},
		},
		{
			name: "whitespace handling",
			tag:  "  PORT  ,  required  ,  default = 8080  ,  prefix = API_  ,  envjson  ",
			want: TagMetadata{
				Name:     "PORT",
				Required: true,
				Default:  "8080",
				Prefix:   "API_",
				EnvJSON:  true,
			},
		},
		{
			name: "empty tag",
			tag:  "",
			want: TagMetadata{
				Name: "",
			},
		},
		{
			name:    "invalid format (key-value without key)",
			tag:     "PORT,=8080",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseTag(tc.tag)
			if (err != nil) != tc.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ParseTag() got = %+v, want %+v", got, tc.want)
			}
		})
	}
}
