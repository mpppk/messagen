package messagen

import (
	"math/rand"

	"golang.org/x/xerrors"
)

type DefinitionMap map[DefinitionType][]*Definition
type Message string
type Labels map[string]Message

func (g Labels) Set(id DefinitionType, message Message) {
	g[string(id)] = message
}

func (g Labels) Get(id DefinitionType) (Message, bool) {
	msg, ok := g[string(id)]
	return msg, ok
}

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

func (d *DefinitionRepository) Generate(defType DefinitionType) (string, error) {
	currentLabels := Labels{}
	def, ok := d.pickRandom(defType)
	if !ok {
		return "", xerrors.Errorf("failed to generate message. Root Definition type not found: %s", defType)
	}

	msg, err := generate(def, currentLabels, d)
	return string(msg), err
}

func generate(def *Definition, currentLabels Labels, repo *DefinitionRepository) (Message, error) {
	defTemplate := def.Templates.GetRandom()
	if len(defTemplate.Depends) == 0 {
		return Message(defTemplate.Raw), nil
	}

	for _, defType := range defTemplate.Depends {
		if _, ok := currentLabels.Get(defType); ok {
			continue
		}

		for _, candidateDef := range repo.List(defType) {
			if ok, _ := def.CanBePicked(currentLabels); ok {
				message, err := generate(candidateDef, currentLabels, repo)
				if err != nil {
					return "", err
				}
				currentLabels.Set(defType, message)
			}
		}
	}
	return defTemplate.Execute(currentLabels)
}
