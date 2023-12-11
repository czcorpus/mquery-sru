// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package handler

import (
	"fmt"
	"strings"
)

func transformFCSQuery(fcs_query string) (string, error) {
	query := strings.ReplaceAll(fcs_query, "+", " ")                                        // convert URL spaces
	exactMatch := false                                                                     // attr=".*value.*"
	if strings.Contains(strings.ToLower(query), "exact") && !strings.Contains(query, "=") { // lemma EXACT dog
		pos := strings.Index(strings.ToLower(query), "exact") // first occurrence of EXACT
		query = query[:pos] + "=" + query[pos+5:]             // 1st exact > =
		exactMatch = true
	}

	var attr, term string
	if strings.Contains(query, "=") { // lemma=word | lemma="word" | lemma="w1 w2" | word=""
		items := strings.Split(query, "=")
		attr = strings.TrimSpace(items[0])
		term = strings.TrimSpace(items[1])
	} else { // "w1 w2" | "word" | word
		attr = "word" // TODO
		term = strings.TrimSpace(query)
	}

	if strings.Contains(attr, "\"") {
		return "", fmt.Errorf("attr contains invalid character `\"`")
	}

	rq := make([]string, 0, 10)
	if strings.Contains(term, "\"") { // "word" | "word1 word2" | "" | "it is \"good\""
		if term[0] != '"' || term[len(term)-1] != '"' { // check q. marks
			return "", fmt.Errorf("invalid attr value position of \"")
		}
		term = strings.TrimSpace(term[1 : len(term)-1])
		if term == "" {
			return "", fmt.Errorf("empty attr value")
		}
	} else if strings.Contains(term, " ") {
		return "", fmt.Errorf("multi-word terms has to be surrounded by \"")
	}

	for _, t := range strings.Split(term, " ") {
		if exactMatch {
			rq = append(rq, fmt.Sprintf("[%s=\"%s\"]", attr, t))
		} else {
			rq = append(rq, fmt.Sprintf("[%s=\".*%s.*\"]", attr, t))
		}
	}
	return strings.Join(rq, " "), nil
}
