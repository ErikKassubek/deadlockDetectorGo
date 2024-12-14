// Copyright (c) 2024 Erik Kassubek
//
// File: advocate_fuzzing.go
// Brief: Fuzzing
//
// Author: Erik Kassubek
// Created: 2024-12-10
//
// License: BSD-3-Clause

package advocate

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

/*
 * Initialize fuzzing
 * Args:
 * 	pathSelect (string): path to the file containing the select
 * 		preferred cases
 */
func InitFuzzing(pathSelect string) {
	prefSel, err := readFile(pathSelect)
	if err != nil {
		panic(err)
	}

	runtime.InitFuzzing(prefSel)
}

/*
 * Read the file containing the preferred select cases
 * Args:
 * 	pathSelect (string): path to the file containing the select
 * 		preferred cases
 * Returns:
 * 	map[string][]int: key: file:line of select, values: list of preferred cases
 * 	error
 */
func readFile(pathSelect string) (map[string][]int, error) {
	res := make(map[string][]int)

	file, err := os.Open(pathSelect)
	if err != nil {
		return res, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		elems := strings.Split(line, ";")
		if len(elems) != 2 {
			return res, fmt.Errorf("Incorrect line in fuzzing select file: %s", line)
		}

		ids := strings.Split(elems[1], ",")

		if len(ids) == 0 {
			continue
		}

		res[elems[0]] = make([]int, len(ids))
		for i, id := range ids {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				return res, err
			}
			res[elems[0]][i] = idInt
		}
	}

	return res, nil
}
