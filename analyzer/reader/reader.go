package reader

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"analyzer/logging"
	"analyzer/trace"
)

/*
 * Read and build the trace from a file
 * Args:
 *   file_path (string): The path to the log file
 *   buffer_size (int): The size of the buffer for the scanner
 * Returns:
 *   int: The number of routines in the trace
 */
func CreateTraceFromFile(file_path string, buffer_size int) int {
	logging.Debug("Create trace from file "+file_path+"...", logging.INFO)
	file, err := os.Open(file_path)
	if err != nil {
		logging.Debug("Error opening file: "+file_path, logging.ERROR)
		panic(err)
	}

	logging.Debug("Count number of routines...", logging.DEBUG)
	numberOfRoutines := 0
	scanner := bufio.NewScanner(file)
	mb := 1048576 // 1 MB
	scanner.Buffer(make([]byte, 0, buffer_size*mb), buffer_size*mb)
	for scanner.Scan() {
		numberOfRoutines++
	}
	file.Close()
	logging.Debug("Number of routines: "+strconv.Itoa(numberOfRoutines), logging.INFO)

	file2, err := os.Open(file_path)
	if err != nil {
		logging.Debug("Error opening file: "+file_path, logging.ERROR)
		panic(err)
	}

	logging.Debug("Create trace with "+strconv.Itoa(numberOfRoutines)+" routines...", logging.DEBUG)

	scanner = bufio.NewScanner(file2)
	scanner.Buffer(make([]byte, 0, buffer_size*mb), buffer_size*mb)
	routine := 0
	for scanner.Scan() {
		routine++
		line := scanner.Text()
		processLine(line, routine, numberOfRoutines)
	}

	if err := scanner.Err(); err != nil {
		logging.Debug("Error reading file line.", logging.ERROR)
		if err.Error() != "token too long" {
			logging.Debug("Reader buffer size to small. Increase with -b.", logging.ERROR)
		}
		panic(err)
	}

	trace.Sort()

	logging.Debug("Trace created", logging.INFO)
	return numberOfRoutines
}

/*
 * Process one line from the log file.
 * Args:
 *   line (string): The line to process
 *   routine (int): The routine id, equal to the line number
 *   numberOfRoutines (int): The number of routines in the log file
 */
func processLine(line string, routine int, numberOfRoutines int) {
	logging.Debug("Read routine "+strconv.Itoa(routine), logging.DEBUG)
	elements := strings.Split(line, ";")
	for _, element := range elements {
		processElement(element, routine, numberOfRoutines)
	}
}

/*
 * Process one element from the log file.
 * Args:
 *   element (string): The element to process
 *   routine (int): The routine id, equal to the line number
 *   numberOfRoutines (int): The number of routines in the log file
 */
func processElement(element string, routine int, numberOfRoutines int) {
	if element == "" {
		logging.Debug("Routine "+strconv.Itoa(routine)+" is empty", logging.DEBUG)
		return
	}
	fields := strings.Split(element, ",")
	var err error = nil
	switch fields[0] {
	case "A":
		err = trace.AddTraceElementAtomic(routine, numberOfRoutines, fields[1], fields[2], fields[3])
	case "C":
		err = trace.AddTraceElementChannel(routine, numberOfRoutines, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7], fields[8])
	case "M":
		err = trace.AddTraceElementMutex(routine, numberOfRoutines, fields[1], fields[2],
			fields[3], fields[4], fields[5], fields[6], fields[7])
	case "G":
		err = trace.AddTraceElementFork(routine, numberOfRoutines, fields[1], fields[2])
	case "S":
		err = trace.AddTraceElementSelect(routine, numberOfRoutines, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	case "W":
		err = trace.AddTraceElementWait(routine, numberOfRoutines, fields[1], fields[2], fields[3],
			fields[4], fields[5], fields[6], fields[7])
	case "O":
		err = trace.AddTraceElementOnce(routine, numberOfRoutines, fields[1], fields[2], fields[3],
			fields[4], fields[5])
	default:
		panic("Unknown element type in: " + element)
	}

	if err != nil {
		panic(err)
	}

}
