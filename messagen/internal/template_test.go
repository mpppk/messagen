package internal

import (
	"reflect"
	"testing"
	"text/template"
)

func newTemplateOrPanic(rawTemplate RawTemplate) *Template {
	t, err := NewTemplate(rawTemplate)
	if err != nil {
		panic(err)
	}
	return t
}

func newTemplatesOrPanic(rawTemplates ...RawTemplate) Templates {
	templates, err := NewTemplates(rawTemplates)
	if err != nil {
		panic(err)
	}
	return templates
}

func TestRawTemplate_extractDefRefIDFromRawTemplate(t *testing.T) {
	tests := []struct {
		name             string
		r                RawTemplate
		wantDefRefIDList []DefinitionType
	}{
		{
			name:             "should extract RefID",
			r:                "{{.id1}}test{{.id2}}",
			wantDefRefIDList: []DefinitionType{"id1", "id2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDefRefIDList := tt.r.extractDefRefIDFromRawTemplate(); !reflect.DeepEqual(gotDefRefIDList, tt.wantDefRefIDList) {
				t.Errorf("RawTemplate.extractDefRefIDFromRawTemplate() = %v, want %v", gotDefRefIDList, tt.wantDefRefIDList)
			}
		})
	}
}

func TestNewTemplate(t *testing.T) {
	type args struct {
		rawTemplate RawTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    *Template
		wantErr bool
	}{
		{
			name: "should have valid depends ref Type",
			args: args{
				rawTemplate: "{{.id1}}test{{.id2}}",
			},
			want: &Template{
				Raw:     "{{.id1}}test{{.id2}}",
				Depends: []DefinitionType{"id1", "id2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplate(tt.args.rawTemplate)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil {
				if tt.want != nil {
					t.Errorf("NewTemplate() = nil, want %#v", tt.want)
				}
				return
			}

			if !reflect.DeepEqual(got.Raw, tt.want.Raw) {
				t.Errorf("NewTemplate().Raw = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(got.Depends, tt.want.Depends) {
				t.Errorf("NewTemplate().Raw = %#v, want %#v", got, tt.want)
			}
			if got.tmpl == nil {
				t.Errorf("NewTemplate().tmpl is nil")
			}
		})
	}
}

func TestNewTemplates(t *testing.T) {
	type args struct {
		rawTemplates []RawTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    Templates
		wantErr bool
	}{
		{
			name: "should convert template rawTemplates to Templates",
			args: args{
				rawTemplates: []RawTemplate{"{{.id1}}", "{{.id2}}"},
			},
			want: Templates{
				newTemplateOrPanic("{{.id1}}"),
				newTemplateOrPanic("{{.id2}}"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplates(tt.args.rawTemplates)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_Execute(t *testing.T) {
	type fields struct {
		Raw     RawTemplate
		Depends []DefinitionType
		tmpl    *template.Template
	}
	type args struct {
		state State
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				Raw: "aaa{{.id1}}ccc",
			},
			args: args{
				state: State{"id1": "bbb"},
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewTemplate(tt.fields.Raw)
			if err != nil {
				t.Errorf("failed to create new template: error = %v", err)
			}
			got, err := r.Execute(tt.args.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("Template.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Template.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplates_DeleteByIndex(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name string
		t    Templates
		args args
		want Templates
	}{
		{
			name: "Delete last one element",
			t:    newTemplatesOrPanic("a"),
			args: args{
				i: 0,
			},
			want: newTemplatesOrPanic(),
		},
		{
			name: "Delete first element",
			t:    newTemplatesOrPanic("a", "b", "c"),
			args: args{
				i: 0,
			},
			want: newTemplatesOrPanic("b", "c"),
		},
		{
			name: "Delete middle element",
			t:    newTemplatesOrPanic("a", "b", "c"),
			args: args{
				i: 1,
			},
			want: newTemplatesOrPanic("a", "c"),
		},
		{
			name: "Delete last element",
			t:    newTemplatesOrPanic("a", "b", "c"),
			args: args{
				i: 2,
			},
			want: newTemplatesOrPanic("a", "b"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.DeleteByIndex(tt.args.i)
			if len(tt.t) != len(tt.want) {
				t.Errorf("After Templates.DeleteByIndex() actual = %v (len:%d), want %v (len:%d)", tt.t, len(tt.t), tt.want, len(tt.want))
				return
			}

			if len(tt.t) == 0 {
				return
			}

			if !reflect.DeepEqual(tt.t, tt.want) {
				t.Errorf("After Templates.DeleteByIndex() = %v, want %v", tt.t, tt.want)
			}
		})
	}
}

func TestTemplates_PopRandom(t *testing.T) {
	tests := []struct {
		name   string
		t      Templates
		want   *Template
		wantOk bool
	}{
		{
			name:   "",
			t:      newTemplatesOrPanic("a"),
			want:   newTemplateOrPanic("a"),
			wantOk: true,
		},
		{
			name:   "",
			t:      newTemplatesOrPanic(),
			want:   nil,
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := tt.t.PopRandom()
			if gotOk != tt.wantOk {
				t.Errorf("PopRandom() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if !gotOk {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PopRandom() got = %v, want %v", got, tt.want)
			}
		})
	}
}
