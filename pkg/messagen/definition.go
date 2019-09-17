package messagen

type DefinitionID string
type DefinitionWeight float32

type RawDefinition struct {
	ID             DefinitionID
	RawTemplates   []RawTemplate
	Requires       []DefinitionID
	Excludes       []DefinitionID
	Alias          map[DefinitionID]DefinitionID
	AllowDuplicate bool
	Weight         DefinitionWeight
}

type Definition struct {
	RawDefinition
	Templates Templates
}
