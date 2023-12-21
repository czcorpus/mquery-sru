// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
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

package advanced

import (
	"fcs/general"
	"regexp"
	"strings"
)

const EOF = 0

type advancedTransformer struct {
	input       string
	parseResult node
	fcsError    *general.FCSError
}

func (t *advancedTransformer) Error(e string) {
	t.fcsError = &general.FCSError{
		Code:    general.CodeQueryCannotProcess,
		Ident:   e,
		Message: "Cannot process query",
	}
}

type tokenDef struct {
	regex *regexp.Regexp
	token int
}

var tokens = []tokenDef{
	{
		regex: regexp.MustCompile(`^NOT`),
		token: NOT,
	},
	{
		regex: regexp.MustCompile(`^AND`),
		token: AND,
	},
	{
		regex: regexp.MustCompile(`^OR`),
		token: OR,
	},
	{
		regex: regexp.MustCompile(`^PROX`),
		token: PROX,
	},
	{
		regex: regexp.MustCompile(`^\".*\"`),
		token: TERM,
	},
	{
		regex: regexp.MustCompile(`^[\w\d]*`),
		token: TERM,
	},
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func (t *advancedTransformer) Lex(lval *yySymType) int {
	// Skip spaces.
	for ; len(t.input) > 0 && isSpace(t.input[0]); t.input = t.input[1:] {
	}

	// Check if the input has ended.
	if len(t.input) == 0 {
		return EOF
	}

	// Check if one of the regular expressions matches.
	for _, tokDef := range tokens {
		str := tokDef.regex.FindString(t.input)
		if str != "" {
			t.input = t.input[len(str):]
			// Pass string content to the parser.
			switch tokDef.token {
			case TERM:
				lval.String = strings.Trim(str, "\"")
			default:
				lval.String = str
			}
			return tokDef.token
		}
	}

	// Otherwise return the next letter.
	ret := int(t.input[0])
	t.input = t.input[1:]
	return ret
}

func (t *advancedTransformer) CreateCQL(attr string) (string, *general.FCSError) {
	return t.parseResult.transform(attr)
}

func ParseQuery(input string) (*advancedTransformer, *general.FCSError) {
	t := &advancedTransformer{input: input}
	yyParse(t)
	if t.fcsError != nil {
		return nil, t.fcsError
	}
	return t, nil
}
