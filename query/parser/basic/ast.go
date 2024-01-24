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

package basic

import (
	"errors"
	"fcs/corpus"
	"fmt"
	"strings"
)

type Query struct {
	binaryOperatorQuery *binaryOperatorQuery
	structureMapping    corpus.StructureMapping
	posAttrs            []corpus.PosAttr
	errors              []error
}

func (q *Query) getDefaultAttrsExp(word string) string {
	var ans strings.Builder
	for i, p := range q.posAttrs {
		if p.IsBasicSearchAttr {
			if i > 0 {
				ans.WriteString(fmt.Sprintf(` | %s="%s"`, p.Name, word))

			} else {
				ans.WriteString(fmt.Sprintf(`%s="%s"`, p.Name, word))
			}
		}
	}
	return "[" + ans.String() + "]"
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
	return q.binaryOperatorQuery.Generate(q)
}

// -----

type binaryOperatorQueryRest struct {
	operation         string
	nonRecursiveQuery *nonRecursiveQuery
}

type binaryOperatorQuery struct {
	nonRecursiveQuery *nonRecursiveQuery
	rest              []*binaryOperatorQueryRest
}

func (boq *binaryOperatorQuery) AddRest(op string, nrq *nonRecursiveQuery) {
	boq.rest = append(boq.rest, &binaryOperatorQueryRest{operation: op, nonRecursiveQuery: nrq})
}

func (boq *binaryOperatorQuery) operatorAt(idx int) string {
	if idx < len(boq.rest) {
		return boq.rest[idx].operation
	}
	return ""
}

func (boq *binaryOperatorQuery) Generate(ast *Query) string {
	var rest strings.Builder
	for _, v := range boq.rest {
		rest.WriteString(" " + v.nonRecursiveQuery.Generate(ast))
	}
	if boq.operatorAt(0) == "AND" {
		return fmt.Sprintf(
			"(%s within ([]{0,10} %s []{0,10} within <%s />))",
			boq.nonRecursiveQuery.Generate(ast),
			rest.String(),
			ast.structureMapping.SentenceStruct,
		)

	} else if boq.operatorAt(0) == "OR" {
		return fmt.Sprintf(
			"(%s | %s)",
			boq.nonRecursiveQuery.Generate(ast),
			rest.String(),
		)

	} else if len(boq.rest) == 0 {
		return boq.nonRecursiveQuery.Generate(ast)

	} else {
		return fmt.Sprintf(
			"(?? %s %s)",
			boq.nonRecursiveQuery.Generate(ast),
			rest.String(),
		)
	}
}

// ----

type nonRecursiveQuery struct {
	parenthesisExpr     *parenthesisExpr
	unaryOperator       string
	binaryOperatorQuery *binaryOperatorQuery
	term                *term
}

func (nrq *nonRecursiveQuery) Generate(ast *Query) string {
	if nrq.parenthesisExpr != nil {
		return nrq.parenthesisExpr.Generate(ast)
	}
	if nrq.binaryOperatorQuery != nil {
		return fmt.Sprintf(
			"%s %s",
			nrq.unaryOperator,
			nrq.binaryOperatorQuery.Generate(ast),
		)
	}
	if nrq.term != nil {
		return nrq.term.Generate(ast)
	}
	ast.AddError(errors.New("invalid nonRecursiveQuery state"))
	return "??"
}

// ----

type parenthesisExpr struct {
	binaryOperatorQuery *binaryOperatorQuery
}

func (pe *parenthesisExpr) Generate(ast *Query) string {
	// NOTE: We don't need to generate parentheses here
	// ans the only contained non-terminal is binaryOperatorQuery
	// and it always produces an expression in parentheses.
	// And we don't want double ones.
	return pe.binaryOperatorQuery.Generate(ast)
}

// ---

type term struct {
	text       *text
	quotedText *quotedText
}

func (t *term) Generate(ast *Query) string {
	if t.text != nil {
		return t.text.Generate(ast)
	}
	if t.quotedText != nil {
		return t.quotedText.Generate(ast)
	}
	ast.AddError(errors.New("invalid term state"))
	return "??"
}

// ----

type quotedText struct {
	words []*word
}

func (qt *quotedText) Generate(ast *Query) string {
	var ans strings.Builder
	for _, v := range qt.words {
		ans.WriteString(" " + ast.getDefaultAttrsExp(v.Generate(ast)))
	}
	return ans.String()
}

func (qt *quotedText) AddWord(w *word) {
	qt.words = append(qt.words, w)
}

// -----

type text struct {
	word *word
}

func (t *text) Generate(ast *Query) string {
	return ast.getDefaultAttrsExp(t.word.Generate(ast))
}

// ------

type word struct {
	value string
}

func (w *word) Generate(ast *Query) string {
	return w.value
}

// -----

func fromIdxOfUntypedSlice(arr any, idx int) any {
	if arr == nil {
		return nil
	}
	v := arr.([]any)
	return v[idx]
}
