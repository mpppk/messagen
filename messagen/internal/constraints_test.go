package internal

import (
	"reflect"
	"regexp"
	"testing"
)

func TestRawConstraintValue_Compile(t *testing.T) {
	tests := []struct {
		name    string
		r       RawConstraintValue
		want    *regexp.Regexp
		wantErr bool
	}{
		{
			name:    "can compile",
			r:       ".*",
			want:    regexp.MustCompile(".*"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Compile()
			if (err != nil) != tt.wantErr {
				t.Errorf("Compile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Compile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawConstraintValue_Parse(t *testing.T) {
	type args struct {
		isRegExp bool
	}
	tests := []struct {
		name    string
		r       RawConstraintValue
		args    args
		want    *ConstraintValue
		wantErr bool
	}{
		{
			name: "can parse",
			r:    "aaa",
			args: args{
				isRegExp: false,
			},
			want: &ConstraintValue{
				Raw:      "aaa",
				IsRegExp: false,
				re:       nil,
			},
			wantErr: false,
		},
		{
			name: "can parse regexp",
			r:    ".*aaa",
			args: args{
				isRegExp: true,
			},
			want: &ConstraintValue{
				Raw:      ".*aaa",
				IsRegExp: true,
				re:       regexp.MustCompile(".*aaa"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Parse(tt.args.isRegExp)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConstraints(t *testing.T) {
	type args struct {
		raw RawConstraints
	}
	tests := []struct {
		name    string
		args    args
		want    *Constraints
		wantErr bool
	}{
		{
			name: "create new Constraints struct",
			args: args{
				raw: RawConstraints{"Key": "Value"},
			},
			want: &Constraints{
				raw:    RawConstraints{"Key": "Value"},
				defMap: map[DefinitionType]RawConstraintKey{"Key": "Key"},
				values: []*Constraint{
					newConstraintOrPanic("Key", "Value"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConstraints(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConstraints() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestConstraints_Get(t *testing.T) {
	type args struct {
		key RawConstraintKey
	}
	tests := []struct {
		name        string
		constraints *Constraints
		args        args
		want        RawConstraintValue
		want1       bool
	}{
		{
			name:        "",
			constraints: newConstraintsOrPanic(RawConstraints{"k1!": "v1", "k2": "v2"}),
			args: args{
				key: "k1!",
			},
			want:  "v1",
			want1: true,
		},
		{
			name:        "",
			constraints: newConstraintsOrPanic(RawConstraints{"k1!": "v1", "k2": "v2"}),
			args: args{
				key: "k3",
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.constraints.Get(tt.args.key)
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestConstraints_GetByDefinitionType(t *testing.T) {
	type args struct {
		defType DefinitionType
	}
	tests := []struct {
		name        string
		constraints *Constraints
		args        args
		want        RawConstraintValue
		want1       bool
	}{
		{
			name:        "",
			constraints: newConstraintsOrPanic(RawConstraints{"k1!": "v1", "k2": "v2"}),
			args: args{
				defType: "k1",
			},
			want:  "v1",
			want1: true,
		},
		{
			name:        "",
			constraints: newConstraintsOrPanic(RawConstraints{"k1!": "v1", "k2": "v2"}),
			args: args{
				defType: "k3",
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.constraints.GetByDefinitionType(tt.args.defType)
			if got != tt.want {
				t.Errorf("GetByDefinitionType() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetByDefinitionType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
