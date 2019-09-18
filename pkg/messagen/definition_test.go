package messagen

import "testing"

func newDefinitionOrPanic(rawDefinition *RawDefinition) *Definition {
	def, err := NewDefinition(rawDefinition)
	if err != nil {
		panic(err)
	}
	return def
}

func TestDefinition_CanBePicked(t *testing.T) {
	type args struct {
		labels Labels
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
					Labels:         Labels{},
					Requires:       Labels{"Key": "Value"},
					Alias:          Alias{},
					AllowDuplicate: false,
					Weight:         0,
				},
			),
			args: args{
				labels: Labels{"Key": "Value"},
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
					Labels:         Labels{},
					Requires:       Labels{"Key": "Value"},
					Alias:          Alias{},
					AllowDuplicate: false,
					Weight:         0,
				},
			),
			args: args{
				labels: Labels{"Key": "OtherValue"},
			},
			want: false,
			//want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := tt.definition.CanBePicked(tt.args.labels)
			if got != tt.want {
				t.Errorf("Definition.CanBePicked() got = %v, want %v", got, tt.want)
			}
			//if got1 != tt.want1 {
			//	t.Errorf("Definition.CanBePicked() got1 = %v, want %v", got1, tt.want1)
			//}
		})
	}
}
