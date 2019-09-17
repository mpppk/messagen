package messagen

import "testing"

func TestDefinitionRepository_Generate(t *testing.T) {
	type fields struct {
		m DefinitionMap
	}
	type args struct {
		id DefinitionID
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
					"test": &Definition{
						RawDefinition: RawDefinition{
							ID:             "test",
							AllowDuplicate: false,
						},
						Templates: NewTemplates([]*Template{newTemplateOrPanic("aaa")}),
					},
				},
			},
			args: args{
				id: "test",
			},
			want:    "aaa",
			wantErr: false,
		},
		{
			name: "one variable in template",
			fields: fields{
				m: DefinitionMap{
					"test": &Definition{
						RawDefinition: RawDefinition{
							ID:             "test",
							AllowDuplicate: false,
						},
						Templates: NewTemplates([]*Template{newTemplateOrPanic("aaa{{.NestTest}}ccc")}),
					},
					"NestTest": &Definition{
						RawDefinition: RawDefinition{
							ID:             "NestTest",
							AllowDuplicate: false,
						},
						Templates: NewTemplates([]*Template{newTemplateOrPanic("bbb")}),
					},
				},
			},
			args: args{
				id: "test",
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
