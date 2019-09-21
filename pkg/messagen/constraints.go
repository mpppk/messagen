package messagen

import (
	"regexp"

	"golang.org/x/xerrors"
)

type RawConstraintValue string

func (r RawConstraintValue) Compile() (*regexp.Regexp, error) {
	re, err := regexp.Compile(string(r))
	if err != nil {
		return nil, xerrors.Errorf("failed to compile RawConstraintValue. invalid regexp(%s): %w", r, err)
	}
	return re, nil
}

func (r RawConstraintValue) Match(msg Message) bool {
	return string(r) == string(msg)
}

func (r RawConstraintValue) Parse(isRegExp bool) (*ConstraintValue, error) {
	v := &ConstraintValue{
		Raw: r,
	}
	if !isRegExp {
		return v, nil
	}
	v.IsRegExp = true
	re, err := r.Compile()
	if err != nil {
		return nil, xerrors.Errorf("failed to parse constraint value: %w", err)
	}
	v.re = re
	return v, nil
}

type ConstraintValue struct {
	Raw      RawConstraintValue
	IsRegExp bool
	re       *regexp.Regexp
}

func (c *ConstraintValue) Match(msg Message) bool {
	if c.IsRegExp {
		return c.re.MatchString(string(msg))
	}
	return c.Raw.Match(msg)
}

type RawConstraints map[RawConstraintKey]RawConstraintValue
type Constraints struct {
	raw    RawConstraints
	defMap map[DefinitionType]RawConstraintKey
}

func NewConstraints(raw RawConstraints) (*Constraints, error) {
	rawConstraints := raw
	if rawConstraints == nil {
		rawConstraints = RawConstraints{}
	}
	c := &Constraints{
		raw:    rawConstraints,
		defMap: map[DefinitionType]RawConstraintKey{},
	}
	for key, value := range raw {
		if err := c.Set(key, value); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Constraints) Set(rawKey RawConstraintKey, value RawConstraintValue) error {
	c.raw[rawKey] = value
	key, err := rawKey.Parse()
	if err != nil {
		return xerrors.Errorf("failed to parse RawConstraintKey: %w", err)
	}
	c.defMap[key.DefinitionType] = rawKey
	return nil
}

func (c *Constraints) Get(key RawConstraintKey) (RawConstraintValue, bool) {
	v, ok := c.raw[key]
	return v, ok
}

func (c *Constraints) GetByDefinitionType(defType DefinitionType) (RawConstraintValue, bool) {
	key, ok := c.defMap[defType]
	if !ok {
		return "", false
	}
	v, ok := c.raw[key]
	return v, ok
}

func (c *Constraints) ListUnsatisfied(state State) (*Constraints, error) {
	unsatisfiedConstraints, err := NewConstraints(nil)
	if err != nil {
		return nil, xerrors.Errorf("failed to create new constraints: %w", err)
	}

	for rkey, rvalue := range c.raw {
		key, err := rkey.Parse()
		if err != nil {
			return unsatisfiedConstraints, xerrors.Errorf("failed to parse raw constraints key: %w", err)
		}

		if ok, reason := key.IsValid(); !ok {
			return unsatisfiedConstraints,
				xerrors.Errorf("invalid constraints key is found(%s). reason: %s", key, reason)
		}

		value, err := rvalue.Parse(key.HasRegExpValue)
		if err != nil {
			return unsatisfiedConstraints, xerrors.Errorf("error occurred in ListUnsatisfied: %w", err)
		}

		msg, ok := state.Get(key.DefinitionType)
		if !ok && !key.IsAllowedToNotExist {
			if err := unsatisfiedConstraints.Set(rkey, rvalue); err != nil {
				return unsatisfiedConstraints, xerrors.Errorf("failed to set constraint key: %w", err)
			}
			continue
		}

		if ok := value.Match(msg); !ok {
			if err := unsatisfiedConstraints.Set(rkey, rvalue); err != nil {
				return unsatisfiedConstraints, xerrors.Errorf("failed to set constraint key: %w", err)
			}
			continue
		}

		// TODO: Add more checks
	}
	return unsatisfiedConstraints, nil
}

func (c *Constraints) AreSatisfied(state State) (bool, error) {
	constraints, err := c.ListUnsatisfied(state)
	if err != nil {
		return false, xerrors.Errorf("failed to check constraints: %w", err)
	}
	return len(constraints.raw) == 0, nil
}
