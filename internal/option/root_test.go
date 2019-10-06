package option_test

import (
	"reflect"
	"testing"

	"github.com/mpppk/messagen/internal/option"
)

func Test_newCmdConfigFromRawConfig(t *testing.T) {
	type args struct {
		rawConfig *option.RootCmdRawConfig
	}
	tests := []struct {
		name string
		args args
		want *option.RootCmdConfig
	}{
		{
			name: "Toggle property should have false if RootCmdRawConfig has false",
			args: args{
				rawConfig: &option.RootCmdRawConfig{},
			},
			want: &option.RootCmdConfig{},
		},
		{
			name: "Toggle property should have true if RootCmdRawConfig has true",
			args: args{
				rawConfig: &option.RootCmdRawConfig{},
			},
			want: &option.RootCmdConfig{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := option.NewCmdConfigFromRawConfig(tt.args.rawConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newRootCmdConfigFromRawConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
