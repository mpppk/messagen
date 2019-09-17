package messagen

import (
	"fmt"

	"golang.org/x/xerrors"
)

type DefinitionMap map[DefinitionID]*Definition
type GeneratedMessage string
type GeneratedMessageMap map[string]GeneratedMessage

func (g GeneratedMessageMap) Set(id DefinitionID, message GeneratedMessage) {
	g[string(id)] = message
}

func (g GeneratedMessageMap) Get(id DefinitionID) (GeneratedMessage, bool) {
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

func (d *DefinitionRepository) Get(defID DefinitionID) (*Definition, bool) {
	def, ok := d.m[defID]
	return def, ok
}

func (d *DefinitionRepository) Add(def *Definition) error {
	if _, ok := d.m[def.ID]; ok {
		return xerrors.Errorf("definition already exist. ID:%s", def.ID)
	}
	d.m[def.ID] = def
	return nil
}

func (d *DefinitionRepository) Generate(id DefinitionID) (string, error) {
	messageMap := GeneratedMessageMap{}
	def, ok := d.Get(id)
	if !ok {
		return "", fmt.Errorf("failed to get Definition. ID: %s", id)
	}
	msg, err := generate(def, messageMap, d)
	return string(msg), err
}

func generate(def *Definition, m GeneratedMessageMap, repo *DefinitionRepository) (GeneratedMessage, error) {
	defTemplate := def.Templates.GetRandom()
	if len(defTemplate.Depends) == 0 {
		return GeneratedMessage(defTemplate.Raw), nil
	}

	for _, defID := range defTemplate.Depends {
		if _, ok := m.Get(defID); ok {
			continue
		}

		newDef, ok := repo.Get(defID)
		if !ok {
			return "", xerrors.Errorf("failed to get Definition. ID:%s", defID)
		}

		s, err := generate(newDef, m, repo)
		if err != nil {
			return "", err
		}

		m.Set(defID, s)
	}
	return defTemplate.Execute(m)
}
