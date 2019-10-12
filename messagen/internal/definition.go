package internal

import (
	"golang.org/x/xerrors"
)

type DefinitionType string
type DefinitionID int
type DefinitionWeight float32
type Alias struct {
	ReferType      DefinitionType
	AllowDuplicate bool
}

type Aliases map[AliasName]*Alias

func (a Aliases) IsAlias(defType DefinitionType) bool {
	_, ok := a[AliasName(defType)]
	return ok
}

type RawDefinition struct {
	Type           DefinitionType
	RawTemplates   []RawTemplate
	Constraints    *Constraints
	Aliases        Aliases
	AllowDuplicate bool
	OrderBy        []DefinitionType
	Weight         DefinitionWeight
}

type Definition struct {
	*RawDefinition
	ID        DefinitionID
	Templates Templates
}

func NewDefinition(rawDefinition *RawDefinition) (*Definition, error) {
	templates, err := NewTemplates(rawDefinition.RawTemplates, rawDefinition.OrderBy)
	if err != nil {
		return nil, xerrors.Errorf("failed to create Definition: %w", err)
	}

	if rawDefinition.Constraints == nil {
		constraints, err := NewConstraints(nil)
		if err != nil {
			return nil, xerrors.Errorf("failed to set empty constraints to definition: %w", err)
		}
		rawDefinition.Constraints = constraints
	}

	if rawDefinition.Weight == 0 {
		rawDefinition.Weight = 1
	}
	return &Definition{
		RawDefinition: rawDefinition,
		Templates:     templates,
	}, nil
}

func (d *Definition) CanBePicked(state *State) (bool, error) {
	if ok, err := d.Constraints.AreSatisfied(state); err != nil {
		return false, xerrors.Errorf("failed to check definition can be picked: %w", err)
	} else {
		return ok, nil
	}
}

func (d *Definition) IsAlias(defType DefinitionType) bool {
	return d.Aliases.IsAlias(defType)
}

func (d *Definition) Copy() (*Definition, error) {
	def := *d
	templates, err := def.Templates.Copy(def.OrderBy)
	if err != nil {
		return nil, err
	}
	def.Templates = templates
	return &def, nil
}

type Definitions []*Definition

//func NewDefinitions(rawDefs ...*RawDefinition) (Definitions, error) {
//	var definitions Definitions
//	for _, rawDef := range rawDefs {
//		def, err := NewDefinition(rawDef)
//		if err != nil {
//			return nil, xerrors.Errorf("failed to create new definitions: %w", err)
//		}
//		definitions = append(definitions, def)
//	}
//	return definitions, nil
//}

func (d *Definitions) PopByIndex(index int) *Definition {
	def := (*d)[index]
	d.DeleteByIndex(index)
	return def
}

func (d *Definitions) DeleteByIndex(i int) {
	if i == 0 {
		*d = (*d)[1:]
		return
	}
	if len(*d)-1 == i {
		*d = (*d)[:len(*d)-1]
		return
	}
	*d = append((*d)[:i], (*d)[i+1:]...)
}

func (d *Definitions) Copy() (Definitions, error) {
	var newDefinitions []*Definition
	for _, definition := range *d {
		newDef, err := definition.Copy()
		if err != nil {
			return nil, err
		}
		newDefinitions = append(newDefinitions, newDef)
	}
	return newDefinitions, nil
}

type DefinitionWithAlias struct {
	*Definition
	aliasName AliasName
	alias     *Alias
}
