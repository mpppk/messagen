package messagen

import (
	"golang.org/x/xerrors"
)

type DefinitionType string
type DefinitionWeight float32
type Alias map[DefinitionType]DefinitionType

type RawDefinition struct {
	Type           DefinitionType
	RawTemplates   []RawTemplate
	Constraints    *Constraints
	Alias          Alias
	AllowDuplicate bool
	Weight         DefinitionWeight
}

type Definition struct {
	*RawDefinition
	Templates Templates
}

func NewDefinition(rawDefinition *RawDefinition) (*Definition, error) {
	templates, err := NewTemplates(rawDefinition.RawTemplates)
	if err != nil {
		return nil, xerrors.Errorf("failed to create Definition: %w", err)
	}

	return &Definition{
		RawDefinition: rawDefinition,
		Templates:     templates,
	}, nil
}

func (d *Definition) CanBePicked(state State) (bool, error) {
	if ok, err := d.Constraints.AreSatisfied(state); err != nil {
		return false, xerrors.Errorf("failed to check definition can be picked: %w", err)
	} else {
		return ok, nil
	}
}
