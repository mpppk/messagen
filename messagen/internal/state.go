package internal

import (
	"golang.org/x/xerrors"
)

type State map[string]Message

func (s State) Set(defType DefinitionType, msg Message) {
	s[string(defType)] = msg
}

func (s State) SetByConstraint(constraint *Constraint) (bool, error) {
	if !constraint.key.WillAddValue {
		return false, nil
	}

	msg, err := constraint.value.ToMessage()
	if err != nil {
		return false, xerrors.Errorf("failed to set constraint value to state: %w", err)
	}

	if _, ok := s.Get(constraint.key.DefinitionType); ok {
		return false, nil
	}

	s.Set(constraint.key.DefinitionType, msg)
	return true, nil
}

func (s State) SetByConstraints(constraints *Constraints) (int, error) {
	cnt := 0
	for rKey, rValue := range constraints.raw {
		constraint, err := NewConstraint(rKey, rValue)
		if err != nil {
			return -1, xerrors.Errorf("failed to set state from constraints: %w", err)
		}

		ok, err := s.SetByConstraint(constraint)
		if err != nil {
			return -1, xerrors.Errorf("failed to set state from constraints: %w", err)
		}
		if ok {
			cnt++
		}
	}
	return cnt, nil
}

func (s State) Get(defType DefinitionType) (Message, bool) {
	v, ok := s[string(defType)]
	return v, ok
}

func (s State) Copy() State {
	ns := State{}
	for key, value := range s {
		ns[key] = value
	}
	return ns
}
