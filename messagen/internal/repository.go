package internal

import (
	"errors"

	"golang.org/x/sync/errgroup"

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

	msgChan := make(chan Message)
	errChan := make(chan error)
	for _, def := range defs {
		go func(def *Definition) {
			defMsgChan, defErrChan, err := generate(def, initialState, d)
			if err != nil {
				errChan <- err
				return
			}
			select {
			case msg := <-defMsgChan:
				msgChan <- msg
			case err := <-defErrChan:
				defErrChan <- err
			}
		}(def)
	}
	select {
	case msg := <-msgChan:
		return msg, nil
	case err := <-errChan:
		return "", err
	}
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

func generate(def *Definition, state State, repo *DefinitionRepository) (chan Message, chan error, error) {
	messageChan := make(chan Message)
	errChan := make(chan error)
	templates, err := repo.applyTemplatePickers(def.Templates, state)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		templateMessageChan, templateErrChan := resolveTemplates(templates, state, repo)
		select {
		case msg, ok := <-templateMessageChan:
			if ok {
				messageChan <- msg
			} else {
				return
			}
		case err, ok := <-templateErrChan:
			if ok {
				errChan <- err
			}
		}
	}()

	return messageChan, errChan, nil
}

func resolveTemplates(templates Templates, state State, repo *DefinitionRepository) (chan Message, chan error) {
	eg := errgroup.Group{}
	messageChan := make(chan Message)
	for _, defTemplate := range templates {
		defTemplate := defTemplate
		eg.Go(func() error {
			if len(defTemplate.Depends) == 0 {
				messageChan <- Message(defTemplate.Raw)
				return nil
			}
			newState := state.Copy()
			stateChan, err := resolveDefDepends(defTemplate, newState, repo)
			if err != nil {
				return err
			}
			for satisfiedState := range stateChan {
				msg, err := defTemplate.Execute(satisfiedState)
				if err != nil {
					return err
				}
				messageChan <- msg
			}
			return nil
		})
	}
	errChan := make(chan error)
	go func() {
		if err := eg.Wait(); err != nil {
			errChan <- err
		}
	}()
	return messageChan, errChan
}

func resolveDefDepends(template *Template, state State, repo *DefinitionRepository) (chan State, error) {
	stateChan := make(chan State)
	if template.IsSatisfiedState(state) {
		go func() {
			stateChan <- state
		}()
		return stateChan, nil
	}

	defType, _ := template.GetFirstUnsatisfiedDef(state)
	pickDefStateChan, err := pickDef(defType, state, repo)
	if err != nil {
		return nil, err
	}

	errChan := make(chan error)
	go func() {
		for newState := range pickDefStateChan {
			satisfiedStateChan, err := resolveDefDepends(template, newState, repo)
			if err != nil {
				errChan <- err
				return
			}
			go func() {
				for satisfiedState := range satisfiedStateChan {
					stateChan <- satisfiedState
				}
			}()
		}
	}()

	go func() {
		panic(<-errChan) // FIXME
	}()

	return stateChan, nil
}

func pickDef(defType DefinitionType, state State, repo *DefinitionRepository) (chan State, error) {
	candidateDefs, err := repo.pickDefinitions(defType, state)
	if err != nil {
		return nil, xerrors.Errorf("failed to pick definitions")
	}

	stateChan := make(chan State)
	eg := errgroup.Group{}
	for _, candidateDef := range candidateDefs {
		if ok, _ := candidateDef.CanBePicked(state); !ok {
			continue
		}
		candidateDef := candidateDef
		eg.Go(func() error {
			defMessageChan, defErrChan, err := generate(candidateDef, state, repo)
			if err != nil {
				return err
			}

			if defMessageChan == nil {
				return errors.New("message chan is nil")
			}

			if defErrChan == nil {
				return errors.New("err chan is nil")
			}

			select {
			case defMessage, ok := <-defMessageChan:
				if !ok {
					return nil
				}
				newState := state.Copy()
				newState.Set(defType, defMessage)
				if _, err := newState.SetByConstraints(candidateDef.Constraints); err != nil {
					return xerrors.Errorf("failed to update state while message generating: %w", err)
				}
				stateChan <- newState
			case err := <-defErrChan:
				return err
			}
			return nil
		})
	}

	go func() {
		if err := eg.Wait(); err != nil {
			panic(err) // FIXME
		}
	}()
	return stateChan, nil
}
