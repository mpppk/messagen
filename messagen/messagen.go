// messagen is the tree structured template engine with declarative API.
package messagen

import (
	"github.com/mpppk/messagen/messagen/internal"
)

type Definition struct {
	Type           string            `yaml:"Type"`
	Templates      []string          `yaml:"Templates"`
	Constraints    map[string]string `yaml:"Constraints"`
	Aliases        map[string]*Alias `yaml:"Aliases"`
	AllowDuplicate bool              `yaml:"AllowDuplicate"`
	OrderBy        []string          `yaml:"OrderBy"`
	Weight         float32           `yaml:"Weight"`
}

type Alias struct {
	Type           string `yaml:"Type"`
	AllowDuplicate bool   `yaml:"AllowDuplicate"`
}

func (a *Alias) toAlias() *internal.Alias {
	return &internal.Alias{
		ReferType:      internal.DefinitionType(a.Type),
		AllowDuplicate: a.AllowDuplicate,
	}
}

func newAliases(aliases map[string]*Alias) internal.Aliases {
	newAliases := internal.Aliases{}
	for key, alias := range aliases {
		newAliases[internal.AliasName(key)] = alias.toAlias()
	}
	return newAliases
}

func (d *Definition) toRawDefinition() (*internal.RawDefinition, error) {
	var rawTemplates []internal.RawTemplate
	for _, rt := range d.Templates {
		rawTemplates = append(rawTemplates, internal.RawTemplate(rt))
	}
	rawConstraints := internal.RawConstraints{}
	for key, value := range d.Constraints {
		rawConstraints[internal.RawConstraintKey(key)] = internal.RawConstraintValue(value)
	}
	constraints, err := internal.NewConstraints(rawConstraints)
	if err != nil {
		return nil, err
	}

	return &internal.RawDefinition{
		Type:           internal.DefinitionType(d.Type),
		RawTemplates:   rawTemplates,
		Constraints:    constraints,
		AllowDuplicate: d.AllowDuplicate,
		Aliases:        newAliases(d.Aliases),
		OrderBy:        d.getOrderBy(),
		Weight:         internal.DefinitionWeight(d.Weight),
	}, nil
}

func (d *Definition) getOrderBy() (orderBy []internal.DefinitionType) {
	for _, o := range d.OrderBy {
		orderBy = append(orderBy, internal.DefinitionType(o))
	}
	return
}

type Messagen struct {
	repo *internal.DefinitionRepository
}
type Option struct {
	TemplatePickers    []internal.TemplatePicker
	DefinitionPickers  []internal.DefinitionPicker
	TemplateValidators []internal.TemplateValidator
}

func New(opt *Option) (*Messagen, error) {
	templatePickers := []internal.TemplatePicker{internal.RandomTemplatePicker}
	if opt != nil && opt.TemplatePickers != nil {
		templatePickers = opt.TemplatePickers
	}

	var definitionPickers []internal.DefinitionPicker
	if opt != nil && opt.DefinitionPickers != nil {
		definitionPickers = opt.DefinitionPickers
	}

	var templateValidators []internal.TemplateValidator
	if opt != nil && opt.TemplateValidators != nil {
		templateValidators = opt.TemplateValidators
	}

	return &Messagen{
		repo: internal.NewDefinitionRepository(
			&internal.DefinitionRepositoryOption{
				TemplatePickers:    templatePickers,
				DefinitionPickers:  definitionPickers,
				TemplateValidators: templateValidators,
			},
		),
	}, nil
}

func (m *Messagen) AddDefinition(defs ...*Definition) error {
	for _, def := range defs {
		rawDef, err := def.toRawDefinition()
		if err != nil {
			return err
		}
		if err := m.repo.Add(rawDef); err != nil {
			return err
		}
	}
	return nil
}

func (m *Messagen) Generate(defType string, state map[string]string, num uint) ([]string, error) {
	msgs, err := m.repo.Generate(internal.DefinitionType(defType), newState(state), num)
	if err != nil {
		return nil, err
	}
	var strMsgs []string
	for _, msg := range msgs {
		strMsgs = append(strMsgs, string(msg))
	}
	return strMsgs, err
}

func newState(s map[string]string) *internal.State {
	state := internal.NewState(nil)
	for key, value := range s {
		state.Set(internal.DefinitionType(key), internal.Message(value))
	}
	return state
}
