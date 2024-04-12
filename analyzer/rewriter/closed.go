package rewriter

import (
	"analyzer/bugs"
	"analyzer/trace"
	"errors"
)

/*
 * Given a send/recv on a closed channel, rewrite the trace to make the bug
 * occur. The trace is rewritten as follows:
 * We assume, that the send/recv on the closed channel did not actually occur
 * in the program run. Let c be the close and a the send or receive operation.
 * The global trace then has the form:
 * ~~~~
 * T = T1 ++ [a] ++ T2 ++ [c] ++ T3
 * ~~~~~~
 * We now, that a, c and all Elements in T2 are concurrent. Otherwise, a possible send/recv on close would not be possible. We can therefor reorder the trace in the following manner:
 * ~~~~
 * T = T1 ++ [X_s, c, a, X_e]
 * ~~~~~~
 * For send on close, this should lead to a crash of the program. For recv on close, it will probably lead to a different execution of program after the
 * object. We therefor disable the replay after c and a have been executed and
 * let the rest of the program run freely. To tell the replay to disable the
 * replay, by adding a stop character X_e.
 */

/*
* Create a new trace for send/recv on closed channel
* Let c be the close, a the send/recv, X a start marker, X' a stop marker and
* T1, T2, T3 partial traces.
* The trace before the rewrite looks as follows:
* 	T1 ++ [a] ++ T2 ++ [c] ++ T3
* We know, that a, c and all elements in T2 are concurrent. Otherwise the bug
* would not have been detected. We are also not interested in T3. For T2
* we only need the elements, that are before c. We call the subtrace with those
* elements T2'. We can therefore rewrite the trace as follows:
* 	T1 ++ T2' ++ [X, c, a, X']
* Args:
*   bug (Bug): The bug to create a trace for
* Returns:
*   error: An error if the trace could not be created
 */
func rewriteClosedChannel(bug bugs.Bug) error {
	println("Start rewriting trace for receive on closed channel...")

	if len(bug.TraceElement1) == 0 || bug.TraceElement1[0] == nil { // close
		return errors.New("TraceElement1 is nil") // send/recv
	}
	if len(bug.TraceElement2) == 0 || bug.TraceElement2[0] == nil {
		return errors.New("TraceElement2 is nil")
	}

	t1 := (*bug.TraceElement1[0]).GetTSort() // close
	t2 := (*bug.TraceElement2[0]).GetTSort() // send/recv

	if t1 < t2 { // actual close before send/recv
		return errors.New("Close is before send/recv")
	}

	// remove T3 -> T1 ++ [a] ++ T2 ++ [c]
	trace.ShortenTrace(t1, true)

	// transform T2 to T2' -> T1 ++ T2' ++ [c, a]
	// This is done by removing all elements in T2, that are concurrent to c (including a)
	// and then adding a after c
	trace.RemoveConcurrent(bug.TraceElement1[0])
	(*bug.TraceElement2[0]).SetTSort(t1 + 1)
	trace.AddElementToTrace(*bug.TraceElement2[0])

	// add a start and stop marker -> T1 ++ T2' ++ [X, c, a, X']
	trace.AddTraceElementReplay(t1-1, true)
	trace.AddTraceElementReplay(t1+2, false)

	return nil
}
