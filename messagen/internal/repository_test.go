package internal

import (
	"reflect"
	"testing"
)

func TestDefinitionRepository_Generate(t *testing.T) {
	type args struct {
		defType      DefinitionType
		initialState *State
	}

	tests := []struct {
		name    string
		opt     *DefinitionRepositoryOption
		defs    []*RawDefinition
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "no variable in template",
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa"},
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
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				},
				{
					Type:         "NestTest",
					RawTemplates: []RawTemplate{"bbb"},
				},
				{
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"k999": "v999"},
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
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}{{.NestTest2}}"},
				}, {
					Type:         "NestTest",
					RawTemplates: []RawTemplate{"bbb"},
				}, {
					Type:           "NestTest2",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"NestTest": "xxx"},
				}, {
					Type:           "NestTest2",
					RawTemplates:   []RawTemplate{"ccc"},
					RawConstraints: RawConstraints{"NestTest": "bbb"},
				},
			},
			args: args{
				defType: "Test",
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
		{
			name: "unresolvable template in template",
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"k999": "v999"},
				},
			},
			args: args{
				defType: "Test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unresolvable template in template2",
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"k999": "v999"},
				},
			},
			args: args{
				defType: "Test",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "unresolvable template in template2",
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				}, {
					Type:         "NestTest",
					RawTemplates: []RawTemplate{"{{.NestTest2}}"},
				}, {
					Type:         "NestTest",
					RawTemplates: []RawTemplate{"bbb"},
				}, {
					Type:         "NestTest2",
					RawTemplates: []RawTemplate{"{{.NoExistDef}}"},
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
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"k1!": "_"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"ddd"},
					RawConstraints: RawConstraints{},
				},
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
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"K1?": "V2", "K2": "V2"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"bbb"},
					RawConstraints: RawConstraints{"K1?": "V1", "K2": "V2", "K3?": "V3"},
				},
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
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}{{.NestTest2}}"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"bbb"},
					RawConstraints: RawConstraints{"K1+": "V1"},
				}, {
					Type:           "NestTest2",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"K1!": "_"},
				}, {
					Type:           "NestTest2",
					RawTemplates:   []RawTemplate{"ccc"},
					RawConstraints: RawConstraints{"K1": "V1"},
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
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"aaa{{.NestTest}}ccc"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"bbb"},
					RawConstraints: RawConstraints{"K1/": ".?1"},
				}, {
					Type:           "NestTest",
					RawTemplates:   []RawTemplate{"xxx"},
					RawConstraints: RawConstraints{"K1/": ".?2"},
				},
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
			d := NewDefinitionRepository(tt.opt)
			if err := d.Add(tt.defs...); err != nil {
				t.Errorf("unexpected error occurred in DefinitionRepository.Add(): %s", err)
			}
			got, err := d.Generate(tt.args.defType, tt.args.initialState, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefinitionRepository.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got[0] != tt.want {
				t.Errorf("DefinitionRepository.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefinitionRepository_Generate_WithValidator(t *testing.T) {
	type args struct {
		defType      DefinitionType
		initialState *State
	}

	tests := []struct {
		name    string
		opt     *DefinitionRepositoryOption
		defs    []*RawDefinition
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "maxStrLen validator",
			opt: &DefinitionRepositoryOption{
				TemplateValidators: []TemplateValidator{MaxStrLenValidator(2)},
			},
			defs: []*RawDefinition{
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"xxx"},
				},
				{
					Type:         "Test",
					RawTemplates: []RawTemplate{"{{.NestTest}}"},
				},
				{
					Type:         "NestTest",
					RawTemplates: []RawTemplate{"yyy"},
				},
				{
					Type:         "NestTest",
					RawTemplates: []RawTemplate{"aa"},
				},
			},
			args: args{
				defType: "Test",
			},
			want:    "aa",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDefinitionRepository(tt.opt)
			if err := d.Add(tt.defs...); err != nil {
				t.Errorf("unexpected error occurred in DefinitionRepository.Add(): %s", err)
			}
			got, err := d.Generate(tt.args.defType, tt.args.initialState, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefinitionRepository.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got[0] != tt.want {
				t.Errorf("DefinitionRepository.Generate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomTemplatePicker(t *testing.T) {
	def := newDefinitionWithAliasOrPanic(&RawDefinition{
		Type:         "Test",
		RawTemplates: []RawTemplate{"a"},
	}, "", nil)
	type args struct {
		def   *DefinitionWithAlias
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
