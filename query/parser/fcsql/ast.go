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

package fcsql

import (
	"fmt"
	"strings"

	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/query/compiler"

	"github.com/rs/zerolog/log"
)

const (
	mainQueryOpNone mainQueryOp = iota
	mainQueryOpSequence
	mainQueryOpOr

	basicExpressionTypeGroup beType = iota
	basicExpressionTypeNot
	basicExpressionTypeAttrOpRegexp
)

type mainQueryOp int

type beType int

// ----

type Query struct {
	mainQuery        *mainQuery
	within           *withinPart
	structureMapping corpus.StructureMapping
	posAttrs         []corpus.PosAttr
	errors           []error
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

// TranslatePosAttr transforms a FCS-QL attribute specifier (e.g. `text`, `p_tag:pos`)
// into a real corpus positional attribute.
// Please note that it also supports `word` alias for the `text` layer
func (q *Query) TranslatePosAttr(qualifier, name string) string {
	if qualifier != "" {
		for _, p := range q.posAttrs {
			if p.Name == qualifier && (string(p.Layer) == name || p.Layer == "text" && name == "word") {
				return p.Name
			}
		}

	} else {
		for _, p := range q.posAttrs {
			if (string(p.Layer) == name || p.Layer == "text" && name == "word") && p.IsLayerDefault {
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
	q.errors = make([]error, 0, 20)
	if q.within != nil {
		return fmt.Sprintf(
			"%s %s",
			q.mainQuery.Generate(q),
			q.within.Generate(q),
		)
	}
	return q.mainQuery.Generate(q)
}

// ----

type quantifiedQuery struct {
	basicQuery *basicQuery
	quantifier string
}

func (qq *quantifiedQuery) Generate(ast compiler.AST) string {
	if qq.quantifier != "" {
		return fmt.Sprintf("%s%s", qq.basicQuery.Generate(ast), qq.quantifier)
	}
	return qq.basicQuery.Generate(ast)
}

// -----

type mainQuery struct {
	quantifiedQuery *quantifiedQuery
	mainQuery       *mainQuery
	operator        mainQueryOp
}

func (mq *mainQuery) Generate(ast compiler.AST) string {
	switch mq.operator {
	case mainQueryOpNone:
		return mq.quantifiedQuery.Generate(ast)
	case mainQueryOpSequence:
		return fmt.Sprintf(
			"%s %s", mq.quantifiedQuery.Generate(ast), mq.mainQuery.Generate(ast))
	case mainQueryOpOr:
		return fmt.Sprintf(
			"%s | %s", mq.quantifiedQuery.Generate(ast), mq.mainQuery.Generate(ast))
	default:
		return "??"
	}
}

// -------

type basicExpression struct {
	attribute     *attribute
	operator      string
	expression    *expression
	flaggedRegexp *flaggedRegexp
	exprType      beType
}

func (be *basicExpression) Generate(ast compiler.AST) string {
	switch be.exprType {
	case basicExpressionTypeGroup:
		return fmt.Sprintf("(%s)", be.expression.Generate(ast))
	case basicExpressionTypeNot:
		return fmt.Sprintf("!%s", be.expression.Generate(ast))
	case basicExpressionTypeAttrOpRegexp:
		return fmt.Sprintf(
			"%s%s%s", be.attribute.Generate(ast), be.operator, be.flaggedRegexp.Generate(ast))
	default:
		return "??"
	}
}

// ------

type expressionTailItem struct {
	operator string
	value    *basicExpression
}

type expression struct {
	basicExpression *basicExpression
	tailValues      []*expressionTailItem
}

func (e *expression) AddTailItem(operator string, value *basicExpression) {
	e.tailValues = append(
		e.tailValues,
		&expressionTailItem{operator: operator, value: value},
	)
}

func (e *expression) Generate(ast compiler.AST) string {
	if e == nil {
		return ""
	}
	var ans strings.Builder
	ans.WriteString(e.basicExpression.Generate(ast))
	for _, te := range e.tailValues {
		ans.WriteString(fmt.Sprintf(" %s %s", te.operator, te.value.Generate(ast)))
	}
	return ans.String()
}

// -------

type attribute struct {
	name  string
	value string
}

func (a *attribute) Generate(ast compiler.AST) string {
	return ast.TranslatePosAttr(a.name, a.value)
}

// -------

type regexp struct {
	quotedString *quotedString
}

func (r *regexp) WithPrefix(p string) string {
	return r.quotedString.WithPrefix(p)
}

func (r *regexp) Generate(ast compiler.AST) string {
	return r.quotedString.Generate(ast)
}

// -------

type flaggedRegexp struct {
	regexp *regexp
	flags  []string
}

func (fr *flaggedRegexp) Generate(ast compiler.AST) string {
	// TODO add support for additional stuff besides case sensitivity
	var flag string
	for _, f := range fr.flags {
		if f == "i" || f == "I" || f == "c" || f == "C" {
			flag = "(?i)"

		} else {
			log.Warn().Str("flag", flag).Msg("requested unsupported regexp flag")
		}
	}
	if flag != "" {
		return fr.regexp.WithPrefix(flag)
	}
	return fr.regexp.Generate(ast)
}

func (fr *flaggedRegexp) AttachUntypedFlag(v any) error {
	vt, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid value for flaggedRegexp flag")
	}
	fr.flags = append(fr.flags, vt)
	return nil
}

// ----

type withinPart struct {
	value string
}

func (wp *withinPart) Generate(ast compiler.AST) string {
	return fmt.Sprintf("within <%s />", ast.TranslateWithinCtx(wp.value))
}

// ----

type implicitQuery struct {
	flaggedRegexp *flaggedRegexp
}

func (wp *implicitQuery) Generate(ast compiler.AST) string {
	return wp.flaggedRegexp.Generate(ast)
}

// ------

type segmentQuery struct {
	expression *expression
}

func (wp *segmentQuery) Generate(ast compiler.AST) string {
	return fmt.Sprintf("[%s]", wp.expression.Generate(ast))
}

// -------

type basicQuery struct {
	value any
}

func (sq *basicQuery) Generate(ast compiler.AST) string {
	if sq.GetInnerQuery() != nil {
		return fmt.Sprintf("(%s)", sq.GetInnerQuery().Generate(ast))

	} else if sq.GetImplicitQuery() != nil {
		return sq.GetImplicitQuery().Generate(ast)

	} else if sq.GetSegmentQuery() != nil {
		return sq.GetSegmentQuery().Generate(ast)
	}
	return "??"
}

func (sq *basicQuery) GetInnerQuery() *mainQuery {
	v, ok := sq.value.(*mainQuery)
	if !ok {
		return nil
	}
	return v
}

func (sq *basicQuery) GetImplicitQuery() *implicitQuery {
	v, ok := sq.value.(*implicitQuery)
	if !ok {
		return nil
	}
	return v
}

func (sq *basicQuery) GetSegmentQuery() *segmentQuery {
	v, ok := sq.value.(*segmentQuery)
	if !ok {
		return nil
	}
	return v
}

// -----

type quotedString struct {
	value  string
	regexp string
}

func (qs *quotedString) Generate(ast compiler.AST) string {
	if qs.regexp != "" {
		return fmt.Sprintf(`"%s"`, qs.regexp)
	}
	return fmt.Sprintf(`"%s"`, qs.value)
}

func (qs *quotedString) WithPrefix(p string) string {
	return fmt.Sprintf(`"%s%s"`, p, qs.value)
}

func (qs *quotedString) Append(s string) {
	qs.value = qs.value + s
}

// -----

func fromIdxOfUntypedSlice(arr any, idx int) any {
	if arr == nil {
		return nil
	}
	v := arr.([]any)
	return v[idx]
}
