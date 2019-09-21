package internal

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

func (c *ConstraintValue) ToMessage() (Message, error) {
	if c.IsRegExp {
		return "", xerrors.Errorf("failed to convert constraint value to message. value is RegExp")
	}
	return Message(c.Raw), nil
}

type RawConstraints map[RawConstraintKey]RawConstraintValue

type Constraint struct {
	key   *ConstraintKey
	value *ConstraintValue
}

func NewConstraint(rawKey RawConstraintKey, rawValue RawConstraintValue) (*Constraint, error) {
	key, err := rawKey.Parse()
	if err != nil {
		return nil, xerrors.Errorf("failed to create Constraint: %w", key)
	}

	if ok, _ := key.IsValid(); !ok {
		return nil, xerrors.Errorf("invalid constraints key is found(%s)", key)
	}

	value, err := rawValue.Parse(key.HasRegExpValue)
	if err != nil {
		return nil, xerrors.Errorf("failed to create Constraint: %w", key)
	}
	return &Constraint{key: key, value: value}, nil
}

func (c *Constraint) IsSatisfied(state State) bool {
	msg, ok := state.Get(c.key.DefinitionType)

	// ? operator check
	if !ok {
		return c.key.IsAllowedToNotExist
	}

	// ! operator check
	if c.key.MustNotExist && ok {
		return false
	}

	if ok := c.value.Match(msg); !ok {
		return false
	}

	return true
}

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
		constraint, err := NewConstraint(rkey, rvalue)
		if err != nil {
			return nil, xerrors.Errorf("error occurred in ListUnsatisfied: %w", err)
		}

		if !constraint.IsSatisfied(state) {
			if err := unsatisfiedConstraints.Set(rkey, rvalue); err != nil {
				return unsatisfiedConstraints, xerrors.Errorf("failed to set constraint key: %w", err)
			}
		}
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
