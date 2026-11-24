package handlers

import (
	"reflect"
	"testing"
)

func TestM1R3IntakeAliasList(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{"nil sql null stays nil", nil, nil},
		{"explicit empty string yields no candidates", "", nil},
		{"whitespace only yields no candidates", "  ", nil},
		{"comma text", "甲公司,乙公司", []string{"甲公司", "乙公司"}},
		{"mixed separators and blanks", "甲公司、 乙公司 ，,丙公司", []string{"甲公司", "乙公司", "丙公司"}},
		{"legacy array", []interface{}{"甲公司", " ", nil, "乙公司"}, []string{"甲公司", "乙公司"}},
		{"typed array", []string{"甲公司"}, []string{"甲公司"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := intakeAliasList(tc.input)
			if tc.want == nil && got != nil {
				t.Fatalf("want nil, got %#v", got)
			}
			if tc.want != nil && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %#v, got %#v", tc.want, got)
			}
		})
	}
}
