package internal

import (
	"errors"
	"fmt"
	"math/rand"

	"golang.org/x/xerrors"
)

type definitionMap map[DefinitionType][]*Definition
type Message string
type TemplatePicker func(templates *Templates, state State) (Templates, error)

func RandomTemplatePicker(templates *Templates, state State) (Templates, error) {
	var newTemplates Templates
	for {
		if len(*templates) == 0 {
			break
		}
		tmpl, ok := templates.PopRandom()
		if !ok {
			return nil, xerrors.Errorf("failed to pop template random from %v", templates)
		}
		newTemplates = append(newTemplates, tmpl)
	}
	return newTemplates, nil
}

func AscendingOrderTemplatePicker(templates *Templates, state State) (Templates, error) {
	return *templates, nil
}

type DefinitionRepository struct {
	m               definitionMap
	templatePickers []TemplatePicker
}

type DefinitionRepositoryOption struct {
	TemplatePickers []TemplatePicker
}

func NewDefinitionRepository(opt *DefinitionRepositoryOption) *DefinitionRepository {
	templatePickers := opt.TemplatePickers
	if templatePickers == nil {
		templatePickers = []TemplatePicker{}
	}
	return &DefinitionRepository{
		m:               definitionMap{},
		templatePickers: templatePickers,
	}
}

func (d *DefinitionRepository) List(defType DefinitionType) (defs []*Definition) {
	defs, ok := d.m[defType]
	if !ok {
		return []*Definition{}
	}
	return defs
}

func (d *DefinitionRepository) pickRandom(defType DefinitionType) (*Definition, bool) {
	defs, ok := d.m[defType]
	if !ok {
		return nil, false
	}
	return defs[rand.Intn(len(defs))], true
}

func (d *DefinitionRepository) Add(rawDefs ...*RawDefinition) error {
	for _, rawDefinition := range rawDefs {
		def, err := NewDefinition(rawDefinition)
		if err != nil {
			return xerrors.Errorf("failed to add definition to repository: %w", err)
		}
		d.addDefinition(def)
	}
	return nil
}

func (d *DefinitionRepository) addDefinition(def *Definition) {
	if defs, ok := d.m[def.Type]; ok {
		d.m[def.Type] = append(defs, def)
		return
	}
	d.m[def.Type] = []*Definition{def}
	return
}

func (d *DefinitionRepository) Generate(defType DefinitionType, initialState State) (Message, error) {
	if initialState == nil {
		initialState = State{}
	}
	def, ok := d.pickRandom(defType)
	if !ok {
		return "", xerrors.Errorf("failed to generate message. Root Definition type not found: %s", defType)
	}

	msg, err := generate(def, initialState, d)
	return msg, err
}

func (d *DefinitionRepository) applyTemplatePickers(templates Templates, state State) (newTemplates Templates, err error) {
	newTemplates, err = (&templates).Copy()
	if err != nil {
		return nil, err
	}
	for _, picker := range d.templatePickers {
		if len(newTemplates) == 0 {
			return Templates{}, nil
		}
		newTemplates, err = picker(&newTemplates, state)
		if err != nil {
			return nil, err
		}
	}
	return newTemplates, nil
}

func generate(def *Definition, state State, repo *DefinitionRepository) (Message, error) {
	templates, err := repo.applyTemplatePickers(def.Templates, state)
	if err != nil {
		return "", err
	}

	if len(templates) == 0 {
		return "", errors.New("TODO: return recoverable error") // TODO
	}

	defTemplate := templates[0] // FIXME
	if len(defTemplate.Depends) == 0 {
		return Message(defTemplate.Raw), nil
	}

	for _, defType := range defTemplate.Depends {
		if _, ok := state.Get(defType); ok {
			continue
		}

		if _, err := pickDef(defType, state, repo); err != nil {
			return "", xerrors.Errorf("failed to pick depend definition: %w", err)
		}
	}
	return defTemplate.Execute(state)
}

func pickDef(defType DefinitionType, state State, repo *DefinitionRepository) (Message, error) {
	for _, candidateDef := range repo.List(defType) {
		if ok, _ := candidateDef.CanBePicked(state); ok {
			message, err := generate(candidateDef, state, repo)
			if err != nil {
				return "", err
			}
			state.Set(defType, message)
			if _, err := state.SetByConstraints(candidateDef.Constraints); err != nil {
				return "", xerrors.Errorf("failed to update state while message generating: %w", err)
			}
			return message, nil
		}
	}
	return "", fmt.Errorf("all depend definition are not satisfied constraints: %s", defType)
}
