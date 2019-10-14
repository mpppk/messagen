package internal

import (
	"fmt"

	"golang.org/x/xerrors"
)

type definitionMap map[DefinitionType][]*Definition
type Message string

func AscendingOrderTemplatePicker(def *DefinitionWithAlias, state *State) (Templates, error) {
	return def.Templates, nil
}

type DefinitionRepository struct {
	m                  definitionMap
	templatePickers    []TemplatePicker
	definitionPickers  []DefinitionPicker
	templateValidators []TemplateValidator
	maxID              DefinitionID
}

type DefinitionRepositoryOption struct {
	TemplatePickers    []TemplatePicker
	DefinitionPickers  []DefinitionPicker
	TemplateValidators []TemplateValidator
}

func NewDefinitionRepository(opt *DefinitionRepositoryOption) *DefinitionRepository {
	templatePickers := []TemplatePicker{NotAllowAliasDuplicateTemplatePicker}
	if opt != nil && opt.TemplatePickers != nil {
		templatePickers = append(templatePickers, opt.TemplatePickers...)
	}

	definitionPickers := []DefinitionPicker{ConstraintsSatisfiedDefinitionPicker, RandomWithWeightDefinitionPicker}
	if opt != nil && opt.DefinitionPickers != nil {
		definitionPickers = append(definitionPickers, opt.DefinitionPickers...)
	}

	var templateValidators []TemplateValidator
	if opt != nil && opt.TemplateValidators != nil {
		templateValidators = opt.TemplateValidators
	}

	return &DefinitionRepository{
		m:                  definitionMap{},
		templatePickers:    templatePickers,
		definitionPickers:  definitionPickers,
		templateValidators: templateValidators,
		maxID:              0,
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
	def.ID = d.maxID
	d.maxID++
	if defs, ok := d.m[def.Type]; ok {
		d.m[def.Type] = append(defs, def)
		return
	}
	d.m[def.Type] = []*Definition{def}
	return
}

func (d *DefinitionRepository) Generate(defType DefinitionType, initialState *State, num uint) (messages []Message, err error) {
	msgChan, errChan := d.Start(defType, initialState)

	if num == 0 {
		return nil, fmt.Errorf("failed to generate messages. num must be greater than 1")
	}

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				if len(messages) == 0 {
					return nil, xerrors.Errorf("valid message does not exist")
				} else {
					return messages, nil
				}
			}
			messages = append(messages, msg)
			if len(messages) == int(num) {
				return messages, nil
			}
		case err := <-errChan:
			return nil, err
		}
	}
}

func (d *DefinitionRepository) Start(defType DefinitionType, initialState *State) (msgChan chan Message, errChan chan error) {
	stateChan := make(chan *State)
	msgChan = make(chan Message)
	errChan = make(chan error)
	if initialState == nil {
		initialState = NewState(nil)
	}
	defs, err := d.pickDefinitions(defType, initialState)
	if err != nil {
		errChan <- xerrors.Errorf("failed to generate message: %w", err)
		close(msgChan)
		return
	}

	go func() {
		for _, def := range defs {
			defWithAlias := &DefinitionWithAlias{
				Definition: def,
				aliasName:  "",
				alias:      nil,
			}
			subStateChan, templateErrChan := resolveTemplates(defWithAlias, initialState, d)
			if err := pipeStateChan(subStateChan, stateChan, templateErrChan); err != nil {
				errChan <- err
			}
		}
		close(stateChan)
	}()

	go func() {
		for {
			select {
			case state, ok := <-stateChan:
				if ok {
					msg, ok := state.Get(defType)
					if !ok {
						errChan <- fmt.Errorf("error occurred in Generate. message not found. def type: %s", defType)
					}
					msgChan <- msg
				} else {
					close(msgChan)
					return
				}
			case err, ok := <-errChan:
				if ok {
					errChan <- err
				}
				return
			}
		}
	}()
	return msgChan, errChan
}

func (d *DefinitionRepository) applyTemplatePickers(def *DefinitionWithAlias, state *State) (newTemplates Templates, err error) {
	newDef := *def
	newTemplates, err = def.Templates.Copy(newDef.OrderBy)
	if err != nil {
		return nil, err
	}
	newDef.Templates = newTemplates

	for _, picker := range d.templatePickers {
		if len(newTemplates) == 0 {
			return Templates{}, nil
		}
		newTemplates, err = picker(&newDef, state)
		newDef.Templates = newTemplates
		if err != nil {
			return nil, err
		}
	}
	return newTemplates, nil
}

func (d *DefinitionRepository) pickDefinitions(defType DefinitionType, state *State) (Definitions, error) {
	return d.applyDefinitionPickers(d.List(defType), state)
}

func (d *DefinitionRepository) applyDefinitionPickers(defs Definitions, state *State) (Definitions, error) {
	newDefinitions, err := defs.Copy()
	if err != nil {
		return nil, xerrors.Errorf("failed to pick definitions: %w", err)
	}
	for _, definitionPicker := range d.definitionPickers {
		newDefinitions, err = definitionPicker(&newDefinitions, state)
	}
	return newDefinitions, nil
}

func (d *DefinitionRepository) applyTemplateValidators(template *Template, state *State) (bool, error) {
	for _, templateValidator := range d.templateValidators {
		if ok, err := templateValidator(template, state); err != nil {
			return false, err
		} else if !ok {
			return false, nil
		}
	}
	return true, nil
}

func pipeStateChan(fromStateChan, toStateChan chan *State, errChan chan error) error {
	for {
		select {
		case newState, ok := <-fromStateChan:
			if ok {
				toStateChan <- newState
			} else {
				return nil
			}
		case err, ok := <-errChan:
			if !ok {
				return fmt.Errorf("err chan closed")
			}
			return err
		}
	}
}

func resolveTemplates(def *DefinitionWithAlias, state *State, repo *DefinitionRepository) (chan *State, chan error) {
	stateChan := make(chan *State)
	errChan := make(chan error)
	templates, err := repo.applyTemplatePickers(def, state)
	if err != nil {
		errChan <- err
		return stateChan, errChan
	}

	go func() {
		for _, defTemplate := range templates {
			defTemplate := defTemplate
			newState := state.Copy(def.OrderBy)
			if len(*defTemplate.Depends) == 0 {
				if err := newState.Update(def, defTemplate, Message(defTemplate.Raw)); err != nil {
					errChan <- err
					return
				}
				if ok, err := repo.applyTemplateValidators(defTemplate, newState); err != nil {
					errChan <- err
					return
				} else if ok {
					stateChan <- newState
				}
				continue
			}
			subStateChan, errChan2 := resolveDefDepends(defTemplate, newState, repo, def.Aliases)
		L:
			for {
				select {
				case satisfiedState, ok := <-subStateChan:
					if !ok {
						break L
					}
					msg, err := defTemplate.Execute(satisfiedState)
					if err != nil {
						errChan <- err
						return
					}

					newSatisfiedState := satisfiedState.Copy(def.OrderBy)
					if err := newSatisfiedState.Update(def, defTemplate, msg); err != nil {
						errChan <- err
						return
					}
					if ok, err := repo.applyTemplateValidators(defTemplate, newSatisfiedState); err != nil {
						errChan <- err
						return
					} else if ok {
						stateChan <- newSatisfiedState
					}
				case err := <-errChan2:
					errChan <- err
					return
				}
			}
		}
		close(stateChan)
	}()
	return stateChan, errChan
}

func resolveDefDepends(template *Template, state *State, repo *DefinitionRepository, aliases Aliases) (chan *State, chan error) {
	errChan := make(chan error)
	stateChan := make(chan *State)
	if template.IsSatisfiedState(state) {
		go func() {
			stateChan <- state
			close(stateChan)
		}()
		return stateChan, errChan
	}

	defType, _ := template.GetFirstUnsatisfiedDef(state)
	alias, ok := aliases[AliasName(defType)]
	var aliasName AliasName
	if ok {
		aliasName = AliasName(defType)
		defType = alias.ReferType
	}
	pickDefStateChan, _ := pickDef(defType, aliasName, alias, state, repo) // FIXME: handle error

	go func() {
		for newState := range pickDefStateChan {
			if ok, err := repo.applyTemplateValidators(template, newState); err != nil {
				errChan <- err
				return
			} else if !ok {
				continue
			}

			satisfiedStateChan, errChan2 := resolveDefDepends(template, newState, repo, aliases)
			if err := pipeStateChan(satisfiedStateChan, stateChan, errChan2); err != nil {
				errChan <- err
				return
			}
		}
		close(stateChan)
	}()

	return stateChan, errChan
}

func pickDef(defType DefinitionType, aliasName AliasName, alias *Alias, state *State, repo *DefinitionRepository) (chan *State, chan error) {
	stateChan := make(chan *State)
	errChan := make(chan error)
	candidateDefs, err := repo.pickDefinitions(defType, state)
	if err != nil {
		errChan <- xerrors.Errorf("failed to pick definitions", err)
		return stateChan, errChan
	}

	go func() {
		for _, candidateDef := range candidateDefs {
			candidateDef := candidateDef
			candidateDefWithAlias := &DefinitionWithAlias{
				Definition: candidateDef,
				aliasName:  aliasName,
				alias:      alias,
			}
			subStateChan, templateErrChan := resolveTemplates(candidateDefWithAlias, state, repo)
			if err := pipeStateChan(subStateChan, stateChan, templateErrChan); err != nil {
				errChan <- err
			}
		}
		close(stateChan)
	}()
	return stateChan, errChan
}
