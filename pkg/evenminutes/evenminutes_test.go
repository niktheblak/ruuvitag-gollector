package evenminutes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNext(t *testing.T) {
	ts := time.Date(2019, time.January, 1, 12, 1, 12, 321, time.UTC)
	interval := 5 * time.Minute
	next := Next(ts, interval)
	assert.Equal(t, time.Date(2019, time.January, 1, 12, 5, 0, 0, time.UTC), next)

	ts = time.Date(2019, time.January, 1, 12, 13, 12, 321, time.UTC)
	next = Next(ts, interval)
	assert.Equal(t, time.Date(2019, time.January, 1, 12, 15, 0, 0, time.UTC), next)
}

func TestUntil(t *testing.T) {
	ts := time.Date(2019, time.January, 1, 12, 1, 12, 321, time.UTC)
	interval := 5 * time.Minute
	d := Until(ts, interval)
	expected := time.Date(2019, time.January, 1, 12, 5, 0, 0, time.UTC)
	assert.WithinDuration(t, expected, ts.Add(d), 200*time.Millisecond)
}

func TestLessThanMinute(t *testing.T) {
	ts := time.Date(2019, time.January, 1, 12, 1, 12, 321, time.UTC)
	interval := 1 * time.Second
	next := Next(ts, interval)
	assert.Equal(t, time.Date(2019, time.January, 1, 12, 1, 13, 321, time.UTC), next)
}
