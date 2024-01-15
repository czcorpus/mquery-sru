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

type query struct {
	mainQuery *mainQuery
	within    *withinPart
}

func (q *query) String() string {
	fmt.Println("> query.String()")
	if q.within != nil {
		return fmt.Sprintf("%s %s", q.mainQuery.String(), q.within.String())
	}
	return q.mainQuery.String()
}

// ----

type quantifiedQuery struct {
	simpleQuery *simpleQuery
	quantifier  string
}

func (qq *quantifiedQuery) String() string {
	fmt.Println("> quantifiedQuery.String()")
	if qq.quantifier != "" {
		return fmt.Sprintf("%s%s", qq.simpleQuery.String(), qq.quantifier)
	}
	return qq.simpleQuery.String()
}

// -----

type mainQuery struct {
	quantifiedQuery *quantifiedQuery
	mainQuery       *mainQuery
	operator        mainQueryOp
}

func (mq *mainQuery) String() string {
	fmt.Println("> mainQuery.String()")
	switch mq.operator {
	case mainQueryOpNone:
		return mq.quantifiedQuery.String()
	case mainQueryOpSequence:
		return fmt.Sprintf("%s %s", mq.quantifiedQuery.String(), mq.mainQuery.String())
	case mainQueryOpOr:
		return fmt.Sprintf("%s | %s", mq.quantifiedQuery.String(), mq.mainQuery.String())
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

func (be *basicExpression) String() string {
	fmt.Println("> basicExpression.String()")
	switch be.exprType {
	case basicExpressionTypeGroup:
		return fmt.Sprintf("(%s)", be.expression.String())
	case basicExpressionTypeNot:
		return fmt.Sprintf("!%s", be.expression.String())
	case basicExpressionTypeAttrOpRegexp:
		return fmt.Sprintf("%s%s%s", be.attribute.String(), be.operator, be.flaggedRegexp.String())
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

func (e *expression) String() string {
	fmt.Println("> expression.String()")
	if e == nil {
		return ""
	}
	var ans strings.Builder
	ans.WriteString(e.basicExpression.String())
	for _, te := range e.tailValues {
		ans.WriteString(fmt.Sprintf(" %s %s", te.operator, te.value.String()))
	}
	return ans.String()
}

// -------

type attribute struct {
	name  string
	value string
}

func (a *attribute) String() string {
	fmt.Println("> attribute.String()")
	if a.name != "" {
		return fmt.Sprintf("%s:%s", a.name, a.value)
	}
	return a.value
}

// -------

type regexp struct {
	quotedString *quotedString
}

func (r *regexp) WithPrefix(p string) string {
	return r.quotedString.WithPrefix(p)
}

func (r *regexp) String() string {
	return r.quotedString.String()
}

// -------

type flaggedRegexp struct {
	regexp *regexp
	flags  []string
}

func (fr *flaggedRegexp) String() string {
	// TODO add support for additional stuff besides case sensitivity
	fmt.Println("> flaggedRegexp.String()")
	var flag string
	for _, f := range fr.flags {
		if f == "i" || f == "I" || f == "c" || f == "C" {
			flag = "($i)"

		} else {
			log.Warn().Str("flag", flag).Msg("requested unsupported regexp flag")
		}
	}
	if flag != "" {
		return fr.regexp.WithPrefix(flag)
	}
	return fr.regexp.String()
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

func (wp *withinPart) String() string {
	fmt.Println("> withinPart.String()")
	return fmt.Sprintf("within <%s />", wp.value)
}

// ----

type implicitQuery struct {
	flaggedRegexp *flaggedRegexp
}

func (wp *implicitQuery) String() string {
	fmt.Println("> implicitQuery.String()")
	return wp.flaggedRegexp.String()
}

// ------

type segmentQuery struct {
	expression *expression
}

func (wp *segmentQuery) String() string {
	fmt.Println("> segmentQuery.String()")
	return fmt.Sprintf("[%s]", wp.expression.String())
}

// -------

type simpleQuery struct {
	value any
}

func (sq *simpleQuery) String() string {
	fmt.Println("> simpleQuery.String()")
	if sq.GetInnerQuery() != nil {
		return fmt.Sprintf("(%s)", sq.GetInnerQuery().String())

	} else if sq.GetImplicitQuery() != nil {
		return sq.GetImplicitQuery().String()

	} else if sq.GetSegmentQuery() != nil {
		return sq.GetSegmentQuery().String()
	}
	return "??"
}

func (sq *simpleQuery) GetInnerQuery() *mainQuery {
	v, ok := sq.value.(*mainQuery)
	if !ok {
		return nil
	}
	return v
}

func (sq *simpleQuery) GetImplicitQuery() *implicitQuery {
	v, ok := sq.value.(*implicitQuery)
	if !ok {
		return nil
	}
	return v
}

func (sq *simpleQuery) GetSegmentQuery() *segmentQuery {
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

func (qs *quotedString) String() string {
	fmt.Println("> quotedString.String()")
	if qs.regexp != "" {
		return fmt.Sprintf(`"%s"`, qs.regexp)
	}
	return fmt.Sprintf(`"%s"`, qs.value)
}

func (qs *quotedString) WithPrefix(p string) string {
	return fmt.Sprintf(p + qs.value)
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
