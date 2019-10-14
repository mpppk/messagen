package internal

import (
	"math/rand"
	"reflect"
	"testing"
)

func newDefinitionsOrPanic(rawDefs ...*RawDefinition) Definitions {
	var definitions []*Definition
	for _, rawDef := range rawDefs {
		def, err := NewDefinition(rawDef)
		if err != nil {
			panic(err)
		}
		definitions = append(definitions, def)
	}
	return definitions
}

func TestRandomWithWeightDefinitionPicker(t *testing.T) {
	definitions := newDefinitionsOrPanic([]*RawDefinition{
		{
			Type:         "Root",
			RawTemplates: []RawTemplate{"a"},
		},
		{
			Type:         "Root",
			RawTemplates: []RawTemplate{"b"},
		},
	}...)
	type args struct {
		definitions *Definitions
		state       *State
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
				state:       NewState(nil),
			},
			want: newDefinitionsOrPanic([]*RawDefinition{
				{
					Type:         "Root",
					RawTemplates: []RawTemplate{"b"},
				},
				{
					Type:         "Root",
					RawTemplates: []RawTemplate{"a"},
				},
			}...),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		rand.Seed(0)
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomWithWeightDefinitionPicker(tt.args.definitions, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomWithWeightDefinitionPicker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("RandomWithWeightDefinitionPicker() got = %#v, want %#v", got, tt.want)
			}

			for i, g := range got {
				gRawTemplates := g.RawTemplates
				wantRawTemplates := tt.want[i].RawTemplates
				if !reflect.DeepEqual(gRawTemplates, wantRawTemplates) {
					t.Errorf("RandomWithWeightDefinitionPicker() got = %#v, want %#v", gRawTemplates, wantRawTemplates)
				}
			}
		})
	}
}

func TestSortByConstraintPriorityDefinitionPicker(t *testing.T) {
	tests := []struct {
		name           string
		rawDefinitions []*RawDefinition
		want           []DefinitionType
		wantErr        bool
	}{
		{
			name: "",
			rawDefinitions: []*RawDefinition{
				{
					Type:           "Root0",
					RawConstraints: RawConstraints{},
				},
				{
					Type:           "Root1",
					RawConstraints: RawConstraints{"A:3": "B"},
				},
				{
					Type:           "Root2",
					RawConstraints: RawConstraints{"A:4": "X"},
				},
				{
					Type:           "Root3",
					RawConstraints: RawConstraints{"A:2": "X", "B:2": "X"},
				},
			},
			want:    []DefinitionType{"Root2", "Root3", "Root1", "Root0"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definitions := newDefinitionsOrPanic(tt.rawDefinitions...)
			got, err := SortByConstraintPriorityDefinitionPicker(&definitions, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("SortByConstraintPriorityDefinitionPicker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, def := range got {
				if def.Type != tt.want[i] {
					t.Errorf("SortByConstraintPriorityDefinitionPicker()[%d] got = %v, want %v", i, def.Type, tt.want[i])
				}
			}
		})
	}
}
