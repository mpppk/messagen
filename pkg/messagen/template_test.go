package messagen

import (
	"reflect"
	"testing"
	"text/template"
)

func TestRawTemplate_extractDefRefIDFromRawTemplate(t *testing.T) {
	tests := []struct {
		name             string
		r                RawTemplate
		wantDefRefIDList []DefinitionID
	}{
		{
			name:             "should extract RefID",
			r:                "{{.id1}}test{{.id2}}",
			wantDefRefIDList: []DefinitionID{"id1", "id2"},
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
			name: "should have valid depends ref ID",
			args: args{
				rawTemplate: "{{.id1}}test{{.id2}}",
			},
			want: &Template{
				Raw:     "{{.id1}}test{{.id2}}",
				Depends: []DefinitionID{"id1", "id2"},
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
	templates := []*Template{
		newTemplateOrPanic("{{.id1}}"),
		newTemplateOrPanic("{{.id2}}"),
	}

	type args struct {
		slice []*Template
	}
	tests := []struct {
		name string
		args args
		want Templates
	}{
		{
			name: "should convert template slice to Templates",
			args: args{
				slice: templates,
			},
			want: Templates(templates),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTemplates(tt.args.slice); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_Execute(t *testing.T) {
	type fields struct {
		Raw     RawTemplate
		Depends []DefinitionID
		tmpl    *template.Template
	}
	type args struct {
		m GeneratedMessageMap
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    GeneratedMessage
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				Raw: "aaa{{.id1}}ccc",
			},
			args: args{
				m: GeneratedMessageMap{"id1": "bbb"},
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
			//{
			//	Raw:     tt.fields.Raw,
			//	Depends: tt.fields.Depends,
			//	tmpl:    tt.fields.tmpl,
			//}
			got, err := r.Execute(tt.args.m)
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
