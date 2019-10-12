package internal

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"

	"golang.org/x/xerrors"
)

type definitionMap map[DefinitionType][]*Definition
type Message string

func AscendingOrderTemplatePicker(def *DefinitionWithAlias, state *State) (Templates, error) {
	return def.Templates, nil
}

type DefinitionRepository struct {
	m                 definitionMap
	templatePickers   []TemplatePicker
	definitionPickers []DefinitionPicker
	maxID             DefinitionID
}

type DefinitionRepositoryOption struct {
	TemplatePickers   []TemplatePicker
	DefinitionPickers []DefinitionPicker
}

func NewDefinitionRepository(opt *DefinitionRepositoryOption) *DefinitionRepository {
	templatePickers := []TemplatePicker{NotAllowAliasDuplicateTemplatePicker}
	if opt != nil && opt.TemplatePickers != nil {
		templatePickers = append(templatePickers, opt.TemplatePickers...)
	}

	definitionPickers := []DefinitionPicker{}
	if opt != nil && opt.DefinitionPickers != nil {
		definitionPickers = append(definitionPickers, opt.DefinitionPickers...)
	}
	return &DefinitionRepository{
		m:                 definitionMap{},
		templatePickers:   templatePickers,
		definitionPickers: definitionPickers,
		maxID:             0,
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

	wg := sync.WaitGroup{}
	for _, def := range defs {
		wg.Add(1)
		go func(def *Definition) {
			defWithAlias := &DefinitionWithAlias{
				Definition: def,
				aliasName:  "",
				alias:      nil,
			}
			stateChan, defErrChan, err := generate(defWithAlias, initialState, d)
			if err != nil {
				errChan <- err
				return
			}
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
						wg.Done()
						return
					}
				case err, ok := <-defErrChan:
					if ok {
						errChan <- err
					}
					wg.Done()
					return
				}
			}
		}(def)
	}

	go func() {
		wg.Wait()
		close(msgChan)
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

func generate(def *DefinitionWithAlias, state *State, repo *DefinitionRepository) (chan *State, chan error, error) {
	stateChan := make(chan *State)
	errChan := make(chan error)
	templates, err := repo.applyTemplatePickers(def, state)
	if err != nil {
		return nil, nil, err
	}

	newDef := *def
	newDef.Templates = templates

	go func() {
		subStateChan, templateErrChan := resolveTemplates(&newDef, state, repo)
		if err := pipeStateChan(subStateChan, stateChan, templateErrChan); err != nil {
			errChan <- err
		} else {
			close(stateChan)
		}
	}()
	return stateChan, errChan, nil
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
		case err := <-errChan:
			return err
		}
	}
}

func resolveTemplates(def *DefinitionWithAlias, state *State, repo *DefinitionRepository) (chan *State, chan error) {
	stateChan := make(chan *State)
	eg := errgroup.Group{}
	templates := def.Templates
	for _, defTemplate := range templates {
		defTemplate := defTemplate
		eg.Go(func() error {
			newState := state.Copy(def.OrderBy)
			if len(*defTemplate.Depends) == 0 {
				if err := newState.Update(def, defTemplate, Message(defTemplate.Raw)); err != nil {
					return err
				}
				stateChan <- newState
				return nil
			}
			subStateChan, err := resolveDefDepends(defTemplate, newState, repo, def.Aliases)
			if err != nil {
				return err
			}
			for satisfiedState := range subStateChan {
				msg, err := defTemplate.Execute(satisfiedState)
				if err != nil {
					return err
				}
				newSatisfiedState := satisfiedState.Copy(def.OrderBy)
				if err := newSatisfiedState.Update(def, defTemplate, msg); err != nil {
					return err
				}
				stateChan <- newSatisfiedState
			}
			return nil
		})
	}
	errChan := make(chan error)
	go func() {
		if err := eg.Wait(); err != nil {
			errChan <- err
		}

		close(stateChan)
	}()
	return stateChan, errChan
}

func resolveDefDepends(template *Template, state *State, repo *DefinitionRepository, aliases Aliases) (chan *State, error) {
	stateChan := make(chan *State)
	if template.IsSatisfiedState(state) {
		go func() {
			stateChan <- state
			close(stateChan)
		}()
		return stateChan, nil
	}

	defType, _ := template.GetFirstUnsatisfiedDef(state)
	alias, ok := aliases[AliasName(defType)]
	var aliasName AliasName
	if ok {
		aliasName = AliasName(defType)
		defType = alias.ReferType
	}
	pickDefStateChan, err := pickDef(defType, aliasName, alias, state, repo)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error)
	go func() {
		for newState := range pickDefStateChan {
			satisfiedStateChan, err := resolveDefDepends(template, newState, repo, aliases)
			if err != nil {
				errChan <- err
				return
			}
			wg.Add(1)
			go func() {
				for satisfiedState := range satisfiedStateChan {
					stateChan <- satisfiedState
				}
				wg.Done()
			}()
		}
		wg.Done()
	}()

	go func() {
		panic(<-errChan) // FIXME goroutine leak
	}()

	go func() {
		wg.Wait()
		close(stateChan)
	}()

	return stateChan, nil
}

func pickDef(defType DefinitionType, aliasName AliasName, alias *Alias, state *State, repo *DefinitionRepository) (chan *State, error) {
	candidateDefs, err := repo.pickDefinitions(defType, state)
	if err != nil {
		return nil, xerrors.Errorf("failed to pick definitions")
	}

	stateChan := make(chan *State)
	eg := errgroup.Group{}
	for _, candidateDef := range candidateDefs {
		if ok, _ := candidateDef.CanBePicked(state); !ok {
			continue
		}

		candidateDef := candidateDef
		candidateDefWithAlias := &DefinitionWithAlias{
			Definition: candidateDef,
			aliasName:  aliasName,
			alias:      alias,
		}
		eg.Go(func() error {
			subStateChan, defErrChan, err := generate(candidateDefWithAlias, state, repo)
			if err != nil {
				return err
			}

			if subStateChan == nil {
				return errors.New("message chan is nil")
			}

			if defErrChan == nil {
				return errors.New("err chan is nil")
			}

			if err := pipeStateChan(subStateChan, stateChan, defErrChan); err != nil {
				return err
			}
			return nil
		})
	}

	go func() {
		if err := eg.Wait(); err != nil {
			panic(err) // FIXME
		}
		close(stateChan)
	}()
	return stateChan, nil
}
