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

type PickedTemplateMap map[DefinitionID]map[AliasName]*Templates

func (p PickedTemplateMap) copy(orderBy []DefinitionType) (PickedTemplateMap, error) {
	newP := PickedTemplateMap{}
	for id, aliasTemplates := range p {
		newP[id] = map[AliasName]*Templates{}
		for aliasName, templates := range aliasTemplates {
			newTemplates, err := templates.Copy(orderBy)
			if err != nil {
				return nil, err
			}
			newP[id][aliasName] = &newTemplates
		}
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

func (s *State) SetByDef(def *Definition, aliasName AliasName, msg Message) error {
	if aliasName == "" {
		s.Set(def.Type, msg)
	} else {
		s.SetAlias(def.ID, aliasName, msg)
	}
	if _, err := s.SetByConstraints(def.Constraints); err != nil {
		return xerrors.Errorf("failed to update state while message generating: %w", err)
	}
	return nil
}

func (s *State) Update(def *Definition, pickedTemplate *Template, aliasName AliasName, msg Message) error {
	if err := s.SetByDef(def, aliasName, msg); err != nil {
		return err
	}
	s.AddPickedTemplate(def.ID, aliasName, pickedTemplate)
	return nil
}

func (s *State) AddPickedTemplate(defID DefinitionID, aliasName AliasName, template *Template) {
	aliasTemplates, ok := s.pickedTemplates[defID]
	if !ok {
		aliasTemplates = map[AliasName]*Templates{}
	}
	_, ok = aliasTemplates[aliasName]
	if ok {
		t := append(*aliasTemplates[aliasName], template)
		aliasTemplates[aliasName] = &t
	} else {
		aliasTemplates[aliasName] = &Templates{template}
	}
	s.pickedTemplates[defID] = aliasTemplates
}

func (s *State) Get(defType DefinitionType) (Message, bool) {
	v, ok := s.m[string(defType)]
	return v, ok
}

func (s *State) IsPickedTemplate(defID DefinitionID, aliasName AliasName, template *Template) bool {
	aliasTemplates, ok := s.pickedTemplates[defID]
	if !ok {
		return false
	}

	templates, ok := aliasTemplates[aliasName]
	if ok && templates.Has(template) {
		return true
	}

	templates, ok = aliasTemplates[""]
	return ok && templates.Has(template)
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
