// DEDEGO-FILE-START

package runtime

import (
	"sync/atomic"
)

/*
 * Get a string representation of an uint32
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func uint64ToString(n uint64) string {
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return uint64ToString(n/10) + string(rune(n%10+'0'))
	}
}

/*
 * Get a string representation of an int64
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func int64ToString(n int64) string {
	if n < 0 {
		return "-" + int64ToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return int64ToString(n/10) + string(rune(n%10+'0'))

	}
}

/*
 * Get a string representation of an int32
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func int32ToString(n int32) string {
	if n < 0 {
		return "-" + int32ToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return int32ToString(n/10) + string(rune(n%10+'0'))

	}
}

/*
 * Get a string representation of an int
 * Args:
 * 	n: int to convert
 * Return:
 * 	string representation of the int
 */
func intToString(n int) string {
	if n < 0 {
		return "-" + intToString(-n)
	}
	if n < 10 {
		return string(rune(n + '0'))
	} else {
		return intToString(n/10) + string(rune(n%10+'0'))

	}
}

var dedegoCurrentRoutineId atomic.Uint64
var dedegoCurrentObjectId atomic.Uint64
var dedegoGlobalCounter atomic.Uint64

/*
 * Get a new id for a routine
 * Return:
 * 	new id
 */
func GetDedegoRoutineId() uint64 {
	return dedegoCurrentRoutineId.Add(1)
}

/*
 * Get a new id for a mutex, channel or waitgroup
 * Return:
 * 	new id
 */
func GetDedegoObjectId() uint64 {
	return dedegoCurrentObjectId.Add(1)
}

/*
 * Get a new counter
 * Return:
 * 	new counter value
 */
func GetDedegoCounter() uint64 {
	return dedegoGlobalCounter.Add(1)
}

// DEDEGO-FILE-END
