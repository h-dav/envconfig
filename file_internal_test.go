package envconfig

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_identifyParser(t *testing.T) {
	type testCase struct {
		filepath string
		want     parser
		wantErr  error
	}

	testCases := map[string]testCase{
		"expect env parser for env file": {
			filepath: "example.env",
			want:     envFileParser{},
		},
		"expect error due to invalid file extension": {
			filepath: "example.invalid",
			wantErr: &FileTypeValidationError{
				Filepath: "example.invalid",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn,
			func(t *testing.T) {
				t.Parallel()

				got, err := identifyFileParser(tc.filepath)

				if !cmp.Equal(tc.wantErr, err) {
					t.Errorf("wantErr: %#v, got: %#v", tc.wantErr, err)
				}

				if reflect.TypeOf(tc.want) != reflect.TypeOf(got) {
					t.Errorf(
						"want: %q, got: %q",
						reflect.TypeOf(tc.wantErr),
						reflect.TypeOf(got),
					)
				}
			},
		)
	}
}
