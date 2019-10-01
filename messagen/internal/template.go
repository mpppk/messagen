package internal

import (
	"bytes"
	"math/rand"
	"regexp"
	"text/template"

	"golang.org/x/xerrors"
)

type RawTemplate string

func (r RawTemplate) extractDefRefTypeFromRawTemplate() (defTypes DefinitionTypes) {
	re := regexp.MustCompile(`\{\{\.(.*?)\}\}`)
	for _, match := range re.FindAllStringSubmatch(string(r), -1) {
		defTypes = append(defTypes, DefinitionType(match[1]))
	}
	return
}

type DefinitionTypes []DefinitionType

func (d *DefinitionTypes) popByIndex(index int) DefinitionType {
	def := (*d)[index]
	d.deleteByIndex(index)
	return def
}

func (d *DefinitionTypes) deleteByIndex(i int) {
	if i == 0 {
		*d = (*d)[1:]
		return
	}
	if len(*d)-1 == i {
		*d = (*d)[:len(*d)-1]
		return
	}
	*d = append((*d)[:i], (*d)[i+1:]...)
}

func (d *DefinitionTypes) sortByOrderBy(orderBy DefinitionTypes) {
	var defs []DefinitionType
	for _, o := range orderBy {
		for i, def := range *d {
			if def == o {
				defs = append(defs, d.popByIndex(i))
			}
		}
	}
	*d = append(defs, *d...)
}

func (d *DefinitionTypes) copy() DefinitionTypes {
	dst := make([]DefinitionType, len(*d))
	copy(dst, *d)
	return dst
}

type Template struct {
	Raw     RawTemplate
	Depends *DefinitionTypes
	tmpl    *template.Template
}

func NewTemplate(rawTemplate RawTemplate, orderBy []DefinitionType) (*Template, error) {
	defTypes := rawTemplate.extractDefRefTypeFromRawTemplate()
	tmpl, err := template.New(string(rawTemplate)).Parse(string(rawTemplate))
	if err != nil {
		return nil, xerrors.Errorf("failed to create new template: %w", err)
	}

	defTypes.sortByOrderBy(orderBy)

	return &Template{
		Raw:     rawTemplate,
		Depends: &defTypes,
		tmpl:    tmpl,
	}, err
}

func (t *Template) Execute(state *State) (Message, error) {
	buf := &bytes.Buffer{}
	if err := t.tmpl.Execute(buf, state.m); err != nil {
		return "", xerrors.Errorf("failed to execute template. template:%s  state:%#v : %w", t.Raw, state, err)
	}
	return Message(buf.String()), nil
}

func (t *Template) IsSatisfiedState(state *State) bool {
	_, ok := t.GetFirstUnsatisfiedDef(state)
	return !ok
}

func (t *Template) GetFirstUnsatisfiedDef(state *State) (DefinitionType, bool) {
	for _, defType := range *t.Depends {
		if _, ok := state.Get(defType); ok {
			continue
		}
		return defType, true
	}
	return "", false
}

func (t *Template) Equals(template *Template) bool {
	return t.Raw == template.Raw
}

type Templates []*Template

func NewTemplates(rawTemplates []RawTemplate, orderBy []DefinitionType) (Templates, error) {
	var templates []*Template
	for _, rawTemplate := range rawTemplates {
		t, err := NewTemplate(rawTemplate, orderBy)
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

func (t *Templates) Has(template *Template) bool {
	for _, tmpl := range *t {
		if tmpl.Equals(template) {
			return true
		}
	}
	return false
}

func (t *Templates) Subtract(templateList ...*Template) Templates {
	templates := Templates(templateList)
	var newTemplates Templates
	for _, tmpl := range *t {
		if !(&templates).Has(tmpl) {
			newTemplates = append(newTemplates, tmpl)
		}
	}
	return newTemplates
}

func (t *Templates) Add(template *Template) {
	*t = append(*t, template)
}

func (t *Templates) Copy() (Templates, error) {
	var newRawTemplates []RawTemplate
	for _, tmpl := range *t {
		newRawTemplates = append(newRawTemplates, tmpl.Raw)
	}
	return NewTemplates(newRawTemplates, nil)
}
