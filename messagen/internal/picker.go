package internal

import (
	"math/rand"

	"golang.org/x/xerrors"
)

type TemplatePicker func(def *Definition, aliasName AliasName, alias *Alias, state *State) (Templates, error)
type DefinitionPicker func(defs *Definitions, state *State) ([]*Definition, error)

func RandomTemplatePicker(def *Definition, aliasName AliasName, alias *Alias, state *State) (Templates, error) {
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

func NotAllowAliasDuplicateTemplatePicker(def *Definition, aliasName AliasName, alias *Alias, state *State) (Templates, error) {
	if aliasName != "" && alias.AllowDuplicate {
		return def.Templates, nil
	}
	aliasTemplates, ok := state.pickedTemplates[def.ID]
	if !ok {
		return def.Templates, nil
	}

	var templates Templates
	if templates1, ok := aliasTemplates[aliasName]; ok {
		templates = *templates1
	}
	if templates2, ok2 := aliasTemplates[""]; ok2 {
		templates = append(templates, *templates2...)
	}
	if !ok {
		return def.Templates, nil
	}

	return def.Templates.Subtract(templates...), nil
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
			return i - 1
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
