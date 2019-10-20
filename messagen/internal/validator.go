package internal

import "unicode/utf8"

type TemplateValidator = func(template *Template, state *State) (bool, error)

func MaxStrLenValidator(maxLen int) TemplateValidator {
	return func(template *Template, state *State) (bool, error) {
		incompleteMsg, _, err := template.ExecuteWithIncompleteState(state)
		if err != nil {
			return false, err
		}
		return utf8.RuneCountInString(string(incompleteMsg)) <= maxLen, nil
	}
}
