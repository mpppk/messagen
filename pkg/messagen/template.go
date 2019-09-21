package messagen

import (
	"bytes"
	"math/rand"
	"regexp"
	"text/template"

	"golang.org/x/xerrors"
)

type RawTemplate string

func (r RawTemplate) extractDefRefIDFromRawTemplate() (defRefIDList []DefinitionType) {
	re := regexp.MustCompile(`\{\{\.(.*?)\}\}`)
	for _, match := range re.FindAllStringSubmatch(string(r), -1) {
		defRefIDList = append(defRefIDList, DefinitionType(match[1]))
	}
	return
}

type Template struct {
	Raw     RawTemplate
	Depends []DefinitionType
	tmpl    *template.Template
}

func NewTemplate(rawTemplate RawTemplate) (*Template, error) {
	ids := rawTemplate.extractDefRefIDFromRawTemplate()
	tmpl, err := template.New(string(rawTemplate)).Parse(string(rawTemplate))
	if err != nil {
		return nil, xerrors.Errorf("failed to create new template: %w", err)
	}
	return &Template{
		Raw:     rawTemplate,
		Depends: ids,
		tmpl:    tmpl,
	}, err
}

func (r *Template) Execute(state State) (Message, error) {
	buf := &bytes.Buffer{}
	if err := r.tmpl.Execute(buf, state); err != nil {
		return "", xerrors.Errorf("failed to execute template. template:%s  state:%#v : %w", r.Raw, state, err)
	}
	return Message(buf.String()), nil
}

type Templates []*Template

func NewTemplates(rawTemplates []RawTemplate) (Templates, error) {
	var templates []*Template
	for _, rawTemplate := range rawTemplates {
		t, err := NewTemplate(rawTemplate)
		if err != nil {
			return nil, xerrors.Errorf("failed to create Templates: %w", err)
		}
		templates = append(templates, t)
	}
	return Templates(templates), nil
}

func (t Templates) GetRandom() *Template {
	i := rand.Intn(len(t))
	return t[i]
}
