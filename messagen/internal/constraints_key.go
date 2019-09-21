package internal

import (
	"strings"

	"golang.org/x/xerrors"
)

var specialConstraintKeyRunes = "!?/+"

func reverseRunes(runes []rune) []rune {
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return runes
}

type RawConstraintKeyRune rune

func (r RawConstraintKeyRune) IsSpecial() bool {
	return strings.ContainsRune(specialConstraintKeyRunes, rune(r))
}

type RawConstraintKey string

func (r RawConstraintKey) toReversedRunes() (rawLabelRunes []RawConstraintKeyRune) {
	runes := make([]rune, len([]rune(r)))
	copy(runes, []rune(r))
	runes = reverseRunes(runes)
	for _, r := range runes {
		rawLabelRunes = append(rawLabelRunes, RawConstraintKeyRune(r))
	}
	return
}

func (r RawConstraintKey) Parse() (*ConstraintKey, error) {
	constraintKey := &ConstraintKey{Raw: r}

	index := len(r)
	for i := len(r) - 1; i >= 0; i-- {
		ru := RawConstraintKeyRune([]rune(r)[i])
		if !ru.IsSpecial() {
			index = i
			break
		}
		if err := constraintKey.update(ru); err != nil {
			return nil, xerrors.Errorf("failed to parse constraint key: %w", err)
		}
	}
	constraintKey.DefinitionType = DefinitionType([]rune(r)[:index+1])
	if ok, reason := constraintKey.IsValid(); !ok {
		return nil, xerrors.Errorf("failed to parse constraint key: %s", reason)
	}
	return constraintKey, nil
}

type ConstraintKey struct {
	Raw                 RawConstraintKey
	DefinitionType      DefinitionType
	HasRegExpValue      bool
	IsAllowedToNotExist bool
	MustNotExist        bool
	WillAddValue        bool
}

func (l *ConstraintKey) update(rlr RawConstraintKeyRune) error {
	switch rlr {
	case '!':
		l.MustNotExist = true
	case '?':
		l.IsAllowedToNotExist = true
	case '/':
		l.HasRegExpValue = true
	case '+':
		l.WillAddValue = true
		l.IsAllowedToNotExist = true
	default:
		return xerrors.Errorf("unknown special constraint rune: %s", rlr)
	}
	return nil
}

func (l *ConstraintKey) IsValid() (bool, string) {
	if l.HasRegExpValue && l.WillAddValue {
		return false, "/ and + are exclusive"
	}
	if l.HasRegExpValue && l.MustNotExist {
		return false, "/ and ! are exclusive"
	}
	if l.MustNotExist && l.IsAllowedToNotExist {
		return false, "! and ? are exclusive"
	}
	return true, ""
}
