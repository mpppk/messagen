package internal

import (
	"math/rand"
	"sort"

	"golang.org/x/xerrors"
)

type TemplatePicker func(def *DefinitionWithAlias, state *State) (Templates, error)
type DefinitionPicker func(defs *Definitions, state *State) ([]*Definition, error)

func RandomTemplatePicker(def *DefinitionWithAlias, state *State) (Templates, error) {
	templates := def.Templates
	var newTemplates Templates
	for {
		if len(templates) == 0 {
			break
		}
		tmpl, ok := templates.PopRandom()
		if !ok {
			return nil, xerrors.Errorf("failed to pop template random from %v", templates)
		}
		newTemplates = append(newTemplates, tmpl)
	}
	return newTemplates, nil
}

func NotAllowAliasDuplicateTemplatePicker(def *DefinitionWithAlias, state *State) (Templates, error) {
	if def.alias != nil && def.alias.AllowDuplicate {
		return def.Templates, nil
	}
	pickedTemplates, ok := state.pickedTemplates[def.ID]
	if !ok {
		return def.Templates, nil
	}
	return def.Templates.Subtract(*pickedTemplates...), nil
}

func RandomWithWeightDefinitionPicker(definitions *Definitions, state *State) ([]*Definition, error) {
	var newDefinitions Definitions
	for {
		if len(*definitions) == 0 {
			break
		}
		var weights []DefinitionWeight
		for _, def := range *definitions {
			weights = append(weights, def.Weight)
		}
		def := definitions.PopByIndex(pickDefinitionIndexRandomWithWeight(weights))
		newDefinitions = append(newDefinitions, def)
	}
	return newDefinitions, nil
}

func ConstraintsSatisfiedDefinitionPicker(definitions *Definitions, state *State) ([]*Definition, error) {
	var newDefinitions Definitions
	for _, def := range *definitions {
		if ok, err := def.CanBePicked(state); err != nil {
			return nil, err
		} else if ok {
			newDefinitions = append(newDefinitions, def)
		}
	}
	return newDefinitions, nil
}

func SortByConstraintPriorityDefinitionPicker(definitions *Definitions, _ *State) ([]*Definition, error) {
	sort.SliceStable(*definitions, func(i, j int) bool {
		return (*definitions)[i].Constraints.Priority > (*definitions)[j].Constraints.Priority
	})
	return *definitions, nil
}

func pickDefinitionIndexRandomWithWeight(weights []DefinitionWeight) int {
	if len(weights) == 1 {
		return 0
	}

	weightSum := calcWeightSum(weights)
	r := randomFloat32(0, float64(weightSum))
	currentWeightSum := float32(0)
	for i, weight := range weights { // O(N)
		currentWeightSum += float32(weight)
		if r < currentWeightSum {
			return i
		}
	}
	panic("unexpected error occurred in currentWeightSum")
}

func calcWeightSum(weights []DefinitionWeight) (sum float32) {
	for _, weight := range weights {
		sum += float32(weight)
	}
	return
}

func randomFloat32(min, max float64) float32 {
	return float32(rand.Float64()*(max-min) + min)
}
