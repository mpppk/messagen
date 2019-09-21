package internal

import (
	"testing"
)

func TestDefinitionRepository_Generate(t *testing.T) {
	type fields struct {
		m definitionMap
	}
	type args struct {
		defType      DefinitionType
		initialState State
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
			},
			args: args{
				defType:      "Test",
				initialState: State{"k1": "v1"},
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
			},
			args: args{
				defType:      "Test",
				initialState: State{"K1": "V1", "K2": "V2"},
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
			},
			args: args{
				defType:      "Test",
				initialState: State{"K1": "V1"},
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if tt.name != "two variable in template" {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			d := &DefinitionRepository{
				m: tt.fields.m,
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
