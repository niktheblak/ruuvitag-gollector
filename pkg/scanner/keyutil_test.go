package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsKeys(t *testing.T) {
	r := ContainsKeys(map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	}, map[string]bool{
		"a": true,
		"b": true,
		"c": true,
	})
	assert.True(t, r)
	r = ContainsKeys(map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	}, map[string]bool{
		"a": true,
		"c": true,
	})
	assert.False(t, r)
}
