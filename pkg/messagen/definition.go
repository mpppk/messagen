package messagen

import (
	"fmt"

	"golang.org/x/xerrors"
)

type DefinitionType string
type DefinitionWeight float32
type Alias map[DefinitionType]DefinitionType

type RawDefinition struct {
	Type           DefinitionType
	RawTemplates   []RawTemplate
	Labels         Labels
	Requires       Labels
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

func (d *Definition) CanBePicked(labels Labels) (bool, string) {
	for requireDefType, requireLabel := range d.Requires {
		// 新しいdefがrequireしていて現在のdefにない場合はskip
		label, ok := labels[requireDefType]
		if !ok {
			return false, fmt.Sprintf("labels does not have required Type key: %s actual: %s", d.Type, requireLabel)
		}
		if label != requireLabel {
			return false, fmt.Sprintf("Target Definition Type(%s) does not have required label: %s actual: %s", d.Type, requireLabel, label)
		}
	}
	return true, ""
}

func (d *Definition) ListUnsatisfiedRequires(currentLabels Labels) (types []DefinitionType) {
	for defTypeStr, label := range d.Requires {
		if l, ok := currentLabels[defTypeStr]; ok && l == label {
			continue
		}
		types = append(types, DefinitionType(defTypeStr))
	}
	return
}
