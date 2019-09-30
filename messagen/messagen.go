// messagen is the tree structured template engine with declarative API.
package messagen

import (
	"github.com/mpppk/messagen/messagen/internal"
)

type Definition struct {
	Type           string
	Templates      []string
	Constraints    map[string]string
	Aliases        map[string]*Alias
	AllowDuplicate bool
	OrderBy        []string
	Weight         float32
}

type Alias struct {
	ReferType      string
	AllowDuplicate bool
}

func (a *Alias) toAlias() *internal.Alias {
	return &internal.Alias{
		ReferType:      internal.DefinitionType(a.ReferType),
		AllowDuplicate: a.AllowDuplicate,
	}
}

func newAliases(aliases map[string]*Alias) internal.Aliases {
	newAliases := internal.Aliases{}
	for key, alias := range aliases {
		newAliases[internal.DefinitionType(key)] = alias.toAlias()
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

var defaultTemplatePickers = []internal.TemplatePicker{internal.RandomTemplatePicker, internal.NotAllowAliasDuplicateTemplatePicker}

type Messagen struct {
	repo *internal.DefinitionRepository
}
type Option struct {
	TemplatePickers []internal.TemplatePicker
}

func New(opt *Option) (*Messagen, error) {
	var templatePickers []internal.TemplatePicker
	if opt == nil {
		templatePickers = defaultTemplatePickers
	} else if opt.TemplatePickers == nil {
		templatePickers = defaultTemplatePickers
	} else {
		templatePickers = opt.TemplatePickers
	}

	return &Messagen{
		repo: internal.NewDefinitionRepository(
			&internal.DefinitionRepositoryOption{
				TemplatePickers: templatePickers,
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

func (m *Messagen) Generate(defType string, state map[string]string) (string, error) {
	msg, err := m.repo.Generate(internal.DefinitionType(defType), newState(state))
	return string(msg), err
}

func newState(s map[string]string) *internal.State {
	state := internal.NewState(nil)
	for key, value := range s {
		state.Set(internal.DefinitionType(key), internal.Message(value))
	}
	return state
}
