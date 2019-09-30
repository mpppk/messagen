package internal

import (
	"reflect"
	"testing"
)

func TestDefinitionRepository_Generate(t *testing.T) {
	type fields struct {
		m               definitionMap
		templatePickers []TemplatePicker
	}
	type args struct {
		defType      DefinitionType
		initialState *State
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "no variable in template",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa"},
					})},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "aaa",
			wantErr: false,
		},
		{
			name: "one variable in template",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"bbb"},
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"k999": "v999"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
		{
			name: "two variable in template",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}{{.NestTest2}}"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"bbb"},
						}),
					},
					"NestTest2": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest2",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"NestTest": "xxx"}),
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest2",
							RawTemplates: []RawTemplate{"ccc"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"NestTest": "bbb"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
		{
			name: "unresolvable template in template",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"k999": "v999"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unresolvable template in template2",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTestxxxxxxx": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"k999": "v999"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unresolvable template in template2",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"{{.NestTest2}}"},
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"bbb"},
						}),
					},
					"NestTest2": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest2",
							RawTemplates: []RawTemplate{"{{.NoExistDef}}"},
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
		{
			name: "use ! operator",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"k1!": "_"}),
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"ddd"},
							Constraints:  &Constraints{},
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType:      "Test",
				initialState: NewState(MessageMap{"k1": "v1"}),
			},
			want:    "aaadddccc",
			wantErr: false,
		},
		{
			name: "use ? operator",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1?": "V2", "K2": "V2"}),
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"bbb"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1?": "V1", "K2": "V2", "K3?": "V3"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType:      "Test",
				initialState: NewState(MessageMap{"K1": "V1", "K2": "V2"}),
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
		{
			name: "use + operator",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}{{.NestTest2}}"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"bbb"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1+": "V1"}),
						}),
					},
					"NestTest2": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest2",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1!": "_"}),
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest2",
							RawTemplates: []RawTemplate{"ccc"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1": "V1"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType: "Test",
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
		{
			name: "use / operator",
			fields: fields{
				m: definitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:         "Test",
						RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
					})},
					"NestTest": []*Definition{
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"bbb"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1/": ".?1"}),
						}),
						newDefinitionOrPanic(&RawDefinition{
							Type:         "NestTest",
							RawTemplates: []RawTemplate{"xxx"},
							Constraints:  newConstraintsOrPanic(RawConstraints{"K1/": ".?2"}),
						}),
					},
				},
				templatePickers: []TemplatePicker{AscendingOrderTemplatePicker},
			},
			args: args{
				defType:      "Test",
				initialState: NewState(MessageMap{"K1": "V1"}),
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefinitionRepository{
				m:               tt.fields.m,
				templatePickers: tt.fields.templatePickers,
			}
			got, err := d.Generate(tt.args.defType, tt.args.initialState)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefinitionRepository.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DefinitionRepository.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomTemplatePicker(t *testing.T) {
	def := newDefinitionOrPanic(&RawDefinition{
		Type:         "Test",
		RawTemplates: []RawTemplate{"a"},
	})
	type args struct {
		def   *Definition
		state *State
	}
	tests := []struct {
		name    string
		args    args
		want    Templates
		wantErr bool
	}{
		{
			name: "",
			args: args{
				def:   def,
				state: NewState(nil),
			},
			want:    newTemplatesOrPanic(nil, "a"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RandomTemplatePicker(tt.args.def, tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("RandomTemplatePicker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RandomTemplatePicker() got = %v, want %v", got, tt.want)
			}
		})
	}
}
