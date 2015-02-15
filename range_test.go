package timeline

import (
	"testing"
	"time"
)

func TestRange(t *testing.T) {
	tab := []struct {
		bp, ep time.Time
		exp    time.Duration
	}{
		{
			bp:  time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:  time.Date(2014, time.February, 12, 0, 0, 0, 0, time.Local),
			exp: 24 * time.Hour,
		},
		{
			bp:  time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:  time.Date(2014, time.February, 11, 0, 0, 0, 1, time.Local),
			exp: 1 * time.Nanosecond,
		},
		{
			bp:  time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:  time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			exp: 0 * time.Nanosecond,
		},
	}
	for _, c := range tab {
		r := NewRange(c.bp, c.ep)
		if d := r.Duration(); d != c.exp {
			t.Errorf("Case[bp=%v, ep=%v]: got %v but %v", c.bp, c.ep, d, c.exp)
		}
	}
}

func TestRangeErr(t *testing.T) {
	bp := time.Date(2014, time.February, 11, 0, 0, 0, 1, time.Local)
	ep := time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local)
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("requires panic: because bp is later than ep")
		}
	}()
	NewRange(bp, ep)
}
