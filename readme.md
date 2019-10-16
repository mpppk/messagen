# messagen
[![GoDoc](https://godoc.org/github.com/mpppk/messagen/messagen?status.svg)](https://godoc.org/github.com/mpppk/messagen)

messagen is the tree structured message generator with flexible constraints and declarative API.
You can use messagen as CLI tool or golang library. 

## Installation
### CLI
Download from GitHub Releases.

### golang library
```bash
$ go get github.com/mpppk/messagen
```

## Getting Started

Think about a message that introducing someone else's name like `He is Liam Smith.` or `She is Emily Williams.`, and you want to change pronoun and first/last name randomly.
What about the following template?

```
{{.Pronoun}} is {{.FirstName}} {{.LastName}}.
# Pronoun is picked from ['He, 'She'] randomly.
# FirstName is picked from ['Liam', 'James', 'Emily', 'Charlotte', ...] randomly.
# LastName is picked from ['Smith', 'Williams', 'Brown'] randomly.
```

This template may work well, but may generate inconsistent messages, because this message have one **constraints**.
If pronoun is `He`, first name must be masculine name.
The same is true if pronoun is `She`.

```
He is Liam Smith. # OK
She is Liam Smith. # NG because Liam is masculine name
```

*messagen* is the tool for generating messages that satisfy the **constraints** between words by declarative API.
*messagen* determines messages that can be generated by a set of **definitions** that group together templates, constraints and others.

Below is the **definitions** of *messagen* written in yaml.

```yaml
# intro.yaml
Definitions:
  - Type: Root
    Templates: ["{{.Pronoun}} is {{.FirstName}} {{.LastName}}."]
  - Type: Pronoun
    Templates: ["He", "She"]
  - Type: FirstName
    Templates: ["Liam", "James", "Benjamin"]
    Constraints: {"Pronoun": "He"}
  - Type: FirstName
    Templates: ["Emily", "Charlotte", "Sofia"]
    Constraints: {"Pronoun": "She"}
  - Type: LastName
    Templates: ["Smith", "Williams", "Brown"]
```

Then, execute *messagen* CLI.

```bash
$ messagen run -f intro.yaml
He is Liam Williams.
```

*messagen* always generate consistent message.

**Type** is an identifier used for **definition** grouping.
Multiple definitions can have same **Type**.
**Definition** can be referenced by describing it as `{{.SomeType}}` in the template. If multiple definitions are found, one of them is picked at random. 
(In golang library, this behavior can be controlled by `DefinitionPicker`. See [pickers section](https://github.com/mpppk/messagen#pickers).)
`Root` is a special type that is the starting point for message generation.

**Templates** is a set of templates used for message generation.
If a definition which have multiple templates is chosen, one of them is picked at random. (In golang library this behavior can be controlled by `TemplatePicker`. See [pickers section](https://github.com/mpppk/messagen#pickers))


**Constraints** is a key value object that determines whether a **definition** is pickable. 
Key is a definition type. Value is a required definition value.

If template includes other definition types, messagen choose one of definition, then pick one of templates, and these processes are repeated recursively.
In other words, the definitions can be regarded as having the tree structure.

In above example, there is only one Root Definition with one template (`"{{.Pronoun}} is {{.FirstName}} {{.LastName}}."`).
The template includes three definition types, `Pronoun`, `FirstName`, and `LastName`.
This can be represented as the following tree structure:

```
state: {}

Root:          ['{{.Pronoun}} is {{.FirstName}} {{.LastName}}.']
├── Pronoun:   ['He', 'She']
├── FirstName: ['Liam', 'Emily', ...]
└── LastName:  ['Smith', 'Williams', 'Brown']
```

By default, definition types are resolved in order from the beginning of template, so `Pronoun` is resolved first in this example.
If `She` is chosen as `Pronoun`, messagen state become as follows.

```
state: {Pronoun: She}

Root:          'She is {{.FirstName}} {{.LastName}}.'
├── Pronoun    ['He'] -- pick random --> 'She'
├── FirstName: ['Emily', 'Charlotte', 'Sofia'] (masculine name is dropped because unsatisfy constraints)
└── LastName:  ['Smith', 'Williams', 'Brown']
```

Next, `FirstName` is resolved. 

```
state: {Pronoun: She, FirstName: Emily}

Root:          'She is Emily {{.LastName}}.'
├── Pronoun    ['He'] 
├── FirstName: ['Charlotte', 'Sofia'] -- pick random --> 'Emily'
└── LastName:  ['Smith', 'Williams', 'Brown']
```

Last, `LastName` is resolved and messagen return the generated message.

```
state: {Pronoun: She, FirstName: Emily, LastName: Smith}

Root:          'She is Emily Smith.'
├── Pronoun    ['He'] 
├── FirstName: ['Charlotte', 'Sofia'] 
└── LastName:  ['Williams', 'Brown'] -- pick random --> 'Smith'
```

```bash
$ messagen run -f intro.yaml 
She is Emily Smith.
```

You can provide initial state by `--state` flag.

```bash
$ messagen run -f intro.yaml --state Pronoun=Male
He is Liam Williams.
```

## golang sample 

messagen can be used not only as a CLI tool but also as a golang library.
Below is a sample of the previous definitions written in golang.

```go
func main() {
	// CLI tool randomly pick a template by default, but in golang, you must specify it explicitly.
	opt := &messagen.Option{
		TemplatePickers: []messagen.TemplatePicker{messagen.RandomTemplatePicker},
    }
	generator, _ := messagen.New(opt)
)

	definitions := []*messagen.Definition{
		{
			Type: "Root",
			Templates: []string{"{{.Pronoun}} is {{.FirstName}} {{.LastName}}."},
		},
		{
			Type:      "Pronoun",
			Templates: []string{"He", "She"},
		},
		{
			Type:        "FirstName",
			Templates:   []string{"Liam", "James", "Benjamin"},
			Constraints: map[string]string{"Pronoun": "He"},
		},
		{
			Type:        "FirstName",
			Templates:   []string{"Emily", "Charlotte", "Sofia"},
			Constraints: map[string]string{"Pronoun": "She"},
		},
		{
			Type:      "LastName",
			Templates: []string{"Smith", "Williams", "Brown"},
		},
	}

	// AddDefinition definitions to generator.
	generator.AddDefinition(definitions...)

	// Set random seed for pick definitions and templates.
	rand.Seed(0)
    
    initialState := map[string]string{"Pronoun": "She"}

	// Generate method generate message according to added definitions.
	// First argument represent definition Type of start point.
	// Second argument represent initial state.
    // Third argument represent num of messages.
	messages, _ := generator.Generate("Root", initialState, 1)
    fmt.Println(messages[0])
}
```