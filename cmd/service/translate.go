// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Institute of the Czech National Corpus,
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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/query/parser/basic"
	"github.com/czcorpus/mquery-sru/query/parser/fcsql"
)

func repl(translate func(string) error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error: %s, Bye.\n", err)
			return
		}
		input = strings.TrimSpace(input)
		if err := translate(input); err != nil {
			fmt.Println(err)
		}
	}
}

func translateBasicQuery(input string) error {
	ast, err := basic.ParseQuery(
		input,
		[]corpus.PosAttr{
			{
				ID:                "id1",
				Name:              "word",
				Layer:             "text",
				IsLayerDefault:    true,
				IsBasicSearchAttr: true,
			},
			{
				ID:                "id2",
				Name:              "lemma",
				Layer:             "lemma",
				IsBasicSearchAttr: true,
			},
			{
				ID:    "id3",
				Name:  "pos",
				Layer: "pos",
			},
		},
		corpus.StructureMapping{
			SentenceStruct:  "s",
			UtteranceStruct: "sp",
			ParagraphStruct: "p",
			TurnStruct:      "sp",
			TextStruct:      "doc",
			SessionStruct:   "doc",
		},
	)

	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}
	outQuery := ast.Generate()
	for i, err := range ast.Errors() {
		return fmt.Errorf("semantic error[%d]: %w", i, err)
	}
	println(outQuery)
	return nil
}

func translateFCSQuery(input string) error {
	ast, err := fcsql.ParseQuery(
		input,
		[]corpus.PosAttr{
			{
				ID:             "id1",
				Name:           "word",
				Layer:          "text",
				IsLayerDefault: true,
			},
			{
				ID:    "id2",
				Name:  "lemma",
				Layer: "lemma",
			},
			{
				ID:    "id3",
				Name:  "pos",
				Layer: "pos",
			},
		},
		corpus.StructureMapping{
			SentenceStruct:  "s",
			UtteranceStruct: "sp",
			ParagraphStruct: "p",
			TurnStruct:      "sp",
			TextStruct:      "doc",
			SessionStruct:   "doc",
		},
	)

	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}
	outQuery := ast.Generate()
	for i, err := range ast.Errors() {
		return fmt.Errorf("semantic error[%d]: %w", i, err)
	}
	println(outQuery)
	return nil
}
