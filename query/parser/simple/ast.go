// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
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

package simple

import (
	"fcs/corpus"
	"fmt"
	"strings"
)

type Query struct {
	query            string
	cqlAttrs         []string
	structureMapping corpus.StructureMapping
	posAttrs         []corpus.PosAttr
	errors           []error
}

func (q *Query) SetCQLAttrs(attrs []string) *Query {
	q.cqlAttrs = attrs
	return q
}

func (q *Query) SetStructureMapping(m corpus.StructureMapping) *Query {
	q.structureMapping = m
	return q
}

func (q *Query) SetPosAttrs(attrs []corpus.PosAttr) *Query {
	q.posAttrs = attrs
	return q
}

func (q *Query) TranslateWithinCtx(v string) string {
	switch v {
	case "sentence", "s":
		return q.structureMapping.SentenceStruct
	case "utterance", "u":
		return q.structureMapping.UtteranceStruct
	case "paragraph", "p":
		return q.structureMapping.ParagraphStruct
	case "turn", "t":
		return q.structureMapping.TurnStruct
	case "text":
		return q.structureMapping.TextStruct
	case "session":
		return q.structureMapping.SessionStruct
	}
	return "??"
}

func (q *Query) TranslatePosAttr(qualifier, name string) string {
	if qualifier != "" {
		for _, p := range q.posAttrs {
			if p.Name == qualifier && string(p.Layer) == name {
				return p.Name
			}
		}

	} else {
		for _, p := range q.posAttrs {
			if string(p.Layer) == name && p.IsLayerDefault {
				return p.Name
			}
		}
	}
	q.AddError(fmt.Errorf("unknown attribute and/or layer %s:%s", qualifier, name))
	return ""
}

func (q *Query) AddError(err error) {
	q.errors = append(q.errors, err)
}

func (q *Query) Errors() []error {
	return q.errors
}

func (q *Query) Generate() string {
	ans := ""
	for _, v := range strings.Split(q.query, " ") {
		if v != "" {
			subquery := make([]string, len(q.cqlAttrs))
			for i, attr := range q.cqlAttrs {
				subquery[i] = fmt.Sprintf("%s=\"%s\"", attr, v)
			}
			ans += fmt.Sprintf("[%s]", strings.Join(subquery, "|"))
		}
	}
	return ans
}
