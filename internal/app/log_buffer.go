package app

type logBuffer struct {
	items       []LogViewItem
	searchLower []string
	start       int
}

func (b *logBuffer) Len() int {
	return len(b.items)
}

func (b *logBuffer) Reset() {
	b.items = nil
	b.searchLower = nil
	b.start = 0
}

func (b *logBuffer) ReleaseSearchCache() {
	b.searchLower = nil
}

func (b *logBuffer) EnsureSearchCache(build func(LogViewItem) string) {
	if len(b.searchLower) == len(b.items) {
		return
	}

	b.linearize()
	b.searchLower = make([]string, len(b.items))
	for index, item := range b.items {
		b.searchLower[index] = build(item)
	}
}

func (b *logBuffer) Append(item LogViewItem, lower string, limit int) (bool, int) {
	if limit <= 0 {
		b.linearize()
		b.items = append(b.items, item)
		b.appendSearchLower(lower)
		return false, 0
	}
	if len(b.items) < limit {
		b.linearize()
		b.items = append(b.items, item)
		b.appendSearchLower(lower)
		return false, 0
	}

	b.items[b.start] = item
	if len(b.searchLower) != 0 {
		b.searchLower[b.start] = lower
	}
	b.start = nextBufferIndex(b.start, len(b.items))
	return true, b.items[b.start].SourceIndex
}

func (b *logBuffer) AppendOrdered(dst []LogViewItem) []LogViewItem {
	if cap(dst) < len(b.items) {
		dst = make([]LogViewItem, 0, len(b.items))
	} else {
		dst = dst[:0]
	}
	if len(b.items) == 0 {
		return dst
	}
	if b.start == 0 {
		return append(dst, b.items...)
	}

	dst = append(dst, b.items[b.start:]...)
	return append(dst, b.items[:b.start]...)
}

func (b *logBuffer) Range(yield func(LogViewItem)) {
	if len(b.items) == 0 {
		return
	}

	b.rangeSegment(b.start, len(b.items), yield)
	if b.start == 0 {
		return
	}
	b.rangeSegment(0, b.start, yield)
}

func (b *logBuffer) RangeWithSearchLower(yield func(LogViewItem, string)) {
	if len(b.items) == 0 {
		return
	}

	b.rangeSegmentWithSearchLower(b.start, len(b.items), yield)
	if b.start == 0 {
		return
	}
	b.rangeSegmentWithSearchLower(0, b.start, yield)
}

func (b *logBuffer) linearize() {
	if b.start == 0 {
		return
	}

	b.items = b.AppendOrdered(nil)
	if len(b.searchLower) != 0 {
		reordered := make([]string, 0, len(b.searchLower))
		reordered = append(reordered, b.searchLower[b.start:]...)
		reordered = append(reordered, b.searchLower[:b.start]...)
		b.searchLower = reordered
	}
	b.start = 0
}

func (b *logBuffer) rangeSegment(start int, end int, yield func(LogViewItem)) {
	for index := start; index < end; index++ {
		yield(b.items[index])
	}
}

func (b *logBuffer) rangeSegmentWithSearchLower(start int, end int, yield func(LogViewItem, string)) {
	for index := start; index < end; index++ {
		yield(b.items[index], b.searchLower[index])
	}
}

func (b *logBuffer) appendSearchLower(lower string) {
	if len(b.searchLower) == 0 {
		return
	}
	b.searchLower = append(b.searchLower, lower)
}

func nextBufferIndex(index int, size int) int {
	index++
	if index == size {
		return 0
	}
	return index
}
