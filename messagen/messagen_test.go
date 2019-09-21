package messagen_test

import (
	"fmt"
	"math/rand"

	"github.com/mpppk/messagen/messagen"
)

func Example() {
	generator, _ := messagen.New()

	rawDefinitions := []*messagen.Definition{
		{
			Type:      "Root",
			Templates: []string{"{{.Pronoun}} is {{.FirstName}} {{.LastName}}."},
		},
		{
			Type:        "Pronoun",
			Templates:   []string{"She"},
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

	_ = generator.Add(rawDefinitions...)

	rand.Seed(42)
	message, _ := generator.Generate("Root", map[string]string{})
	femaleMessage, _ := generator.Generate("Root", map[string]string{"Gender": "Female"})
	maleMessage, _ := generator.Generate("Root", map[string]string{"Gender": "Male"})
	fmt.Printf("%s\n%s\n%s\n", message, femaleMessage, maleMessage)

	// Output:
	// She is Emily Williams.
	// She is Sofia Williams.
	// He is James Brown.
}
