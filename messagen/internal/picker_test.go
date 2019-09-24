package internal

import (
	"reflect"
	"testing"
)

func newDefinitionsOrPanic(rawDefs ...*RawDefinition) Definitions {
	definitions, err := NewDefinitions(rawDefs...)
	if err != nil {
		panic(err)
	}
	return definitions
}

func TestRandomWithWeightDefinitionPicker(t *testing.T) {
	definitions := newDefinitionsOrPanic([]*RawDefinition{
		{
			Type:           "Root",
			RawTemplates:   []RawTemplate{"a"},
			Constraints:    nil,
			Alias:          nil,
			AllowDuplicate: false,
			Weight:         0,
		},
		{
			Type:           "Root",
			RawTemplates:   []RawTemplate{"b"},
			Constraints:    nil,
			Alias:          nil,
			AllowDuplicate: false,
			Weight:         0,
		},
	}...)
	type args struct {
		definitions *Definitions
		state       State
	}
	tests := []struct {
		name    string
		args    args
		want    []*Definition
		wantErr bool
	}{
		{
			name: "",
			args: args{
				definitions: &definitions,
				state:       State{},
			},
			want:    definitions,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomWithWeightDefinitionPicker(tt.args.definitions, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomWithWeightDefinitionPicker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RandomWithWeightDefinitionPicker() got = %v, want %v", got, tt.want)
			}
		})
	}
}
