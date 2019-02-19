package split

import (
	"testing"
	"time"
)

func TestSplit(t *testing.T) {
	d, err := time.ParseDuration("30s")
	if err != nil {
		t.FailNow()
	}

	Execute("YOUR VIDEO FILE HERE", "../../results", d.Seconds())
}
