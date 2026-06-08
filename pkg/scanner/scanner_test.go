package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testAddr1 = "CC:CA:7E:52:CC:34"
	testAddr2 = "FB:E1:B7:04:95:EE"
	testAddr3 = "E8:E0:C6:0B:B8:C5"
)

func TestContainsKeys(t *testing.T) {
	t.Parallel()

	assert.True(t, containsKeys(map[string]string{
		testAddr1: "Living Room",
		testAddr2: "Bedroom",
		testAddr3: "Bathroom",
	}, map[string]bool{
		testAddr1: true,
		testAddr2: true,
		testAddr3: true,
	}))
	assert.False(t, containsKeys(map[string]string{
		testAddr1: "Living Room",
		testAddr2: "Bedroom",
		testAddr3: "Bathroom",
	}, map[string]bool{
		testAddr1: true,
		testAddr2: true,
	}))
}
