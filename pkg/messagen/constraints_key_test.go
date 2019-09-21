package messagen

import (
	"reflect"
	"testing"
)

func TestRawConstraintKeyRune_IsSpecial(t *testing.T) {
	tests := []struct {
		name string
		r    RawConstraintKeyRune
		want bool
	}{
		{
			name: "should return false if 'a' is specified",
			r:    RawConstraintKeyRune('a'),
			want: false,
		},
		{
			name: "should return true if '!' is specified",
			r:    RawConstraintKeyRune('!'),
			want: true,
		},
		{
			name: "should return true if '?' is specified",
			r:    RawConstraintKeyRune('?'),
			want: true,
		},
		{
			name: "should return true if '+' is specified",
			r:    RawConstraintKeyRune('+'),
			want: true,
		},
		{
			name: "should return true if '/' is specified",
			r:    RawConstraintKeyRune('/'),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsSpecial(); got != tt.want {
				t.Errorf("IsSpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawConstraintKey_toReversedRunes(t *testing.T) {
	tests := []struct {
		name              string
		r                 RawConstraintKey
		wantRawLabelRunes []RawConstraintKeyRune
	}{
		{
			name:              "should return reversed runes",
			r:                 "abc",
			wantRawLabelRunes: []RawConstraintKeyRune{'c', 'b', 'a'},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRawLabelRunes := tt.r.toReversedRunes(); !reflect.DeepEqual(gotRawLabelRunes, tt.wantRawLabelRunes) {
				t.Errorf("toReversedRunes() = %v, want %v", gotRawLabelRunes, tt.wantRawLabelRunes)
			}
		})
	}
}

func TestRawConstraintKey_Parse(t *testing.T) {
	tests := []struct {
		name    string
		r       RawConstraintKey
		want    *ConstraintKey
		wantErr bool
	}{
		{
			name: "can par￿se",
			r:    "Test",
			want: &ConstraintKey{
				Raw:                 "Test",
				DefinitionType:      "Test",
				HasRegExpValue:      false,
				IsAllowedToNotExist: false,
				MustNotExist:        false,
				WillAddValue:        false,
			},
			wantErr: false,
		},
		{
			name: "can par￿se !",
			r:    "Test!",
			want: &ConstraintKey{
				Raw:                 "Test!",
				DefinitionType:      "Test",
				HasRegExpValue:      false,
				IsAllowedToNotExist: false,
				MustNotExist:        true,
				WillAddValue:        false,
			},
			wantErr: false,
		},
		{
			name: "can par￿se ?",
			r:    "Test?",
			want: &ConstraintKey{
				Raw:                 "Test?",
				DefinitionType:      "Test",
				HasRegExpValue:      false,
				IsAllowedToNotExist: true,
				MustNotExist:        false,
				WillAddValue:        false,
			},
			wantErr: false,
		},
		{
			name: "can par￿se /",
			r:    "Test/",
			want: &ConstraintKey{
				Raw:                 "Test/",
				DefinitionType:      "Test",
				HasRegExpValue:      true,
				IsAllowedToNotExist: false,
				MustNotExist:        false,
				WillAddValue:        false,
			},
			wantErr: false,
		},
		{
			name: "can be par￿sed",
			r:    "Test?/",
			want: &ConstraintKey{
				Raw:                 "Test?/",
				DefinitionType:      "Test",
				HasRegExpValue:      true,
				IsAllowedToNotExist: true,
				MustNotExist:        false,
				WillAddValue:        false,
			},
			wantErr: false,
		},
		{
			name: "can be par￿sed",
			r:    "Test+",
			want: &ConstraintKey{
				Raw:                 "Test+",
				DefinitionType:      "Test",
				HasRegExpValue:      false,
				IsAllowedToNotExist: false,
				MustNotExist:        false,
				WillAddValue:        true,
			},
			wantErr: false,
		},
		{
			name:    "can be par￿sed",
			r:       "Test+/",
			want:    nil,
			wantErr: true,
		},
		// TODO: Add more tests
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Parse()
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
