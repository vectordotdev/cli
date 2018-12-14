package main

// OrdinalColorScale maps strings to colors, returning the same color for the same string
type OrdinalColorScale struct {
	colors [][3]uint8
	index  map[string]int // TODO this should probably be a bounded cache
	curr   int
}

func NewOrdinalColorScale(colors [][3]uint8) *OrdinalColorScale {
	return &OrdinalColorScale{
		colors: colors,
		index:  map[string]int{},
	}
}

func (o *OrdinalColorScale) Get(s string) [3]uint8 {
	i, ok := o.index[s]
	if !ok {
		i = o.curr
		o.index[s] = i
		o.curr = (o.curr + 1) % len(o.colors)
	}

	return o.colors[i]
}
