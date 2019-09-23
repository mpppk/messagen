package messagen

import (
	"github.com/mpppk/messagen/messagen/internal"
	"golang.org/x/xerrors"
)

func RandomTemplatePicker(templates *internal.Templates, state internal.State) (internal.Templates, error) {
	var newTemplates internal.Templates
	for {
		if len(*templates) == 0 {
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
