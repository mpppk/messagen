package messagen

import (
	"bytes"
	"math/rand"
	"regexp"
	"text/template"

	"golang.org/x/xerrors"
)

type RawTemplate string

func (r RawTemplate) extractDefRefIDFromRawTemplate() (defRefIDList []DefinitionID) {
	re := regexp.MustCompile(`\{\{\.(.*?)\}\}`)
	for _, match := range re.FindAllStringSubmatch(string(r), -1) {
		defRefIDList = append(defRefIDList, DefinitionID(match[1]))
	}
	return
}

type Template struct {
	Raw     RawTemplate
	Depends []DefinitionID
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

func newTemplateOrPanic(rawTemplate RawTemplate) *Template {
	t, err := NewTemplate(rawTemplate)
	if err != nil {
		panic(err)
	}
	return t
}

func (r *Template) Execute(m GeneratedMessageMap) (GeneratedMessage, error) {
	buf := &bytes.Buffer{}
	if err := r.tmpl.Execute(buf, m); err != nil {
		return "", xerrors.Errorf("failed to execute template. template:%s  m:%#v : %w", r.Raw, m, err)
	}
	return GeneratedMessage(buf.String()), nil
}

type Templates []*Template

func NewTemplates(slice []*Template) Templates {
	return Templates(slice)
}

func (t Templates) GetRandom() *Template {
	i := rand.Intn(len(t))
	return t[i]
}
