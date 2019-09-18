package messagen

import "testing"

func TestDefinitionRepository_Generate(t *testing.T) {
	type fields struct {
		m DefinitionMap
	}
	type args struct {
		id DefinitionType
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "no variable in template",
			fields: fields{
				m: DefinitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:           "Test",
						RawTemplates:   []RawTemplate{"aaa"},
						Labels:         Labels{},
						Requires:       Labels{},
						Alias:          Alias{},
						AllowDuplicate: false,
						Weight:         0,
					})},
				},
			},
			args: args{
				id: "Test",
			},
			want:    "aaa",
			wantErr: false,
		},
		{
			name: "one variable in template",
			fields: fields{
				m: DefinitionMap{
					"Test": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:           "Test",
						RawTemplates:   []RawTemplate{"aaa{{.NestTest}}ccc"},
						Labels:         Labels{},
						Requires:       Labels{},
						Alias:          Alias{},
						AllowDuplicate: false,
						Weight:         0,
					})},
					"NestTest": []*Definition{newDefinitionOrPanic(&RawDefinition{
						Type:           "NestTest",
						RawTemplates:   []RawTemplate{"bbb"},
						Labels:         Labels{},
						Requires:       Labels{},
						Alias:          Alias{},
						AllowDuplicate: false,
						Weight:         0,
					})},
				},
			},
			args: args{
				id: "Test",
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefinitionRepository{
				m: tt.fields.m,
			}
			got, err := d.Generate(tt.args.id)
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
