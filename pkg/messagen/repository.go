package messagen

import (
	"fmt"
	"math/rand"

	"golang.org/x/xerrors"
)

type DefinitionMap map[DefinitionType][]*Definition
type Message string

type DefinitionRepository struct {
	m DefinitionMap
}

type DefinitionRepositoryOption struct {
}

func NewDefinitionRepository() *DefinitionRepository {
	return &DefinitionRepository{
		m: DefinitionMap{},
	}
}

func (d *DefinitionRepository) List(defType DefinitionType) (defs []*Definition) {
	defs, ok := d.m[defType]
	if !ok {
		return []*Definition{}
	}
	return defs
}

func (d *DefinitionRepository) pickRandom(defType DefinitionType) (*Definition, bool) {
	defs, ok := d.m[defType]
	if !ok {
		return nil, false
	}
	return defs[rand.Intn(len(defs))], true
}

func (d *DefinitionRepository) Add(def *Definition) {
	if defs, ok := d.m[def.Type]; ok {
		d.m[def.Type] = append(defs, def)
		return
	}
	d.m[def.Type] = []*Definition{def}
	return
}

func (d *DefinitionRepository) Generate(defType DefinitionType, initialState State) (string, error) {
	if initialState == nil {
		initialState = State{}
	}
	def, ok := d.pickRandom(defType)
	if !ok {
		return "", xerrors.Errorf("failed to generate message. Root Definition type not found: %s", defType)
	}

	msg, err := generate(def, initialState, d)
	return string(msg), err
}

func generate(def *Definition, state State, repo *DefinitionRepository) (Message, error) {
	defTemplate := def.Templates.GetRandom()
	if len(defTemplate.Depends) == 0 {
		return Message(defTemplate.Raw), nil
	}

	for _, defType := range defTemplate.Depends {
		if _, ok := state.Get(defType); ok {
			continue
		}

		if _, err := pickDef(defType, state, repo); err != nil {
			return "", xerrors.Errorf("failed to pick depend definition: %w", err)
		}
	}
	return defTemplate.Execute(state)
}

func pickDef(defType DefinitionType, state State, repo *DefinitionRepository) (Message, error) {
	for _, candidateDef := range repo.List(defType) {
		if ok, _ := candidateDef.CanBePicked(state); ok {
			message, err := generate(candidateDef, state, repo)
			if err != nil {
				return "", err
			}
			state.Set(defType, message)
			if _, err := state.SetByConstraints(candidateDef.Constraints); err != nil {
				return "", xerrors.Errorf("failed to update state while message generating: %w", err)
			}
			return message, nil
		}
	}
	return "", fmt.Errorf("all depend definition are not satisfied constraints: %s", defType)
}
