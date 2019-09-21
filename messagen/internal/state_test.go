package internal

import (
	"reflect"
	"testing"
)

func newConstraintOrPanic(rawKey RawConstraintKey, rawValue RawConstraintValue) *Constraint {
	c, err := NewConstraint(rawKey, rawValue)
	if err != nil {
		panic(err)
	}
	return c
}

func TestState_SetByConstraint(t *testing.T) {
	type args struct {
		constraint *Constraint
	}
	tests := []struct {
		name      string
		s         State
		args      args
		want      bool
		wantState State
		wantErr   bool
	}{
		{
			name: "",
			s:    State{"K1": "V1"},
			args: args{
				constraint: newConstraintOrPanic("K2", "V2"),
			},
			want:      false,
			wantState: State{"K1": "V1"},
			wantErr:   false,
		},
		{
			name: "",
			s:    State{"K1": "V1"},
			args: args{
				constraint: newConstraintOrPanic("K2+", "V2"),
			},
			want:      true,
			wantState: State{"K1": "V1", "K2": "V2"},
			wantErr:   false,
		},
		{
			name: "",
			s:    State{"K1": "V1"},
			args: args{
				constraint: newConstraintOrPanic("K1+", "V2"),
			},
			want:      false,
			wantState: State{"K1": "V1"},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SetByConstraint(tt.args.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetByConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SetByConstraint() got = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(tt.s, tt.wantState) {
				t.Errorf("SetByConstraint() = %#v, want %#v", tt.s, tt.wantState)
			}
		})
	}
}

func TestState_SetByConstraints(t *testing.T) {
	type args struct {
		constraints *Constraints
	}
	tests := []struct {
		name    string
		s       State
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "",
			s:    State{"K1": "V1"},
			args: args{
				constraints: newConstraintsOrPanic(RawConstraints{"K2+": "V2", "K3+": "V3"}),
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SetByConstraints(tt.args.constraints)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetByConstraints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SetByConstraints() got = %v, want %v", got, tt.want)
			}
		})
	}
}
