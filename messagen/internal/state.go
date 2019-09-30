package internal

import (
	"golang.org/x/xerrors"
)

type MessageMap map[string]Message

type State struct {
	m               MessageMap
	pickedTemplates map[DefinitionID]*Templates
	aliases         map[DefinitionType][]DefinitionType
}

func NewState(m MessageMap) *State {
	if m == nil {
		m = MessageMap{}
	}
	return &State{
		m:               m,
		pickedTemplates: map[DefinitionID]*Templates{},
		aliases:         map[DefinitionType][]DefinitionType{},
	}
}

func (s *State) Set(defType DefinitionType, msg Message) {
	s.m[string(defType)] = msg
}

func (s *State) SetAlias(defType, aliasName DefinitionType, msg Message) {
	s.m[string(aliasName)] = msg
	aliasNames, ok := s.aliases[defType]
	if ok {
		s.aliases[defType] = append(aliasNames, aliasName)
	} else {
		s.aliases[defType] = []DefinitionType{aliasName}
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

func (s *State) Copy() *State {
	ns := NewState(nil)
	for key, value := range s.m {
		ns.m[key] = value
	}
	return ns
}
