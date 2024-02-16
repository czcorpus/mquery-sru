// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
//                Faculty of Arts, Charles University
//   This file is part of MQUERY.
//
//  MQUERY is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  MQUERY is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with MQUERY.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MQuery-SRU Integration test script.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s corpgen <corpname> <wordCount> <tagCount> <wordsPerSentence>\n\t", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()
	action := flag.Arg(0)

	switch action {
	case "corpgen":
		corpName := flag.Arg(1)
		wordCount := flag.Arg(2)
		iWordCount, err := strconv.Atoi(wordCount)
		if err != nil {
			fmt.Printf("Invalid wordCount: %s", err)
			os.Exit(2)
		}
		tagCount := flag.Arg(3)
		iTagCount, err := strconv.Atoi(tagCount)
		if err != nil {
			fmt.Printf("Invalid tagCount: %s", err)
			os.Exit(2)
		}
		wordsPerSentence := flag.Arg(4)
		iWordsPerSentence, err := strconv.Atoi(wordsPerSentence)
		if err != nil {
			fmt.Printf("Invalid wordsPerSentence: %s", err)
			os.Exit(2)
		}
		err = GenerateTestCorpus("/var/lib/manatee", corpName, iWordCount, iTagCount, iWordsPerSentence)
		if err != nil {
			fmt.Printf("Failed to generate test corpus data: %s", err)
			os.Exit(2)
		}
		fmt.Println("Test corpus data generated")
		return
	default:
		fmt.Println("Unknown action")
		os.Exit(2)
	}
}
