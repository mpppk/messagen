package internal

import (
	"testing"
)

func newDefinitionOrPanic(rawDefinition *RawDefinition) *Definition {
	def, err := NewDefinition(rawDefinition)
	if err != nil {
		panic(err)
	}
	return def
}

func newConstraintsOrPanic(raw RawConstraints) *Constraints {
	c, err := NewConstraints(raw)
	if err != nil {
		panic(err)
	}
	return c
}

func TestDefinition_CanBePicked(t *testing.T) {
	type args struct {
		state State
	}
	tests := []struct {
		name       string
		definition *Definition
		args       args
		want       bool
		//want1      string
	}{
		{
			name: "",
			definition: newDefinitionOrPanic(
				&RawDefinition{
					Type:           "Root",
					RawTemplates:   []RawTemplate{""},
					Constraints:    newConstraintsOrPanic(RawConstraints{"Key": "Value"}),
					Alias:          Alias{},
					AllowDuplicate: false,
					Weight:         0,
				},
			),
			args: args{
				state: State{"Key": "Value"},
			},
			want: true,
			//want1: "",
		},
		{
			name: "",
			definition: newDefinitionOrPanic(
				&RawDefinition{
					Type:           "Root",
					RawTemplates:   []RawTemplate{""},
					Constraints:    newConstraintsOrPanic(RawConstraints{"Key": "Value"}),
					Alias:          Alias{},
					AllowDuplicate: false,
					Weight:         0,
				},
			),
			args: args{
				state: State{"Key": "OtherValue"},
			},
			want: false,
			//want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.definition.CanBePicked(tt.args.state)
			if got != tt.want {
				t.Errorf("Definition.CanBePicked() got = %v, want %v", got, tt.want)
			}
			//if got1 != tt.want1 {
			//	t.Errorf("Definition.CanBePicked() got1 = %v, want %v", got1, tt.want1)
			//}
		})
	}
}
