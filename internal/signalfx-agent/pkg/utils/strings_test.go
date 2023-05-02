package utils

import "testing"

type SplitStringTest struct {
	input  string
	sep    rune
	escape rune
	want   []string
}

var SplitStringTests = []SplitStringTest{
	{input: "kafka.ex-jaeger-transaction.ok", sep: '.', escape: '\\', want: []string{"kafka", "ex-jaeger-transaction", "ok"}},
	{input: "kafka\\.ex-jaeger-transaction\\.ok", sep: '.', escape: '\\', want: []string{"kafka.ex-jaeger-transaction.ok"}},
	{input: "kafka\\.ex-jaeger-transaction.ok", sep: '.', escape: '\\', want: []string{"kafka.ex-jaeger-transaction", "ok"}},
	{input: "kafka\\\\.ex-jaeger-transaction.ok", sep: '.', escape: '\\', want: []string{"kafka\\", "ex-jaeger-transaction", "ok"}},
	{input: "kafka\\\\\\.ex-jaeger-transaction.ok", sep: '.', escape: '\\', want: []string{"kafka\\.ex-jaeger-transaction", "ok"}},
}

func TestSplitString(t *testing.T) {
	for _, test := range SplitStringTests {
		got, err := SplitString(test.input, test.sep, test.escape)
		if err != nil {
			t.Errorf("got error: %+v", err)
		}
		if len(got) != len(test.want) {
			t.Errorf("got %d substrings, want %d substrings", len(got), len(test.want))
		}
		for i := 0; i < len(got); i++ {
			if got[i] != test.want[i] {
				t.Errorf("got substring: %s, want substring: %s", got[i], test.want[i])
			}
		}
	}
}

var SplitStringTestsError = []SplitStringTest{
	{input: "kafka.ex-jaeger-transaction.ok\\", sep: '.', escape: '\\'},
}

func TestSplitStringError(t *testing.T) {
	for _, test := range SplitStringTestsError {
		got, err := SplitString(test.input, test.sep, test.escape)
		for i := 0; i < len(got); i++ {
			if err == nil {
				t.Error("got no error, want error")
			}
		}
	}
}
