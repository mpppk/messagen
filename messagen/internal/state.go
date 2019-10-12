package internal

import (
	"golang.org/x/xerrors"
)

type MessageMap map[string]Message

func (m MessageMap) copy() MessageMap {
	newM := MessageMap{}
	for key, value := range m {
		newM[key] = value
	}
	return newM
}

type PickedTemplateMap map[DefinitionID]*Templates

func (p PickedTemplateMap) copy(orderBy []DefinitionType) (PickedTemplateMap, error) {
	newP := PickedTemplateMap{}
	for id, templates := range p {
		newTemplates, err := templates.Copy(orderBy)
		if err != nil {
			return nil, err
		}
		newP[id] = &newTemplates
	}
	return newP, nil
}

type AliasName string

type AliasMap map[DefinitionID][]AliasName

func (a AliasMap) copy() AliasMap {
	newA := AliasMap{}
	for defType, aliasNames := range a {
		copiedAliasNames := make([]AliasName, len(aliasNames))
		copy(copiedAliasNames, aliasNames)
		newA[defType] = copiedAliasNames
	}
	return newA
}

type State struct {
	m               MessageMap
	pickedTemplates PickedTemplateMap
	aliases         AliasMap
}

func NewState(m MessageMap) *State {
	if m == nil {
		m = MessageMap{}
	}
	return &State{
		m:               m,
		pickedTemplates: PickedTemplateMap{},
		aliases:         AliasMap{},
	}
}

func (s *State) Set(defType DefinitionType, msg Message) {
	s.m[string(defType)] = msg
}

func (s *State) SetAlias(defID DefinitionID, aliasName AliasName, msg Message) {
	s.m[string(aliasName)] = msg
	aliasNames, ok := s.aliases[defID]
	if ok {
		s.aliases[defID] = append(aliasNames, aliasName)
	} else {
		s.aliases[defID] = []AliasName{aliasName}
	}
}

func (s *State) SetByConstraint(constraint *Constraint) (bool, error) {
	if !constraint.key.WillAddValue {
		return false, nil
	}

	msg, err := constraint.value.ToMessage()
	if err != nil {
		return false, xerrors.Errorf("failed to set constraint value to state: %w", err)
	}

	if _, ok := s.Get(constraint.key.DefinitionType); ok {
		return false, nil
	}

	s.Set(constraint.key.DefinitionType, msg)
	return true, nil
}

func (s *State) SetByConstraints(constraints *Constraints) (int, error) {
	cnt := 0
	for rKey, rValue := range constraints.raw {
		constraint, err := NewConstraint(rKey, rValue)
		if err != nil {
			return -1, xerrors.Errorf("failed to set state from constraints: %w", err)
		}

		ok, err := s.SetByConstraint(constraint)
		if err != nil {
			return -1, xerrors.Errorf("failed to set state from constraints: %w", err)
		}
		if ok {
			cnt++
		}
	}
	return cnt, nil
}

func (s *State) SetByDef(def *DefinitionWithAlias, msg Message) error {
	if def.aliasName == "" {
		s.Set(def.Type, msg)
	} else {
		s.SetAlias(def.ID, def.aliasName, msg)
	}
	if _, err := s.SetByConstraints(def.Constraints); err != nil {
		return xerrors.Errorf("failed to update state while message generating: %w", err)
	}
	return nil
}

func (s *State) Update(def *DefinitionWithAlias, pickedTemplate *Template, msg Message) error {
	if err := s.SetByDef(def, msg); err != nil {
		return err
	}
	s.AddPickedTemplate(def.ID, pickedTemplate)
	return nil
}

func (s *State) AddPickedTemplate(defID DefinitionID, template *Template) {
	templates, ok := s.pickedTemplates[defID]
	if ok {
		*templates = append(*templates, template)
	} else {
		s.pickedTemplates[defID] = &Templates{template}
	}
}

func (s *State) Get(defType DefinitionType) (Message, bool) {
	v, ok := s.m[string(defType)]
	return v, ok
}

func (s *State) IsPickedTemplate(defID DefinitionID, template *Template) bool {
	templates, ok := s.pickedTemplates[defID]
	if ok && templates.Has(template) {
		return true
	}
	return false
}

func (s *State) Copy(orderBy []DefinitionType) *State {
	ns := NewState(s.m.copy())
	pickedTemplates, err := s.pickedTemplates.copy(orderBy)
	if err != nil {
		panic(err) // FIXME
	}
	ns.pickedTemplates = pickedTemplates
	ns.aliases = s.aliases.copy()

	return ns
}
