package messagen

import (
	"reflect"
	"testing"
)

func TestParseYaml(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "",
			args: args{
				filePath: "../testdata/hello.yaml",
			},
			want: &Config{
				Definitions: []Definition{
					{
						Type:           "Root",
						Templates:      []string{"hello"},
						Constraints:    nil,
						Aliases:        nil,
						AllowDuplicate: false,
						OrderBy:        nil,
						Weight:         0,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseYaml(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseYaml() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
