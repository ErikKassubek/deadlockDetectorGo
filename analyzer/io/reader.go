// Copyrigth (c) 2024 Erik Kassubek
//
// File: reader.go
// Brief: Read trace files and create the internal trace
//
// Author: Erik Kassubek
// Created: 2023-08-08
//
// License: BSD-3-Clause

/*
Package reader provides functions for reading and processing log files.
*/
package io

import (
	"bufio"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"analyzer/analysis"
)

/*
 * Create the trace from all files in a folder.
 * Args:
 *   filePath (string): The path to the folder
 *   ignoreAtomics (bool): If atomic operations should be ignored
 * Returns:
 *   int: The number of routines
 *   bool: True if the trace contains any elems
 *   error: An error if the trace could not be created
 */
func CreateTraceFromFiles(filePath string, ignoreAtomics bool) (int, bool, error) {
	numberIds := 0

	println("Read trace from " + filePath)

	// traverse all files in the folder
	files, err := os.ReadDir(filePath)
	if err != nil {
		return 0, false, err
	}

	containsElems := false
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name() == "times.log" {
			continue
		}

		routine, err := getRoutineFromFileName(file.Name())
		if err != nil {
			return 0, containsElems, nil
		}
		numberIds = max(numberIds, routine)

		containsElem, err := CreateTraceFromFile(filePath+"/"+file.Name(), routine, ignoreAtomics)
		if err != nil {
			return 0, containsElems, err
		}
		if containsElem {
			containsElems = true
		}
	}

	analysis.Sort()

	return numberIds, containsElems, nil
}

/*
 * Read and build the trace from a file
 * Args:
 *   filePath (string): The path to the log file
 *   routine (int): The routine id
 *   ignoreAtomics (bool): If atomic operations should be ignored
 * Returns:
 *   bool: true if the trace contains any values
 *	 error: An error if the trace could not be created
 */
func CreateTraceFromFile(filePath string, routine int, ignoreAtomics bool) (bool, error) {
	log.Print("Create trace from file " + filePath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Print("Error opening file: " + filePath)
		return false, err
	}

	scanner := bufio.NewScanner(file)

	containsElem := false
	for scanner.Scan() {
		line := scanner.Text()
		processElement(line, routine, ignoreAtomics)
		containsElem = true
	}

	file.Close()

	if err := scanner.Err(); err != nil {
		return containsElem, err
	}

	return containsElem, nil
}

/*
 * Process one element from the log file.
 * Args:
 *   element (string): The element to process
 *   routine (int): The routine id, equal to the line number
 *   ignoreAtomics (bool): If atomic operations should be ignored
 * Returns:
 *   error: An error if the element could not be processed
 */
func processElement(element string, routine int, ignoreAtomics bool) error {
	if element == "" {
		return errors.New("Element is empty")
	}
	fields := strings.Split(element, ",")
	var err error
	switch fields[0] {
	case "A":
		if ignoreAtomics {
			return nil
		}
		err = analysis.AddTraceElementAtomic(routine, fields[1], fields[2], fields[3], fields[4])
	case "C":
		err = analysis.AddTraceElementChannel(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7], fields[8])
	case "M":
		err = analysis.AddTraceElementMutex(routine, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7])
	case "G":
		err = analysis.AddTraceElementFork(routine, fields[1], fields[2], fields[3])
	case "S":
		err = analysis.AddTraceElementSelect(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6])
	case "W":
		err = analysis.AddTraceElementWait(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6], fields[7])
	case "O":
		err = analysis.AddTraceElementOnce(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "N":
		err = analysis.AddTraceElementCond(routine, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "E":
		err = analysis.AddTraceElementRoutineEnd(routine, fields[1])
	default:
		return errors.New("Unknown element type in: " + element)
	}

	if err != nil {
		return err
	}

	return nil
}

func getRoutineFromFileName(fileName string) (int, error) {
	// the file name is "trace_routineID.log"
	// remove the .log at the end
	fileName1 := strings.TrimSuffix(fileName, ".log")
	if fileName1 == fileName {
		return 0, errors.New("File name does not end with .log")
	}

	fileName2 := strings.TrimPrefix(fileName1, "trace_")
	if fileName2 == fileName1 {
		return 0, errors.New("File name does not start with trace_")
	}

	routine, err := strconv.Atoi(fileName2)
	if err != nil {
		return 0, err
	}

	return routine, nil
}
