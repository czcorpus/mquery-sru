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

package test

import (
	"fmt"
	"os"
	"strings"
)

func GenerateTestCorpus(corpname string, wordCount int, tagCount int, wordsPerSentence int) error {
	registryPath := fmt.Sprintf("/var/lib/manatee/registry/%s", corpname)
	verticalPath := fmt.Sprintf("/var/lib/manatee/vert/%s.vert", corpname)
	dataPath := fmt.Sprintf("/var/lib/manatee/data/%s", corpname)

	err := generateVertical(corpname, wordCount, tagCount, wordsPerSentence, verticalPath)
	if err != nil {
		return err
	}
	return generateRegistry(corpname, dataPath, verticalPath, registryPath)

}

func generateVertical(corpname string, wordCount int, tagCount int, wordsPerSentence int, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "<s>")
	iTag := 0
	for i := 0; i < wordCount; i++ {
		tag := fmt.Sprintf("tag%d", iTag+1)
		word := fmt.Sprintf("word_%s%d_%s", corpname, i+1, tag)
		lemma := fmt.Sprintf("lemma_%s%d", corpname, i+1)
		line := strings.Join([]string{word, lemma, tag}, "\t")
		fmt.Fprintln(f, line)
		if i+1 == wordCount {
			fmt.Fprintln(f, "</s>")
		} else {
			if (i+1)%wordsPerSentence == 0 {
				fmt.Fprintln(f, "</s>")
				fmt.Fprintln(f, "<s>")
			}
			iTag++
			if iTag == tagCount {
				iTag = 0
			}
		}
	}
	return nil
}

func generateRegistry(corpname string, dataPath string, verticalPath string, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(
		f,
		`NAME	"%s"
PATH	"%s"
VERT	"%s"
LANGUAGE "Czech"
LOCALE   "cs_CZ.UTF-8"
ENCODING "utf-8"
INFO     "Testovací korpus"

ATTRIBUTE   word

ATTRIBUTE   lemma

ATTRIBUTE   tag

STRUCTURE	s

MAXCONTEXT 0
MAXDETAIL 0
`,
		corpname,
		dataPath,
		verticalPath,
	)
	return err
}
