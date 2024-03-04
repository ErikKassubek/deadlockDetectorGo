package runtime

/*
 * String representation of the replay operation.
 * Return:
 * 	string: string representation of the replay operation
 */
func (ro Operation) ToString() string {
	switch ro {
	case OperationNone:
		return "OperationNone"
	case OperationSpawn:
		return "OperationSpawn"
	case OperationSpawned:
		return "OperationSpawned"
	case OperationChannelSend:
		return "OperationChannelSend"
	case OperationChannelRecv:
		return "OperationChannelRecv"
	case OperationChannelClose:
		return "OperationChannelClose"
	case OperationMutexLock:
		return "OperationMutexLock"
	case OperationMutexUnlock:
		return "OperationMutexUnlock"
	case OperationMutexTryLock:
		return "OperationMutexTryLock"
	case OperationRWMutexLock:
		return "OperationRWMutexLock"
	case OperationRWMutexUnlock:
		return "OperationRWMutexUnlock"
	case OperationRWMutexTryLock:
		return "OperationRWMutexTryLock"
	case OperationRWMutexRLock:
		return "OperationRWMutexRLock"
	case OperationRWMutexRUnlock:
		return "OperationRWMutexRUnlock"
	case OperationRWMutexTryRLock:
		return "OperationRWMutexTryRLock"
	case OperationOnce:
		return "OperationOnce"
	case OperationWaitgroupAddDone:
		return "OperationWaitgroupAddDone"
	case OperationWaitgroupWait:
		return "OperationWaitgroupWait"
	case OperationSelect:
		return "OperationSelect"
	case OperationSelectCase:
		return "OperationSelectCase"
	case OperationSelectDefault:
		return "OperationSelectDefault"
	case OperationCondSignal:
		return "OperationCondSignal"
	case OperationCondBroadcast:
		return "OperationCondBroadcast"
	case OperationCondWait:
		return "OperationCondWait"
	default:
		return "Unknown"
	}
}

/*
 * The replay data structure.
 * The replay data structure is used to store the trace of the program.
 * op: identifier of the operation
 * time: time (tpre) of the operation
 * file: file in which the operation is executed
 * line: line number of the operation
 * blocked: true if the operation is blocked (never finised, tpost=0), false otherwise
 * suc: success of the opeartion
 *     - for mutexes: trylock operations true if the lock was acquired, false otherwise
 * 			for other operations always true
 *     - for once: true if the once was chosen (was the first), false otherwise
 *     - for others: always true
 * PFile: file of the partner (mainly for channel/select)
 * PLine: line of the partner (mainly for channel/select)
 * SelIndex: index of the select case (only for select, otherwise)
 */
type ReplayElement struct {
	Routine  int
	Op       Operation
	Time     int
	File     string
	Line     int
	Blocked  bool
	Suc      bool
	PFile    string
	PLine    int
	SelIndex int
}

type AdvocateReplayTrace []ReplayElement
type AdvocateReplayTraces map[uint64]AdvocateReplayTrace // routine -> trace

var replayEnabled bool
var replayLock mutex
var replayDone int
var replayDoneLock mutex

// read trace
var replayData = make(AdvocateReplayTraces, 0)
var numberElementsInTrace int
var traceElementPositions = make(map[string][]int) // file -> []line

var timeoutMessageCycle = 500 // approx. 10s

var timeOutCancel = false

/*
 * Add a replay trace to the replay data.
 * Arguments:
 * 	routine: the routine id
 * 	trace: the replay trace
 */
func AddReplayTrace(routine uint64, trace AdvocateReplayTrace) {
	if _, ok := replayData[routine]; ok {
		panic("Routine already exists")
	}
	replayData[routine] = trace

	numberElementsInTrace += len(trace)

	for _, e := range trace {
		if _, ok := traceElementPositions[e.File]; !ok {
			traceElementPositions[e.File] = make([]int, 0)
		}
		if !containsInt(traceElementPositions[e.File], e.Line) {
			traceElementPositions[e.File] = append(traceElementPositions[e.File], e.Line)
		}
	}
}

/*
 * Print the replay data.
 */
func (t AdvocateReplayTraces) Print() {
	for id, trace := range t {
		println("\nRoutine: ", id)
		trace.Print()
	}
}

/*
 * Print the replay trace for one routine.
 */
func (t AdvocateReplayTrace) Print() {
	for _, e := range t {
		println(e.Op.ToString(), e.Time, e.File, e.Line, e.Blocked, e.Suc)
	}
}

/*
 * Enable the replay.
 * Arguments:
 * 	timeout: true if the replay should be canceled after a timeout, false otherwise
 */
func EnableReplay(timeout bool) {
	timeOutCancel = timeout

	// run a background routine to check for timeout if no operation is executed
	go checkForTimeoutNoOperation()

	replayEnabled = true
	println("Replay enabled\n")
}

/*
 * Wait until all operations in the trace are executed.
 * This function should be called after the main routine is finished, to prevent
 * the program to terminate before the trace is finished.
 */
func WaitForReplayFinish() {
	timeoutCounter := 0
	for {
		timeoutCounter++
		lock(&replayDoneLock)
		if replayDone >= numberElementsInTrace {
			unlock(&replayDoneLock)
			break
		}
		unlock(&replayDoneLock)

		// check for timeout
		if timeoutCounter%timeoutMessageCycle == 0 {
			waitTime := intToString(int(10 * timeoutCounter / timeoutMessageCycle))
			warningMessage := "\nReplayWarning: Long wait time for finishing replay."
			warningMessage += "The main routine has already finished approx. "
			warningMessage += waitTime
			warningMessage += "s ago, but the trace still contains not executed operations.\n"
			warningMessage += "This can be caused by a stuck replay.\n"
			warningMessage += "Possible causes are:\n"
			warningMessage += "    - The program was altered between recording and replay\n"
			warningMessage += "    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n"
			warningMessage += "    - The program execution path depends on the order of not tracked operations\n"
			warningMessage += "    - The program execution depends on outside input, that was not exactly reproduced\n"
			warningMessage += "If you believe, the program is still running, you can continue to wait.\n"
			warningMessage += "If you believe, the program is stuck, you can cancel the program.\n"
			warningMessage += "If you suspect, that one of these causes is the reason for the long wait time, you can try to change the program to avoid the problem.\n"
			warningMessage += "If the problem persist, this message will be repeated every approx. 10s.\n\n"
			println(warningMessage)
			if timeOutCancel {
				panic("ReplayError: Replay stuck")
			}
		}

		slowExecution()
	}
}

/*
 * Wait until the correct operation is about to be executed.
 * Arguments:
 * 	op: the operation type that is about to be executed
 * 	skip: number of stack frames to skip
 * Return:
 * 	bool: true if trace replay is enabled, false otherwise
 * 	bool: true if the wait was released regularly, false if it was released
 * 		because replay is disables, the operation is ignored or the operation is invalid
 * 	chan ReplayElement: channel to receive the next replay element
 */
func WaitForReplay(op Operation, skip int) (bool, bool, ReplayElement) {
	if !replayEnabled {
		return false, false, ReplayElement{}
	}

	_, file, line, _ := Caller(skip)

	return WaitForReplayPath(op, file, line)
}

/*
 * Wait until the correct atomic operation is about to be executed.
 * Args:
 * 	op: the operation type that is about to be executed
 * 	index: index of the atomic operation
 * Return:
 * 	bool: true if trace replay is enabled, false otherwise
 * 	ReplayElement: the next replay element
 */
// func WaitForReplayAtomic(op int, index uint64) (bool, ReplayElement) {
// 	lock(&advocateAtomicMapLock)
// 	routine := advocateAtomicMapRoutine[index]
// 	unlock(&advocateAtomicMapLock)

// 	if !replayEnabled {
// 		return false, ReplayElement{}
// 	}

// 	for {
// 		next := getNextReplayElement()
// 		// print("Replay: ", next.Time, " ", next.Op, " ", op, " ", next.File, " ", file, " ", next.Line, " ", line, "\n")

// 		if next.Time != 0 {
// 			if int(next.Op) != op || uint64(next.Routine) != routine {
// 				continue
// 			}
// 		}

// 		lock(&replayLock)
// 		replayIndex++
// 		unlock(&replayLock)
// 		// println("Replay: ", next.Time, op, file, line)
// 		return true, next
// 	}
// }

/*
 * Wait until the correct operation is about to be executed.
 * Arguments:
 * 		op: the operation type that is about to be executed
 * 		file: file in which the operation is executed
 * 		line: line number of the operation
 * Return:
 * 	bool: true if trace replay is enabled, false otherwise
 * 	bool: true if the wait was released regularly, false if it was released
 * 		because replay is disables, the operation is ignored or the operation is invalid
 * 	chan ReplayElement: channel to receive the next replay element
 */
func WaitForReplayPath(op Operation, file string, line int) (bool, bool, ReplayElement) {
	if !replayEnabled {
		return false, false, ReplayElement{}
	}

	if AdvocateIgnore(op, file, line) {
		return true, false, ReplayElement{}
	}

	println("WaitForReplayPath", op.ToString(), file, line)
	timeoutCounter := 0
	for {
		nextRoutine, next := getNextReplayElement()

		// currentRoutine := currentRoutineA.id

		// all elements in the trace have been executed
		if nextRoutine == -1 {
			println("The program tried to execute an operation, although all elements in the trace have already been executed.\nDisable Replay")
			replayEnabled = false
			return false, false, ReplayElement{}
		}

		// if the routine is correct, but the next element is not the correct one,
		// the program has diverged from the trace
		// in this case, we print a message and disable the replay

		// print("Replay Wait:\n  Wait: ", op.ToString(), " ", file, " ", line,
		// 	"\n  Next: ", next.Op.ToString(), " ", next.File, " ", next.Line, "\n")

		if next.Time != 0 {
			if (next.Op != op && !correctSelect(next.Op, op)) ||
				next.File != file || next.Line != line {
				// println("Continue")

				timeoutCounter++

				checkForTimeout(timeoutCounter, file, line)
				slowExecution()
				continue
			}
		}

		foundReplayElement(nextRoutine)
		println("Replay Run : ", next.Time, op.ToString(), file, line)
		lock(&replayDoneLock)
		replayDone++
		unlock(&replayDoneLock)
		// _, next = getNextReplayElement()
		// println("Replay Next: ", next.Time, next.Op.ToString(), next.File, next.Line)
		return true, true, next
	}
}

/*
 * At specified timeoutCounter values, do some checks or messages.
 * At the first timeout, check if the position is in the trace.
 * At following timeouts, print a warning message.
 * At the last timeout, panic.
 * Args:
 * 	timeoutCounter: the current timeout counter
 * 	file: file in which the operation is executed
 * 	line: line number of the operation
 */
func checkForTimeout(timeoutCounter int, file string, line int) bool {
	messageCauses := "Possible causes are:\n"
	messageCauses += "    - The program was altered between recording and replay\n"
	messageCauses += "    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n"
	messageCauses += "    - The program execution path depends on the order of not tracked operations\n"
	messageCauses += "    - The program execution depends on outside input, that was not exactly reproduced\n"

	if timeoutCounter == 250 { // ca. 5s
		res := isPositionInTrace(file, line)
		if !res {
			errorMessage := "ReplayError: Program tried to execute an operation that is not in the trace:\n"
			errorMessage += "    File: " + file + "\n"
			errorMessage += "    Line: " + intToString(line) + "\n"
			errorMessage += "This means, that the program replay was not successful.\n"
			errorMessage += messageCauses
			errorMessage += "If you suspect, that one of these causes is the reason for the error, you can try to change the program to avoid the problem.\n"
			errorMessage += "If this is not possible, you can try to rerun the replay, hoping the error does not occur again.\n"
			errorMessage += "If this is not possible or does not work, the program replay is currently not possible.\n\n"

			panic(errorMessage)
		}
	} else if timeoutCounter%timeoutMessageCycle == 0 { // approx. every 10s
		waitTime := intToString(int(10 * timeoutCounter / timeoutMessageCycle))
		warningMessage := "\nReplayWarning: Long wait time of approx. "
		warningMessage += waitTime + "s.\n"
		warningMessage += "The following operation is taking a long time to execute:\n"
		warningMessage += "    File: " + file + "\n"
		warningMessage += "    Line: " + intToString(line) + "\n"
		warningMessage += "This can be caused by a stuck replay.\n"
		warningMessage += messageCauses
		warningMessage += "If you believe, the program is still running, you can continue to wait.\n"
		warningMessage += "If you believe, the program is stuck, you can cancel the program.\n"
		warningMessage += "If you suspect, that one of these causes is the reason for the long wait time, you can try to change the program to avoid the problem.\n"
		warningMessage += "If the problem persist, this message will be repeated every approx. 10s.\n\n"

		println(warningMessage)

		if timeOutCancel {
			panic("ReplayError: Replay stuck")
		}
	}

	return false
}

func checkForTimeoutNoOperation() {
	timeoutCounter := 0
	waitTime := 500 // approx. 10s
	warningMessage := "No traced operation has been executed for approx. 10s.\n"
	warningMessage += "This can be caused by a stuck replay.\n"
	warningMessage += "Possible causes are:\n"
	warningMessage += "    - The program was altered between recording and replay\n"
	warningMessage += "    - The program execution path is not deterministic, e.g. its execution path is determined by a random number\n"
	warningMessage += "    - The program execution path depends on the order of not tracked operations\n"
	warningMessage += "    - The program execution depends on outside input, that was not exactly reproduced\n"
	warningMessage += "If you believe, the program is still running, you can continue to wait.\n"
	warningMessage += "If you believe, the program is stuck, you can cancel the program.\n"
	warningMessage += "If you suspect, that one of these causes is the reason for the long wait time, you can try to change the program to avoid the problem.\n"
	warningMessage += "If the problem persist, this message will be repeated every approx. 10s.\n\n"

	for {
		lock(&timeoutLock)
		timeoutCounter++
		if timeoutCounter%waitTime == 0 {
			message := "\nReplayWarning: Long wait time of approx. "
			message += intToString(int(10*timeoutCounter/waitTime)) + "s.\n"
			message += warningMessage

			println(message)
			if timeOutCancel {
				panic("ReplayError: Replay stuck")
			}
		}
		unlock(&timeoutLock)
		slowExecution()
	}
}

/*
 * Check if the position is in the trace.
 * Args:
 * 	file: file in which the operation is executed
 * 	line: line number of the operation
 * Return:
 * 	bool: true if the position is in the trace, false otherwise
 */
func isPositionInTrace(file string, line int) bool {
	if _, ok := traceElementPositions[file]; !ok {
		return false
	}

	if !containsInt(traceElementPositions[file], line) {
		return false
	}

	return true
}

func correctSelect(next Operation, op Operation) bool {
	if op != OperationSelect {
		return false
	}

	if next != OperationSelectCase && next != OperationSelectDefault {
		return false
	}

	return true
}

func BlockForever() {
	gopark(nil, nil, waitReasonZero, traceEvNone, 1)
}

/*
 * Get the next replay element.
 * Return:
 * 	uint64: the routine of the next replay element or -1 if the trace is empty
 * 	ReplayElement: the next replay element
 */
func getNextReplayElement() (int, ReplayElement) {
	lock(&replayLock)
	defer unlock(&replayLock)

	routine := -1
	// set mintTime to max int
	minTime := int(^uint(0) >> 1)

	for id, trace := range replayData {
		if len(trace) == 0 {
			continue
		}
		elem := trace[0]
		if elem.Time < minTime {
			minTime = elem.Time
			routine = int(id)
		}
	}

	if routine == -1 {
		return -1, ReplayElement{}
	}

	return routine, replayData[uint64(routine)][0]
}

func foundReplayElement(routine int) {
	lock(&replayLock)
	defer unlock(&replayLock)

	// remove the first element from the trace for the routine
	replayData[uint64(routine)] = replayData[uint64(routine)][1:]
}
