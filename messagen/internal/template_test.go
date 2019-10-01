package internal

import (
	"reflect"
	"testing"
	"text/template"
)

func newTemplateOrPanic(rawTemplate RawTemplate, orderBy []DefinitionType) *Template {
	t, err := NewTemplate(rawTemplate, orderBy)
	if err != nil {
		panic(err)
	}
	return t
}

func newTemplatesOrPanic(orderBy []DefinitionType, rawTemplates ...RawTemplate) Templates {
	templates, err := NewTemplates(rawTemplates, orderBy)
	if err != nil {
		panic(err)
	}
	return templates
}

func TestRawTemplate_extractDefRefIDFromRawTemplate(t *testing.T) {
	tests := []struct {
		name            string
		r               RawTemplate
		wantDefRefTypes DefinitionTypes
	}{
		{
			name:            "should extract RefID",
			r:               "{{.id1}}test{{.id2}}",
			wantDefRefTypes: DefinitionTypes{"id1", "id2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDefRefTypes := tt.r.extractDefRefTypeFromRawTemplate(); !reflect.DeepEqual(gotDefRefTypes, tt.wantDefRefTypes) {
				t.Errorf("RawTemplate.extractDefRefTypeFromRawTemplate() = %v, want %v", gotDefRefTypes, tt.wantDefRefTypes)
			}
		})
	}
}

func TestNewTemplate(t *testing.T) {
	type args struct {
		rawTemplate RawTemplate
		orderBy     []DefinitionType
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
				Depends: &DefinitionTypes{"id1", "id2"},
			},
		},
		{
			name: "should consider OrderBy",
			args: args{
				rawTemplate: "{{.id1}}test{{.id2}}",
				orderBy:     DefinitionTypes{"id2", "id1"},
			},
			want: &Template{
				Raw:     "{{.id1}}test{{.id2}}",
				Depends: &DefinitionTypes{"id2", "id1"},
			},
		},
		{
			name: "should consider OrderBy",
			args: args{
				rawTemplate: "{{.id1}}{{.id2}}{{.id3}}",
				orderBy:     DefinitionTypes{"id3", "id1"},
			},
			want: &Template{
				Raw:     "{{.id1}}{{.id2}}{{.id3}}",
				Depends: &DefinitionTypes{"id3", "id1", "id2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplate(tt.args.rawTemplate, tt.args.orderBy)
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
				newTemplateOrPanic("{{.id1}}", nil),
				newTemplateOrPanic("{{.id2}}", nil),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTemplates(tt.args.rawTemplates, nil)
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
		state *State
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
				state: NewState(MessageMap{"id1": "bbb"}),
			},
			want:    "aaabbbccc",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewTemplate(tt.fields.Raw, nil)
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
			t:    newTemplatesOrPanic(nil, "a"),
			args: args{
				i: 0,
			},
			want: newTemplatesOrPanic(nil),
		},
		{
			name: "Delete first element",
			t:    newTemplatesOrPanic(nil, "a", "b", "c"),
			args: args{
				i: 0,
			},
			want: newTemplatesOrPanic(nil, "b", "c"),
		},
		{
			name: "Delete middle element",
			t:    newTemplatesOrPanic(nil, "a", "b", "c"),
			args: args{
				i: 1,
			},
			want: newTemplatesOrPanic(nil, "a", "c"),
		},
		{
			name: "Delete last element",
			t:    newTemplatesOrPanic(nil, "a", "b", "c"),
			args: args{
				i: 2,
			},
			want: newTemplatesOrPanic(nil, "a", "b"),
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
			t:      newTemplatesOrPanic(nil, "a"),
			want:   newTemplateOrPanic("a", nil),
			wantOk: true,
		},
		{
			name:   "",
			t:      newTemplatesOrPanic(nil),
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

func TestTemplates_Subtract(t *testing.T) {
	type args struct {
		templateList []*Template
	}
	tests := []struct {
		name string
		t    Templates
		args args
		want Templates
	}{
		{
			name: "",
			t:    newTemplatesOrPanic(nil, "aaa", "bbb", "ccc"),
			args: args{newTemplatesOrPanic(nil, "aaa")},
			want: newTemplatesOrPanic(nil, "bbb", "ccc"),
		},
		{
			name: "",
			t:    newTemplatesOrPanic(nil, "aaa", "bbb", "ccc"),
			args: args{newTemplatesOrPanic(nil, "bbb")},
			want: newTemplatesOrPanic(nil, "aaa", "ccc"),
		},
		{
			name: "",
			t:    newTemplatesOrPanic(nil, "aaa", "bbb", "ccc"),
			args: args{newTemplatesOrPanic(nil, "ccc")},
			want: newTemplatesOrPanic(nil, "aaa", "bbb"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Subtract(tt.args.templateList...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Templates.Subtract() = %v, want %v", got, tt.want)
			}
		})
	}
}
