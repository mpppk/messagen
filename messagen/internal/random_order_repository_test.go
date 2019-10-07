package internal

import (
	"strings"
	"testing"
)

func TestDefinitionRepository_Generate2(t *testing.T) {
	type fields struct {
		definitions     []*RawDefinition
		templatePickers []TemplatePicker
	}
	type args struct {
		defType      DefinitionType
		initialState *State
	}

	tests := []struct {
		name         string
		fields       fields
		args         args
		wantAnyOrder []Message
		wantErr      bool
	}{
		{
			name: "alias",
			fields: fields{
				definitions: []*RawDefinition{
					{
						Type:         "Test",
						RawTemplates: []RawTemplate{"{{.NestTest}}{{.AliasNestTest}}{{.AnotherNestTest}}"},
						Aliases: Aliases{
							"AliasNestTest":   &Alias{ReferType: "NestTest", AllowDuplicate: false},
							"AnotherNestTest": &Alias{ReferType: "NestTest", AllowDuplicate: false},
						},
					},
					{
						Type:         "NestTest",
						RawTemplates: []RawTemplate{"aaa"},
					},
					{
						Type:         "NestTest",
						RawTemplates: []RawTemplate{"bbb"},
					},
					{
						Type:         "NestTest",
						RawTemplates: []RawTemplate{"ccc"},
					},
				},
				templatePickers: []TemplatePicker{NotAllowAliasDuplicateTemplatePicker, AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			wantAnyOrder: []Message{"aaa", "bbb", "ccc"},
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDefinitionRepository(&DefinitionRepositoryOption{
				TemplatePickers:   tt.fields.templatePickers,
				DefinitionPickers: nil,
			})

			if err := d.Add(tt.fields.definitions...); (err != nil) != tt.wantErr {
				t.Errorf("DefinitionRepository.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := d.Generate(tt.args.defType, tt.args.initialState, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefinitionRepository.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, want := range tt.wantAnyOrder {
				if !strings.Contains(string(got[0]), string(want)) {
					t.Errorf("DefinitionRepository.Generate() = %v, want %v in any order", got, tt.wantAnyOrder)
				}
			}
		})
	}
}
