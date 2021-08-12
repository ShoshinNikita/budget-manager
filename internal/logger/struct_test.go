package logger

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStructToFields(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in         interface{}
		namePrefix string
		want       Fields
	}
	runTestCases := func(t *testing.T, testCases []testCase) {
		for _, tt := range testCases {
			tt := tt
			t.Run("", func(t *testing.T) {
				got := structToFields(tt.in, tt.namePrefix)
				require.Equal(t, tt.want, got)
			})
		}
	}

	t.Run("wrong types", func(t *testing.T) {
		runTestCases(t, []testCase{
			{in: 5, want: nil},
			{in: 15.7, want: nil},
			{in: "hello", want: nil},
			{in: func() error { return nil }, want: nil},
			{in: map[string]int{"1": 2}, want: nil},
		})
	})

	t.Run("basic", func(t *testing.T) {
		type S struct {
			A float64 `json:"a"`
			B string  `json:"b"`
			C map[string]float64
			E **int `json:"e"`
			F *float64
			g string
		}

		var (
			i    = 10
			iPtr = &i
		)
		runTestCases(t, []testCase{
			{
				in:   S{A: 0.15, B: "hello world", C: map[string]float64{"1": 2}, E: &iPtr, F: nil, g: "g"},
				want: Fields{"a": 0.15, "b": "hello world", "C": map[string]float64{"1": 2}, "e": 10, "F": "<nil>"},
			},
			{
				in:         S{A: 0, B: "", E: &iPtr},
				namePrefix: "p",
				want:       Fields{"p.a": 0.0, "p.b": "", "p.C": "<nil>", "p.e": 10, "p.F": "<nil>"},
			},
		})
	})

	t.Run("complex", func(t *testing.T) {
		type Embedded struct {
			A int `json:"a"`
			B int `json:"b"`
		}

		type Empty struct{}

		type Nested1 struct {
			Arr []string `json:"arr"`
		}

		type Nested struct {
			X      int     `json:"x"`
			Nested Nested1 `json:"nested1"`
		}

		type S struct {
			Embedded
			Empty

			A      string `json:"a"`
			Nested Nested `json:"nested"`
		}

		runTestCases(t, []testCase{
			{
				in:         S{Embedded: Embedded{A: 1, B: 2}, A: "qwerty", Nested: Nested{X: 3}},
				namePrefix: "req",
				want:       Fields{"req.Embedded.a": 1, "req.Embedded.b": 2, "req.a": "qwerty", "req.nested.x": 3, "req.nested.nested1.arr": "<nil>"},
			},
		})
	})

	t.Run("format", func(t *testing.T) {
		type S struct {
			T time.Time        `json:"t"`
			D time.Duration    `json:"d"`
			B *strings.Builder `json:"b"` // (*strings.Builder).String() panics if builder is nil
		}

		runTestCases(t, []testCase{
			{
				in: S{
					T: time.Date(2021, time.August, 11, 0, 0, 0, 0, time.UTC),
					D: 5 * time.Second,
				},
				want: Fields{
					"t": "2021-08-11T00:00:00Z",
					"d": "5s",
					"b": "<nil>",
				},
			},
			{
				in: S{
					T: time.Date(2021, time.August, 11, 0, 0, 0, 0, time.UTC),
					B: &strings.Builder{},
				},
				want: Fields{
					"t": "2021-08-11T00:00:00Z",
					"d": "0s",
					"b": "",
				},
			},
		})
	})
}
