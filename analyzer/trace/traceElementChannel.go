package trace

import (
	"errors"
	"strconv"
)

// enum for opC
type opChannel int

const (
	send opChannel = iota
	recv
	close
)

/*
* traceElementChannel is a trace element for a channel
* Fields:
*   routine (int): The routine id
*   tpre (int): The timestamp at the start of the event
*   tpost (int): The timestamp at the end of the event
*   id (int): The id of the channel
*   opC (int, enum): The operation on the channel
*   exec (int, enum): The execution status of the operation
*   oId (int): The id of the other communication
*   qSize (int): The size of the channel queue
*   qCountPre (int): The number of elements in the queue before the operation
*   qCountPost (int): The number of elements in the queue after the operation
*   pos (string): The position of the channel operation in the code
*   partner (*traceElementChannel): The partner of the channel operation
 */
type traceElementChannel struct {
	routine int
	tpre    int
	tpost   int
	id      int
	opC     opChannel
	cl      bool
	oId     int
	qSize   int
	pos     string
	partner *traceElementChannel
}

/*
* Create a new channel trace element
* Args:
*   routine (int): The routine id
*   tpre (string): The timestamp at the start of the event
*   tpost (string): The timestamp at the end of the event
*   id (string): The id of the channel
*   opC (string): The operation on the channel
*   cl (string): Whether the channel was finished because it was closed
*   oId (string): The id of the other communication
*   qSize (string): The size of the channel queue
*   pos (string): The position of the channel operation in the code
 */
func AddTraceElementChannel(routine int, tpre string, tpost string, id string,
	opC string, cl string, oId string, qSize string, pos string) error {
	tpre_int, err := strconv.Atoi(tpre)
	if err != nil {
		return errors.New("tpre is not an integer")
	}

	tpost_int, err := strconv.Atoi(tpost)
	if err != nil {
		return errors.New("tpost is not an integer")
	}

	id_int, err := strconv.Atoi(id)
	if err != nil {
		return errors.New("id is not an integer")
	}

	var opC_int opChannel = 0
	switch opC {
	case "S":
		opC_int = send
	case "R":
		opC_int = recv
	case "C":
		opC_int = close
	default:
		return errors.New("opC is not a valid value")
	}

	cl_bool, err := strconv.ParseBool(cl)
	if err != nil {
		return errors.New("suc is not a boolean")
	}

	oId_int, err := strconv.Atoi(oId)
	if err != nil {
		return errors.New("oId is not an integer")
	}

	qSize_int, err := strconv.Atoi(qSize)
	if err != nil {
		return errors.New("qSize is not an integer")
	}

	elem := traceElementChannel{
		routine: routine,
		tpre:    tpre_int,
		tpost:   tpost_int,
		id:      id_int,
		opC:     opC_int,
		cl:      cl_bool,
		oId:     oId_int,
		qSize:   qSize_int,
		pos:     pos}

	return addElementToTrace(routine, elem)
}

/*
 * Get the routine of the element
 * Returns:
 *   int: The routine of the element
 */
func (elem traceElementChannel) getRoutine() int {
	return elem.routine
}

/*
 * Get the tpre of the element
 * Returns:
 *   int: The tpre of the element
 */
func (elem traceElementChannel) getTpre() int {
	return elem.tpre
}

/*
 * Get the tpost of the element
 * Returns:
 *   int: The tpost of the element
 */
func (elem traceElementChannel) getTpost() int {
	return elem.tpost
}

/*
 * Get the simple string representation of the element
 * Returns:
 *   string: The simple string representation of the element
 */
func (elem traceElementChannel) toString() string {
	return elem.toStringSep(",")
}

func (elem traceElementChannel) toStringSep(sep string) string {
	return "C," + strconv.Itoa(elem.tpre) + sep + strconv.Itoa(elem.tpost) + sep +
		strconv.Itoa(elem.id) + sep + strconv.Itoa(int(elem.opC)) + sep +
		strconv.Itoa(elem.oId) + sep + strconv.Itoa(elem.qSize) + sep + elem.pos
}

// map to store operations where partner has not jet been found
var channelOperations = make([]*traceElementChannel, 0)

func (elem *traceElementChannel) findPartnerChannel() {
	for i, e := range channelOperations {
		if e.id == elem.id && e.opC == elem.opC {
			elem.partner = e
			e.partner = elem
			// remove elem
			channelOperations = append(channelOperations[:i],
				channelOperations[i+1:]...)
			break
		}
	}
	if elem.partner == nil {
		println("append")
		channelOperations = append(channelOperations, elem)
	}
}
