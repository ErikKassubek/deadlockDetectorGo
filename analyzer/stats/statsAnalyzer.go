// Copyright (c) 2024 Erik Kassubek
//
// File: statsAnalyzer.go
// Brief: Collect stats about the analysis and the replay
//
// Author: Erik Kassubek
// Created: 2024-09-20
// Last Changed 2024-09-20
//
// License: BSD-3-Clause

package stats

import (
	"analyzer/explanation"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Parse the analyzer and replay output to collect the corresponding information
 * Args:
 *     pathToResults (string): path to the advocateResult folder
 * Returns:
 *     map[string]int: map with information
 *     error
 */
func statsAnalyzer(pathToResults string) (map[string]map[string]int, error) {
	detected := map[string]int{
		"A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "P01": 0, "P02": 0,
		"P03": 0, "L01": 0, "L02": 0, "L03": 0, "L04": 0, "L05": 0, "L06": 0,
		"L07": 0, "L08": 0, "L09": 0, "L10": 0}
	replayWriten := map[string]int{
		"A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "P01": 0, "P02": 0,
		"P03": 0, "L01": 0, "L02": 0, "L03": 0, "L04": 0, "L05": 0, "L06": 0,
		"L07": 0, "L08": 0, "L09": 0, "L10": 0}
	replaySuccessful := map[string]int{
		"A01": 0, "A02": 0, "A03": 0, "A04": 0, "A05": 0, "P01": 0, "P02": 0,
		"P03": 0, "L01": 0, "L02": 0, "L03": 0, "L04": 0, "L05": 0, "L06": 0,
		"L07": 0, "L08": 0, "L09": 0, "L10": 0}

	res := map[string]map[string]int{
		"detected":         detected,
		"replayWritten":    replayWriten,
		"replaySuccessful": replaySuccessful,
	}

	err := filepath.Walk(pathToResults, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "bug.md" {
			println(path)
			err := processBugFile(path, &res)
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})

	return res, err
}

/*
 * Parse a bug file to get the information
 * Args:
 *     filePath (string): path to the bug file
 *     info (*map[string]map[string]int): map to store the info in
 * Returns:
 *     error
 */
func processBugFile(filePath string, info *map[string]map[string]int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bugType := ""

	// read the file
	scanner := bufio.NewScanner(file)
	println("\n\n\n\n")
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())

		println(text)

		// get detected bug
		if strings.HasPrefix(text, "# ") {
			textSplit := strings.Split(text, ": ")
			if len(textSplit) != 2 {
				continue
			}

			text = textSplit[1]

			bugType = explanation.GetCodeFromDescription(text)
			if bugType == "" {
				return fmt.Errorf("unknown error type %s", text)
			}
			(*info)["detected"][bugType]++
		}

		println("bugType: ", bugType)

		if text == "The rewritten trace can be found in the `rewritten_trace` folder." {
			(*info)["replayWritten"][bugType]++
		}

		if strings.HasPrefix(text, "It exited with the following code: ") {
			code := strings.TrimPrefix(text, "It exited with the following code: ")

			num, err := strconv.Atoi(code)
			if err != nil {
				return err
			}

			if num < 10 || num >= 20 {
				(*info)["replaySuccessful"][bugType]++
			}
		}
	}

	return nil
}
