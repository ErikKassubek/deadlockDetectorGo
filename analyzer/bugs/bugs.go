// Copyrigth (c) 2024 Erik Kassubek
//
// File: bugs.go
// Brief: Operations for handeling found bugs
//
// Author: Erik Kassubek <kassubek.erik@gmail.com>
// Created: 2023-11-30
// LastChange: 2024-09-01
//
// License: BSD-3-Clause

package bugs

import (
	"analyzer/trace"
	"errors"
	"strings"
)

type ResultType string

const (
	Empty ResultType = ""

	// actual
	ASendOnClosed          ResultType = "A01"
	ARecvOnClosed          ResultType = "A02"
	ACloseOnClosed         ResultType = "A03"
	AConcurrentRecv        ResultType = "A04"
	ASelCaseWithoutPartner ResultType = "A05"

	// possible
	PSendOnClosed ResultType = "P01"
	PRecvOnClosed ResultType = "P02"
	PNegWG        ResultType = "P03"

	// leaks
	LUnbufferedWith    = "L01"
	LUnbufferedWithout = "L02"
	LBufferedWith      = "L03"
	LBufferedWithout   = "L04"
	LNilChan           = "L05"
	LSelectWith        = "L06"
	LSelectWithout     = "L07"
	LMutex             = "L08"
	LWaitGroup         = "L09"
	LCond              = "L10"
)

type Bug struct {
	Type          ResultType
	TraceElement1 []*trace.TraceElement
	TraceElement2 []*trace.TraceElement
}

/*
 * Convert the bug to a string
 * Returns:
 *   string: The bug as a string
 */
func (b Bug) ToString() string {
	typeStr := ""
	arg1Str := ""
	arg2Str := ""
	switch b.Type {
	case ASendOnClosed:
		typeStr = "Found send on closed channel:"
		arg1Str = "send: "
		arg2Str = "close: "
	case ARecvOnClosed:
		typeStr = "Found receive on closed channel:"
		arg1Str = "recv: "
		arg2Str = "close: "
	case ACloseOnClosed:
		typeStr = "Found close on closed channel:"
		arg1Str = "close: "
		arg2Str = "close: "
	case AConcurrentRecv:
		typeStr = "Found concurrent Recv on same channel:"
		arg1Str = "recv: "
		arg2Str = "recv: "
	case ASelCaseWithoutPartner:
		typeStr = "Found select case without partner or nil case:"
		arg1Str = "select: "
		arg2Str = "case: "

	case PSendOnClosed:
		typeStr = "Possible send on closed channel:"
		arg1Str = "send: "
		arg2Str = "close: "
	case PRecvOnClosed:
		typeStr = "Possible receive on closed channel:"
		arg1Str = "recv: "
		arg2Str = "close: "
	case PNegWG:
		typeStr = "Possible negative waitgroup counter:"
		arg1Str = "add: "
		arg2Str = "done: "

	case LUnbufferedWith:
		typeStr = "Leak on unbuffered channel with possible partner:"
		arg1Str = "channel: "
		arg2Str = "partner: "
	case LUnbufferedWithout:
		typeStr = "Leak on unbuffered channel without possible partner:"
		arg1Str = "channel: "
		arg2Str = ""
	case LBufferedWith:
		typeStr = "Leak on buffered channel with possible partner:"
		arg1Str = "channel: "
		arg2Str = "partner: "
	case LBufferedWithout:
		typeStr = "Leak on buffered channel without possible partner:"
		arg1Str = "channel: "
		arg2Str = ""
	case LNilChan:
		typeStr = "Leak on nil channel:"
		arg1Str = "channel: "
		arg2Str = ""
	case LSelectWith:
		typeStr = "Leak on select with possible partner:"
		arg1Str = "select: "
		arg2Str = "partner: "
	case LSelectWithout:
		typeStr = "Leak on select without partner:"
		arg1Str = "select: "
		arg2Str = ""
	case LMutex:
		typeStr = "Leak on mutex:"
		arg1Str = "mutex: "
		arg2Str = "last: "
	case LWaitGroup:
		typeStr = "Leak on wait group:"
		arg1Str = "waitgroup: "
		arg2Str = ""
	case LCond:
		typeStr = "Leak on conditional variable:"
		arg1Str = "cond: "
		arg2Str = ""

	default:
		panic("Unknown bug type: " + string(b.Type))
	}

	res := typeStr + "\n\t" + arg1Str
	for i, elem := range b.TraceElement1 {
		if i != 0 {
			res += ";"
		}
		res += (*elem).GetTID()
	}

	if arg2Str != "" {
		res += "\n\t" + arg2Str

		if len(b.TraceElement2) == 0 {
			res += "-"
		}

		for i, elem := range b.TraceElement2 {
			if i != 0 {
				res += ";"
			}
			res += (*elem).GetTID()
		}
	}

	return res
}

/*
 * Print the bug
 */
func (b Bug) Println() {
	println(b.ToString())
}

/*
 * Process the bug that was selected from the analysis results
 * Args:
 *   bugStr: The bug that was selected
 * Returns:
 *   bool: true, if the bug was not a possible, but a actually occuring bug
 *   Bug: The bug that was selected
 *   error: An error if the bug could not be processed
 */
func ProcessBug(bugStr string) (bool, Bug, error) {
	bug := Bug{}

	bugSplit := strings.Split(bugStr, ",")
	if len(bugSplit) != 3 && len(bugSplit) != 2 {
		return false, bug, errors.New("Could not split bug: " + bugStr)
	}

	bugType := bugSplit[0]

	containsArg2 := true
	actual := false

	switch bugType {
	case "A01":
		bug.Type = ASendOnClosed
		actual = true
	case "A02":
		bug.Type = ARecvOnClosed
		actual = true
	case "A03":
		bug.Type = ACloseOnClosed
		actual = true
	case "A04":
		bug.Type = AConcurrentRecv
		actual = true
	case "A05":
		bug.Type = ASelCaseWithoutPartner
		actual = true
	case "P01":
		bug.Type = PSendOnClosed
	case "P02":
		bug.Type = PRecvOnClosed
	case "P03":
		bug.Type = PNegWG
	// case "P4":
	// 	bug.Type = CyclicDeadlock
	// case "P5":
	// 	bug.Type = MixedDeadlock
	case "L01":
		bug.Type = LUnbufferedWith
	case "L02":
		bug.Type = LUnbufferedWithout
		containsArg2 = false
	case "L03":
		bug.Type = LBufferedWith
	case "L04":
		bug.Type = LBufferedWithout
		containsArg2 = false
	case "L05":
		bug.Type = LNilChan
		containsArg2 = false
	case "L06":
		bug.Type = LSelectWith
	case "L07":
		bug.Type = LSelectWithout
		containsArg2 = false
	case "L08":
		bug.Type = LMutex
	case "L09":
		bug.Type = LWaitGroup
		containsArg2 = false
	case "L10":
		bug.Type = LCond
		containsArg2 = false
	default:
		return actual, bug, errors.New("Unknown bug type: " + bugStr)
	}

	bugArg1 := bugSplit[1]
	bugArg2 := ""
	if containsArg2 {
		bugArg2 = bugSplit[2]
	}

	bug.TraceElement1 = make([]*trace.TraceElement, 0)

	for _, bugArg := range strings.Split(bugArg1, ";") {
		if strings.TrimSpace(bugArg) == "" {
			continue
		}

		elem, err := trace.GetTraceElementFromBugArg(bugArg)
		if err != nil {
			println("Could not find: " + bugArg + " in trace")
			return actual, bug, err
		}
		bug.TraceElement1 = append(bug.TraceElement1, elem)
	}

	bug.TraceElement2 = make([]*trace.TraceElement, 0)

	if !containsArg2 {
		return actual, bug, nil
	}

	for _, bugArg := range strings.Split(bugArg2, ";") {
		if strings.TrimSpace(bugArg) == "" {
			continue
		}

		if bugArg[0] == 'T' {
			elem, err := trace.GetTraceElementFromBugArg(bugArg)
			if err != nil {
				return actual, bug, err
			}

			bug.TraceElement2 = append(bug.TraceElement2, elem)
		}
	}

	return actual, bug, nil
}
