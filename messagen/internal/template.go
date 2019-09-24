package internal

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
	t := Templates(templates)
	return t, nil
}

func (t *Templates) PopRandom() (*Template, bool) {
	if len(*t) == 0 {
		return nil, false
	}
	i := rand.Intn(len(*t))
	tmpl := (*t)[i]
	t.DeleteByIndex(i)
	return tmpl, true
}

func (t *Templates) DeleteByIndex(i int) {
	if i == 0 {
		*t = (*t)[1:]
		return
	}
	if len(*t)-1 == i {
		*t = (*t)[:len(*t)-1]
		return
	}
	*t = append((*t)[:i], (*t)[i+1:]...)
}

func (t *Templates) Copy() (Templates, error) {
	var newRawTemplates []RawTemplate
	for _, tmpl := range *t {
		newRawTemplates = append(newRawTemplates, tmpl.Raw)
	}
	return NewTemplates(newRawTemplates)
}
