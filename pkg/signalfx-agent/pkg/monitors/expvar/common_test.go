package expvar

import (
	"testing"

	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type SnakeCaseSliceTest struct {
	given string
	want  []string
}

var SnakeCaseSliceTests = []SnakeCaseSliceTest{
	{given: "System.Cpu", want: []string{"system", "cpu"}},
	{given: "System.Cpu[0]", want: []string{"system", "cpu[0]"}},
	{given: "System.Cpu[0].CacheGCCPUFraction", want: []string{"system", "cpu[0]", "cache_gccpu_fraction"}},
	{given: "System.Cpu.CacheGCCPUFraction", want: []string{"system", "cpu", "cache_gccpu_fraction"}},
}

func TestSnakeCaseSlice(t *testing.T) {
	for _, test := range SnakeCaseSliceTests {
		got, _ := utils.SplitString(test.given, '.', escape)
		got = snakeCaseSlice(got)
		want := test.want
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("got: %s, want: %s", got[i], want[i])
			}
		}
	}
}

type JoinWordsTest struct {
	given []string
	want  string
}

var JoinWordsTests = []JoinWordsTest{
	{given: []string{"mem", "alloc", "size"}, want: "mem_alloc_size"},
	{given: []string{"mem", "alloc", "2", "mallocs"}, want: "mem_alloc_mallocs"},
	{given: []string{"mem", "alloc", "3a", "frees"}, want: "mem_alloc_3a_frees"},
}

func TestJoinWords(t *testing.T) {
	for _, test := range JoinWordsTests {
		want, got := joinWords(test.given, "_"), test.want
		if got != test.want {
			t.Errorf("got: %s, want: %s", got, want)
		}
	}
}
