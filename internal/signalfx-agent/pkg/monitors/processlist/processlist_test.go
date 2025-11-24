package processlist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEscapedCharacters(t *testing.T) {
	m := Monitor{
		lastCPUCounts: make(map[procKey]time.Duration),
	}
	top := &TopProcess{
		CreatedTime: time.Now(),
		Nice:        new(int),
		Username:    "test\"with nested\"quotes",
		Status:      "running",
		Command:     "echo test with line break \n character",
		ProcessID:   200,
	}
	encodedVal := m.encodeProcess(top, 10*time.Second)
	expectedVal := "\"200\":[\"test'with nested'quotes\",0,\"0\",0,0,0,\"running\",0.00,0.00,\"00:00.00\",\"echo test with line break \\n character\"]"
	assert.Equal(t, encodedVal, expectedVal)
}
