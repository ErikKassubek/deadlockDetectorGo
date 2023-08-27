package trace

import (
	"fmt"
	"sort"
)

type sortByTPre []traceElement

func (a sortByTPre) Len() int           { return len(a) }
func (a sortByTPre) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortByTPre) Less(i, j int) bool { return a[i].getTpre() < a[j].getTpre() }

var trace []traceElement = make([]traceElement, 0)

/*
* Get the trace
* Returns:
*   []traceElement: The trace
 */
func GetTrace() []traceElement {
	return trace
}

/*
* Add an element to the trace
* Args:
*   routine (int): The routine id
*   element (TraceElement): The element to add
* Returns:
*   error: An error if the routine does not exist
 */
func addElementToTrace(routine int, element traceElement) error {
	trace = append(trace, element)
	return nil
}

func FindPartner() {
	for _, elem := range trace {
		switch e := elem.(type) {
		case traceElementChannel:
			e.findPartnerChannel()
		}
	}
}

func Sort() {
	sort.Sort(sortByTPre(trace))
}

func CheckTraceChannel() bool {
	res := true
	for i, element := range trace {
		switch elem := element.(type) {
		case traceElementChannel:
			if elem.partner == nil {
				fmt.Println(i, elem.toString(), "Error")
				res = false
			} else {
				fmt.Println(i, elem.toString(), "Ok")
			}
		}

	}
	return res
}
