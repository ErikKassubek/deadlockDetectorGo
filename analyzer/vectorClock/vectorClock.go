package vectorClock

import (
	"analyzer/logging"
	"fmt"
	"runtime"
	"strconv"
)

/*
 * vectorClock is a vector clock
 * Fields:
 *   size (int): The size of the vector clock
 *   clock ([]int): The vector clock
 */
type VectorClock struct {
	size  int
	clock map[int]int
}

/*
 * Create a new vector clock
 * Args:
 *   size (int): The size of the vector clock
 * Returns:
 *   (vectorClock): The new vector clock
 */
func NewVectorClock(size int) VectorClock {
	c := make(map[int]int)
	for i := 1; i <= size; i++ {
		c[i] = 0
	}

	return VectorClock{
		size:  size,
		clock: c,
	}
}

func (vc VectorClock) GetSize() int {
	return vc.size
}

/*
 * Get a string representation of the vector clock
 * Returns:
 *   (string): The string representation of the vector clock
 */
func (vc VectorClock) ToString() string {
	str := "["
	for i := 1; i <= vc.size; i++ {
		str += fmt.Sprint(vc.clock[i])
		if i <= vc.size-1 {
			str += ", "
		}
	}
	str += "]"
	return str
}

/*
 * Increment the vector clock at the given position
 * Args:
 *   routine (int): The routine to increment
 * Returns:
 *   (vectorClock): The vector clock
 */
func (vc VectorClock) Inc(routine int) VectorClock {
	if vc.clock == nil {
		vc.clock = make(map[int]int)
	}
	vc.clock[routine]++
	return vc
}

/*
 * Update the vector clock with the received vector clock
 * Args:
 *   rec (vectorClock): The received vector clock
 * Returns:
 *   (vectorClock): The new vector clock
 */
func (vc VectorClock) Sync(rec VectorClock) VectorClock {
	if vc.size == 0 && rec.size == 0 {
		_, file, line, _ := runtime.Caller(1)
		logging.Debug("Sync of empty vector clocks: "+file+":"+strconv.Itoa(line), logging.ERROR)
	}

	if vc.size == 0 {
		vc = NewVectorClock(rec.size)
	}

	if rec.size == 0 {
		return vc.Copy()
	}
	copy := rec.Copy()
	for i := 1; i <= vc.size; i++ {
		if vc.clock[i] > copy.clock[i] {
			copy.clock[i] = vc.clock[i]
		}
	}

	return copy
}

/*
 * Create a copy of the vector clock
 * Returns:
 *   (vectorClock): The copy of the vector clock
 */
func (vc VectorClock) Copy() VectorClock {
	if vc.size == 0 {
		_, file, line, _ := runtime.Caller(1)
		logging.Debug("Copy of empty vector clock: "+file+":"+strconv.Itoa(line), logging.ERROR)
		panic("")
	}
	newVc := NewVectorClock(vc.size)
	for i := 1; i <= vc.size; i++ {
		newVc.clock[i] = vc.clock[i]
	}
	return newVc
}

/*
 * Get the happens before relation between two operations given there
 * vector clocks
 * Args:
 *   vc1 (vectorClock): The first vector clock
 *   vc2 (vectorClock): The second vector clock
 * Returns:
 *   happensBefore: The happens before relation between the two vector clocks
 */
func GetHappensBefore(vc1 VectorClock, vc2 VectorClock) HappensBefore {
	if isCause(vc1, vc2) {
		return Before
	}
	if isCause(vc2, vc1) {
		return After
	}
	return Concurrent
}

// func GetHappensBefore(pre1 *VectorClock, post1 *VectorClock,
// 	pre2 *VectorClock, post2 *VectorClock) HappensBefore {
// 	isCausePre1 := isCause(pre1, pre2)
// 	isCausePre2 := isCause(pre2, pre1)

// 	isCausePre := None
// 	if isCausePre1 {
// 		isCausePre = Before
// 	} else if isCausePre2 {
// 		isCausePre = After
// 	} else {
// 		return Concurrent
// 	}

// 	isCausePost1 := isCause(post1, post2)
// 	isCausePost2 := isCause(post2, post1)

// 	isCausePost := None
// 	if isCausePost1 {
// 		isCausePost = Before
// 	} else if isCausePost2 {
// 		isCausePost = After
// 	} else {
// 		return Concurrent
// 	}

// 	if isCausePre == isCausePost {
// 		return isCausePre
// 	}
// 	return Concurrent
// }

/*
 * Check if vc1 is a cause of vc2
 * Args:
 *   vc1 (vectorClock): The first vector clock
 *   vc2 (vectorClock): The second vector clock
 * Returns:
 *   bool: True if vc1 is a cause of vc2, false otherwise
 */
func isCause(vc1 VectorClock, vc2 VectorClock) bool {
	atLeastOneSmaller := false
	for i := 1; i <= vc1.size; i++ {
		if vc1.clock[i] > vc2.clock[i] {
			return false
		} else if vc1.clock[i] < vc2.clock[i] {
			atLeastOneSmaller = true
		}
	}
	return atLeastOneSmaller
}
