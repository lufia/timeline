package timeline

import (
	"testing"
	"time"
)

func TestAlloc(t *testing.T) {
	tl := NewTimeline()
	tab := []struct {
		bp, ep     time.Time
		nalloc     time.Duration
		start, end time.Time
	}{
		// 連続した時間の間は必ず1分以上15分未満(設定可能にする？)の休憩を入れる
		{
			bp:     time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 11, 1, 0, 0, 0, time.Local),
			nalloc: 1 * time.Minute,
			start:  time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			end:    time.Date(2014, time.February, 11, 0, 1, 0, 0, time.Local),
		},
		{
			bp:     time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 11, 1, 0, 0, 0, time.Local),
			nalloc: 15 * time.Minute,
			start:  time.Date(2014, time.February, 11, 0, 15, 0, 0, time.Local),
			end:    time.Date(2014, time.February, 11, 0, 30, 0, 0, time.Local),
		},
		{
			bp:     time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 11, 1, 0, 0, 0, time.Local),
			nalloc: 2 * time.Minute,
			start:  time.Date(2014, time.February, 11, 0, 45, 0, 0, time.Local),
			end:    time.Date(2014, time.February, 11, 0, 47, 0, 0, time.Local),
		},
		{ // 後ろに10分間延ばす
			bp:     time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 11, 1, 10, 0, 0, time.Local),
			nalloc: 1 * time.Minute,
			start:  time.Date(2014, time.February, 11, 1, 0, 0, 0, time.Local),
			end:    time.Date(2014, time.February, 11, 1, 1, 0, 0, time.Local),
		},
		{ // 前に30分間延ばす
			bp:     time.Date(2014, time.February, 10, 23, 30, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 11, 1, 0, 0, 0, time.Local),
			nalloc: 30 * time.Minute,
			start:  time.Date(2014, time.February, 10, 23, 30, 0, 0, time.Local),
			end:    time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
		},
	}
	for i, c := range tab {
		t.Log("Case", i, "bp =", c.bp, "ep =", c.ep, "nalloc =", c.nalloc)
		r := NewRange(c.bp, c.ep)
		b, err := tl.Alloc(r, c.nalloc)
		if err != nil {
			t.Errorf("Alloc is able to allocate %v: %v", c.nalloc, err)
		}
		if cap := b.Capacity(); cap != c.nalloc {
			t.Errorf("Got %v Expect %v", cap, c.nalloc)
		}
		if !b.Start.Equal(c.start) {
			t.Errorf("Start: Got %v Expect %v", b.Start, c.start)
		}
		if !b.End.Equal(c.end) {
			t.Errorf("End: Got %v Expect %v", b.End, c.end)
		}
	}
}

func TestAllocAll(t *testing.T) {
	tl := NewTimeline()
	bp := time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local)
	ep := time.Date(2014, time.February, 11, 0, 1, 0, 0, time.Local)
	r := NewRange(bp, ep)
	nalloc := 1 * time.Minute
	b, err := tl.Alloc(r, nalloc)
	if err != nil {
		t.Errorf("Alloc is able to allocate %v: %v", nalloc, err)
	}
	if cap := b.Capacity(); cap != nalloc {
		t.Errorf("Got %v Expect %v", cap, nalloc)
	}
}

func TestAllocErr(t *testing.T) {
	tl := NewTimeline()
	tab := []struct {
		bp, ep time.Time
		nalloc time.Duration
	}{
		{ // Rangeより大きいので確保不可能
			bp:     time.Date(2014, time.February, 11, 0, 0, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 11, 0, 1, 0, 0, time.Local),
			nalloc: 1*time.Minute + 1*time.Nanosecond,
		},
		{ // Rangeの制限によりタイムラインの後半を未使用領域とする(1s足りない)
			bp:     time.Date(2014, time.February, 10, 0, 0, 0, 0, time.Local),
			ep:     time.Date(2014, time.February, 10, 0, 59, 59, 0, time.Local),
			nalloc: 1 * time.Hour,
		},
		{ // Rangeの制限によりタイムラインの前半を未使用領域とする(1ns足りない)
			bp:     time.Date(2014, time.February, 10, 0, 0, 0, 1, time.Local),
			ep:     time.Date(2014, time.February, 10, 1, 0, 0, 0, time.Local),
			nalloc: 1 * time.Hour,
		},
	}
	for i, c := range tab {
		t.Log("Case", i, "bp =", c.bp, "ep =", c.ep, "nalloc =", c.nalloc)
		r := NewRange(c.bp, c.ep)
		b, err := tl.Alloc(r, c.nalloc)
		if err == nil {
			t.Errorf("Alloc want an error but %v: %v", *b, i)
		}
	}
}

func TestAllocFragments(t *testing.T) {
	tl := NewTimeline()
	tab := []struct {
		bp, ep time.Time
		nalloc time.Duration
		expect []Range
	}{
		{
			bp:     Feb11(0, 0, 0, 0),
			ep:     Feb11(0, 1, 0, 0),
			nalloc: 1 * time.Minute,
			expect: []Range{
				NewRange(Feb11(0, 0, 0, 0), Feb11(0, 1, 0, 0)),
			},
		},
	}
	for i, c := range tab {
		t.Log("Case", i, "bp =", c.bp, "ep =", c.ep, "nalloc =", c.nalloc)
		r := NewRange(c.bp, c.ep)
		a, err := tl.AllocFragments(r, c.nalloc)
		if err != nil {
			t.Errorf("Alloc is able to allocate %v: %v", c.nalloc, err)
			continue
		}
		if len(a) != len(c.expect) {
			t.Errorf("Got %d blocks but expect %d blocks", len(a), len(c.expect))
		}
		for j, b := range a {
			if j >= len(c.expect) {
				t.Errorf("Block[%d]: Got %v Expect nothing", j, b.Range())
			} else if !b.Range().Equal(c.expect[j]) {
				t.Errorf("Block[%d]: Got %v Expect %v", j, b.Range(), c.expect[j])
			}
		}
	}
}

// テスト用にサンプル日時(日付は2014/2/11固定)を返す
func Feb11(hour, min, sec, nsec int) time.Time {
	return time.Date(2014, time.February, 11, hour, min, sec, nsec, time.Local)
}
