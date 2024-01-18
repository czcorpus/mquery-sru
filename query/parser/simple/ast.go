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
	"errors"
	"fcs/corpus"
	"fcs/query/compiler"
	"fmt"
	"strings"
)

type Query struct {
	binaryOperatorQuery *binaryOperatorQuery
	cqlAttrs            []string
	structureMapping    corpus.StructureMapping
	posAttrs            []corpus.PosAttr
	errors              []error
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

func (boq *binaryOperatorQuery) Generate(ast compiler.AST) string {
	var rest strings.Builder
	for _, v := range boq.rest {
		rest.WriteString(" " + v.nonRecursiveQuery.Generate(ast))
	}
	return fmt.Sprintf(
		"%s %s",
		boq.nonRecursiveQuery.Generate(ast),
		rest.String(),
	)
}

// ----

type nonRecursiveQuery struct {
	parenthesisExpr     *parenthesisExpr
	unaryOperator       string
	binaryOperatorQuery *binaryOperatorQuery
	term                *term
}

func (nrq *nonRecursiveQuery) Generate(ast compiler.AST) string {
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

func (pe *parenthesisExpr) Generate(ast compiler.AST) string {
	return fmt.Sprintf("(%s)", pe.binaryOperatorQuery.Generate(ast))
}

// ---

type term struct {
	text       *text
	quotedText *quotedText
}

func (t *term) Generate(ast compiler.AST) string {
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

func (qt *quotedText) Generate(ast compiler.AST) string {
	var ans strings.Builder
	for _, v := range qt.words {
		ans.WriteString(" " + v.Generate(ast))
	}
	return fmt.Sprintf(`"%s"`, ans.String())
}

func (qt *quotedText) AddWord(w *word) {
	qt.words = append(qt.words, w)
}

// -----

type text struct {
	word *word
}

func (t *text) Generate(ast compiler.AST) string {
	return t.word.Generate(ast)
}

// ------

type word struct {
	value string
}

func (w *word) Generate(ast compiler.AST) string {
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
