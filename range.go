package timeline

import (
	"fmt"
	"time"
)

// TODO: alignment

// Rangeはタイムラインの範囲をあらわす。
// 空き時間か否かに関わらず範囲を扱う。
type Range struct {
	// 選択開始時間
	bp time.Time

	// 選択終了時間(含まない)
	ep time.Time
}

// タイムライン上から特定の範囲をあらわすRangeを返す
// 必ずbp <= epでなければならない
func NewRange(bp, ep time.Time) Range {
	if bp.After(ep) {
		panic("illegal argument: bp > ep")
	}
	return Range{bp, ep}
}

// 選択した期間を返す
func (r Range) Duration() time.Duration {
	return r.ep.Sub(r.bp)
}

// Rangeが同じ値であればtrueを返す
func (r Range) Equal(r1 Range) bool {
	return r.bp.Equal(r1.bp) && r.ep.Equal(r1.ep)
}

func (r Range) String() string {
	return fmt.Sprintf("%s .. %s", r.bp.String(), r.ep.String())
}
