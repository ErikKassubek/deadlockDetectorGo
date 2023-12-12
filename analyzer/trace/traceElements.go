package trace

// Interface for trace elements
type TraceElement interface {
	getTpre() int
	getTpost() int
	GetTSort() int
	SetTsort(tsort int)
	SetTsortWithoutNotExecuted(tsort int)
	GetRoutine() int
	GetPos() string
	ToString() string
	updateVectorClock()
}
