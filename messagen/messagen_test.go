package messagen_test

import (
	"fmt"
	"math/rand"

	"github.com/mpppk/messagen/messagen"
)

func Example() {
	generator, _ := messagen.New(nil)

	definitions := []*messagen.Definition{
		{
			// Each Definition have one Type, and multiple definitions can have the same Type.
			// Definitions are referred from variable in templates via Type.
			// If referred Type is shared by multiple definitions, they are chosen randomly.
			Type: "Root",

			// Templates are template for generate message.
			// If two ore more templates are given, one of them is picked at random.
			// You can write template like golang text/template format. (but currently functions are unavailable)
			// If Type is embedded by the notation like {{.SomeType}},
			// one of definition that have the Type is chosen and inject generated message.
			// For example, below template refers three Types, Pronoun, FirstName, and LastName.
			Templates: []string{"{{.Pronoun}} is {{.FirstName}} {{.LastName}}."},
		},
		{
			Type:      "Pronoun",
			Templates: []string{"She"},

			// Definition can have Constraints, which key value map for control whether the definition is picked.
			// Constraints key is consisted by Type and Operators.
			// Below Constraints key is "Gender+", which consisted by Type(Gender) and Operator(+).
			// Operator add more constraints related to the key.
			// For example, below "+" operator means...
			// * Gender key is optional. So even if Gender key does not exist, this definitions can be picked.
			//   * If so, Gender key whose value is Female is added.
			// * If Gender key does exist and value is not Female, the definition can not be picked.
			Constraints: map[string]string{"Gender+": "Female"},
		},
		{
			Type:        "Pronoun",
			Templates:   []string{"He"},
			Constraints: map[string]string{"Gender+": "Male"},
		},
		{
			Type:        "FirstName",
			Templates:   []string{"Liam", "James", "Benjamin"},
			Constraints: map[string]string{"Gender+": "Male"},
		},
		{
			Type:        "FirstName",
			Templates:   []string{"Emily", "Charlotte", "Sofia"},
			Constraints: map[string]string{"Gender+": "Female"},
		},
		{
			Type:      "LastName",
			Templates: []string{"Smith", "Williams", "Brown"},
		},
	}

	// AddDefinition definitions to generator.
	_ = generator.AddDefinition(definitions...)

	// Set random seed for pick definitions and templates.
	rand.Seed(0)

	// Generate method generate message according to added definitions.
	// First argument represent definition Type of start point.
	message, _ := generator.Generate("Root", nil)

	// Second argument represent initial state.
	// In below code, Gender key is added with Female value as initial state.
	// Therefore, Pronoun and FirstName definitions that have constraints which include Gender:Female are always picked.
	femaleMessage, _ := generator.Generate("Root", map[string]string{"Gender": "Female"})

	maleMessage, _ := generator.Generate("Root", map[string]string{"Gender": "Male"})
	fmt.Printf("%s\n%s\n%s\n", message, femaleMessage, maleMessage)

	// Output:
	// She is Charlotte Williams.
	// She is Emily Smith.
	// He is James Smith.
}
