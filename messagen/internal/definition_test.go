package internal

import (
	"reflect"
	"testing"
)

func newDefinitionWithAliasOrPanic(rawDefinition *RawDefinition, aliasName AliasName, alias *Alias) *DefinitionWithAlias {
	def := newDefinitionOrPanic(rawDefinition)
	return &DefinitionWithAlias{
		Definition: def,
		aliasName:  aliasName,
		alias:      alias,
	}
}

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
		state *State
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
					RawConstraints: RawConstraints{"Key": "Value"},
					Aliases:        Aliases{},
					AllowDuplicate: false,
					Weight:         0,
				},
			),
			args: args{
				state: NewState(MessageMap{"Key": "Value"}),
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
					RawConstraints: RawConstraints{"Key": "Value"},
					Aliases:        Aliases{},
					AllowDuplicate: false,
					Weight:         0,
				},
			),
			args: args{
				state: NewState(MessageMap{"Key": "OtherValue"}),
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

func TestNewDefinition(t *testing.T) {
	type args struct {
		rawDefinition *RawDefinition
	}
	tests := []struct {
		name    string
		args    args
		want    *Definition
		wantErr bool
	}{
		{
			name: "",
			args: args{
				rawDefinition: &RawDefinition{
					Type:         "Root",
					RawTemplates: []RawTemplate{"{{.aaa}} {{.bbb}}"},
					Order:        []DefinitionType{"bbb"},
				},
			},
			want: &Definition{
				RawDefinition: &RawDefinition{
					Type:         "Root",
					RawTemplates: []RawTemplate{"{{.aaa}} {{.bbb}}"},
					Order:        []DefinitionType{"bbb"},
				},
				ID:        0,
				Templates: newTemplatesOrPanic([]DefinitionType{"bbb"}, "{{.aaa}} {{.bbb}}"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDefinition(tt.args.rawDefinition)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Templates[0].Depends, tt.want.Templates[0].Depends) {
				t.Errorf("NewDefinition() = %#v, want %v", *got, *tt.want)
			}

		})
	}
}
