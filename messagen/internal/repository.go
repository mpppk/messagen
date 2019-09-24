package internal

import (
	"fmt"

	"golang.org/x/xerrors"
)

type definitionMap map[DefinitionType][]*Definition
type Message string

func AscendingOrderTemplatePicker(templates *Templates, state State) (Templates, error) {
	return *templates, nil
}

type DefinitionRepository struct {
	m                 definitionMap
	templatePickers   []TemplatePicker
	definitionPickers []DefinitionPicker
}

type DefinitionRepositoryOption struct {
	TemplatePickers   []TemplatePicker
	DefinitionPickers []DefinitionPicker
}

func NewDefinitionRepository(opt *DefinitionRepositoryOption) *DefinitionRepository {
	templatePickers := opt.TemplatePickers
	if templatePickers == nil {
		templatePickers = []TemplatePicker{}
	}
	definitionPickers := opt.DefinitionPickers
	if definitionPickers == nil {
		definitionPickers = []DefinitionPicker{}
	}
	return &DefinitionRepository{
		m:                 definitionMap{},
		templatePickers:   templatePickers,
		definitionPickers: definitionPickers,
	}
}

func (d *DefinitionRepository) List(defType DefinitionType) (defs Definitions) {
	defs, ok := d.m[defType]
	if !ok {
		return Definitions{}
	}
	return defs
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
	defs, err := d.pickDefinitions(defType, initialState)
	if err != nil {
		return "", xerrors.Errorf("failed to generate message: %w", err)
	}

	for _, def := range defs {
		msg, err := generate(def, initialState, d)
		// TODO: handling recoverable error
		return msg, err
	}
	return "", xerrors.Errorf("failed to generate message. satisfied definitions are don't exist")
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

func (d *DefinitionRepository) pickDefinitions(defType DefinitionType, state State) (Definitions, error) {
	return d.applyDefinitionPickers(d.List(defType), state)
}

func (d *DefinitionRepository) applyDefinitionPickers(defs Definitions, state State) (Definitions, error) {
	newDefinitions, err := defs.Copy()
	if err != nil {
		return nil, xerrors.Errorf("failed to pick definitions: %w", err)
	}
	for _, definitionPicker := range d.definitionPickers {
		newDefinitions, err = definitionPicker(&newDefinitions, state)
	}
	return newDefinitions, nil
}

func generate(def *Definition, state State, repo *DefinitionRepository) (Message, error) {
	templates, err := repo.applyTemplatePickers(def.Templates, state)
	if err != nil {
		return "", err
	}

	if len(templates) == 0 {
		return "", NewNoPickableTemplateError("")
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
	candidateDefs, err := repo.pickDefinitions(defType, state)
	if err != nil {
		return "", xerrors.Errorf("failed to ")
	}
	for _, candidateDef := range candidateDefs {
		if ok, _ := candidateDef.CanBePicked(state); ok {
			message, err := generate(candidateDef, state, repo)
			if e, ok := err.(MessagenError); ok && e.Recoverable() {
				continue
			} else if err != nil {
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
