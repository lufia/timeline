package timeline // import "lufia.org/pkg/timeline"

import (
	"container/list"
	"errors"
	"time"
)

var (
	// 空き時間が足りなくなった場合のエラー
	Busy = errors.New("no free time")
)

// ひとつづきのタイムラインをあらわす
type Timeline struct {
	spare time.Duration
	v     *list.List
}

// 新規タイムラインを返す
func NewTimeline() *Timeline {
	return &Timeline{spare: 15 * time.Minute}
}

// Blockはタイムライン上の連続した範囲をあらわす
type Block struct {
	// 元となったタイムライン
	Timeline *Timeline

	// ブロックの開始日時
	Start time.Time

	// ブロックの終了日時(含まない)
	End time.Time

	// ブロックが使われているならtrue
	Retained bool
}

func (b *Block) Range() Range {
	return NewRange(b.Start, b.End)
}

// ブロックに含まれる時間の長さを返す
func (b *Block) Capacity() time.Duration {
	return b.Range().Duration()
}

// ブロックの中で、Rangeに含まれる部分の長さを返す
func (b *Block) CapacityInRange(r Range) time.Duration {
	if !b.InRange(r) {
		return 0 * time.Second
	}
	ep := b.End
	if r.ep.Before(ep) {
		ep = r.ep
	}
	bp := b.Start
	if r.bp.After(bp) {
		bp = r.bp
	}
	return ep.Sub(bp)
}

// ブロックの範囲とRangeの範囲が一部でも重なっていたらtrue
func (b *Block) InRange(r Range) bool {
	return b.Start.Before(r.ep) && b.End.After(r.bp)
}

// タイムライン上の最も古い時間を返す
func (tl *Timeline) Start() time.Time {
	if tl.v == nil {
		return time.Unix(0, 0)
	} else {
		first := tl.v.Front()
		block := first.Value.(*Block)
		return block.Start
	}
}

// タイムライン上の最も新しい時間(含まない)を返す
func (tl *Timeline) End() time.Time {
	if tl.v == nil {
		return time.Unix(0, 0)
	} else {
		last := tl.v.Back()
		block := last.Value.(*Block)
		return block.End
	}
}

func (tl *Timeline) growStart(bp time.Time) {
	if p := tl.Start(); bp.Before(p) {
		r := NewRange(bp, p)
		b := tl.newBlock(r)
		tl.v.PushFront(b)
	}
}

func (tl *Timeline) growEnd(ep time.Time) {
	if p := tl.End(); ep.After(p) {
		r := NewRange(p, ep)
		b := tl.newBlock(r)
		tl.v.PushBack(b)
	}
}

func (tl *Timeline) newBlock(r Range) *Block {
	return &Block{tl, r.bp, r.ep, false}
}

// Allocはタイムライン上の空き時間から、連続した範囲を確保できれば返す。
// 連続した範囲が足りない場合はBusyエラーを返す。
func (tl *Timeline) Alloc(r Range, n time.Duration) (*Block, error) {
	flist := tl.freelist(r)
	for e := flist.Front(); e != nil; e = e.Next() {
		free := e.Value.(*freeBlock)
		b := free.Block()
		if b.CapacityInRange(r) < n {
			continue
		}
		b = free.Retain(n)
		return b, nil
	}
	return nil, Busy
}

// AllocFragmentsはタイムライン上の空き時間から、nを満たすだけの複数ブロックを確保する。
// 空き時間が範囲内に足りない場合はBusyエラーを返す。
func (tl *Timeline) AllocFragments(r Range, n time.Duration) ([]*Block, error) {
	flist := tl.canContain(r, n)
	if flist == nil {
		return nil, Busy
	}
	a := make([]*Block, 0, 10)
	for e := flist.Front(); e != nil && n > 0; e = e.Next() {
		free := e.Value.(*freeBlock)
		nalloc := free.AlignedCapacityInRange(r)
		if nalloc > n {
			nalloc = n
		}
		b := free.Retain(nalloc)
		a = append(a, b)
		n -= nalloc
	}
	if n > 0 {
		panic("don't contain times")
	}
	return a, nil
}

// Rangeに重なる空きブロックの合計時間がnを含められる場合はフリーリストを返す
func (tl *Timeline) canContain(r Range, n time.Duration) *list.List {
	flist := tl.freelist(r)
	total := 0 * time.Second
	for e := flist.Front(); e != nil; e = e.Next() {
		free := e.Value.(*freeBlock)
		total += free.AlignedCapacityInRange(r)
	}
	if total < n {
		return nil
	}
	return flist
}

// Rangeに少しでも重なったブロックから、空きブロックのみ抽出して返す
func (tl *Timeline) freelist(r Range) *list.List {
	if tl.v == nil {
		tl.v = list.New()
		b := tl.newBlock(r)
		tl.v.PushBack(b)
	}
	tl.growStart(r.bp)
	tl.growEnd(r.ep)
	fl := list.New()
	for e := tl.v.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Block)
		if b.Start.After(r.ep) {
			break
		}
		if b.InRange(r) && !b.Retained {
			fl.PushBack(&freeBlock{e})
		}
	}
	return fl
}

type freeBlock struct {
	block *list.Element
}

func (f *freeBlock) Block() *Block {
	return f.block.Value.(*Block)
}

func (f *freeBlock) Timeline() *Timeline {
	return f.Block().Timeline
}

func (f *freeBlock) AlignedCapacityInRange(r Range) time.Duration {
	tl := f.Timeline()
	cap := f.Block().CapacityInRange(r)
	nbuf := cap % tl.spare
	if nbuf == 0 {
		cap -= tl.spare
		nbuf = tl.spare
	}
	if cap <= 0 {
		return 0
	} else {
		return cap
	}
}

// フリーブロックの先頭からn時間分を取得して、残りを返す。
// 取得後の残り時間が0になった場合はnilを返す。
func (f *freeBlock) retain1(n time.Duration) *freeBlock {
	b := f.Block()
	m := b.Start.Add(n)
	if m.Equal(b.End) { // ちょうど取得した
		b.Retained = true
		return nil
	}
	r := NewRange(m, b.End)
	tl := f.Timeline()
	newb := tl.newBlock(r)
	b.End = m
	b.Retained = true
	tl.v.InsertAfter(newb, f.block)
	return &freeBlock{f.block.Next()}
}

// フリーブロックの先頭からn時間分を取得(Retainedマーク)する。
// もし確保した後ろに一定時間分のバッファが取れればバッファ分も含めて取得する。
func (f *freeBlock) Retain(n time.Duration) *Block {
	newf := f.retain1(n)
	if newf != nil {
		// バッファ時間を確保する
		nb := newf.Block()
		tl := newf.Timeline()
		nbuf := tl.spare - n%tl.spare
		if nb.Capacity() < nbuf {
			nbuf = nb.Capacity()
		}
		newf.retain1(nbuf)
	}
	return f.Block()
}
