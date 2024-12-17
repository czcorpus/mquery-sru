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

package result

import (
	"testing"

	"github.com/czcorpus/mquery-common/concordance"
	"github.com/stretchr/testify/assert"
)

func createSingleResourceEmptyResult() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(3, "corp1")
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{}})
	return r
}

func createSingleResource() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(4, "corp1")
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo3"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo4"}}},
	}})
	return r
}

func createTwoResourcesOneEmpty() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(4, "corp1", "corp2")
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo3"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo4"}}},
	}})
	r.SetRscLines("corp2", ConcResult{Lines: []concordance.Line{}})
	return r
}

func createTwoResourcesSecondSmaller() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(8, "corp1", "corp2") // 8 = "we expect 8 (but we get less)"
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo3"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo4"}}},
	}})
	r.SetRscLines("corp2", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar1"}}},
	}})
	return r
}

func createTwoResourcesFirstSmaller() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(8, "corp1", "corp2") // 8 = "we expect 8 (but we get less)"
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo1"}}},
	}})
	r.SetRscLines("corp2", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar3"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar4"}}},
	}})
	return r
}

func createResource() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(9, "corp1", "corp2", "corp3")
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "foo3"}}},
	}})
	r.SetRscLines("corp2", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar3"}}},
	}})
	r.SetRscLines("corp3", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "baz1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "baz2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "baz3"}}},
	}})
	return r
}

func createResourceWithSomeEmpty() *RoundRobinLineSel {
	r := NewRoundRobinLineSel(9, "corp1", "corp2", "corp3")
	r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{}})
	r.SetRscLines("corp2", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "bar3"}}},
	}})
	r.SetRscLines("corp3", ConcResult{Lines: []concordance.Line{
		{Text: concordance.TokenSlice{&concordance.Token{Word: "baz1"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "baz2"}}},
		{Text: concordance.TokenSlice{&concordance.Token{Word: "baz3"}}},
	}})
	return r
}

func firstWord(line *concordance.Line) string {
	if v, ok := line.Text[0].(*concordance.Token); ok {
		return v.Word
	}
	return ""
}

func TestTypicalSetup(t *testing.T) {
	r := createResource()
	r.Next()
	assert.Equal(t, "foo1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	r.Next()
	assert.Equal(t, "bar1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	r.Next()
	assert.Equal(t, "baz1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp3", r.CurrRscName())
	r.Next()
	assert.Equal(t, "foo2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	r.Next()
	assert.Equal(t, "bar2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	r.Next()
	assert.Equal(t, "baz2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp3", r.CurrRscName())
	r.Next()
	assert.Equal(t, "foo3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	r.Next()
	assert.Equal(t, "bar3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	r.Next()
	assert.Equal(t, "baz3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp3", r.CurrRscName())
	assert.Equal(t, 9, r.nextOutputLineIdx)
	assert.False(t, r.Next())
	assert.Nil(t, r.CurrLine())
	assert.Equal(t, 10, r.nextOutputLineIdx)
}

// TestWithSomeEmpty reflects problem reported
// in https://github.com/czcorpus/mquery-sru/issues/23
func TestWithSomeEmpty(t *testing.T) {
	r := createResourceWithSomeEmpty()
	hasNext := r.Next()
	assert.True(t, hasNext)
	assert.Equal(t, "bar1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	r.Next()
	assert.Equal(t, "baz1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp3", r.CurrRscName())
	r.Next()
	assert.Equal(t, "bar2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	r.Next()
	assert.Equal(t, "baz2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp3", r.CurrRscName())
	r.Next()
	assert.Equal(t, "bar3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	r.Next()
	assert.Equal(t, "baz3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp3", r.CurrRscName())
	assert.Equal(t, 9, r.nextOutputLineIdx)
	assert.False(t, r.Next())
	assert.Nil(t, r.CurrLine())
}

func TestSetRscLinesPanicsIfIterationStarted(t *testing.T) {
	r := createResource()
	r.Next()
	assert.Panics(t, func() {
		r.SetRscLines("corp1", ConcResult{Lines: []concordance.Line{}})
	})
}

func TestWorksWithSingleResource(t *testing.T) {
	r := createSingleResource()
	assert.True(t, r.Next())
	assert.Equal(t, "foo1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo4", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.False(t, r.Next())
}

// TestTwoResourcesOneEmpty is a regression test
func TestTwoResourcesOneEmpty(t *testing.T) {
	r := createTwoResourcesOneEmpty()
	assert.True(t, r.Next())
	assert.Equal(t, "foo1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.False(t, r.Next())
	assert.Nil(t, r.CurrLine())
}

// TestTwoResourcesSecondSmaller is a regression test
func TestTwoResourcesSecondSmaller(t *testing.T) {
	r := createTwoResourcesSecondSmaller()
	assert.True(t, r.Next())
	assert.Equal(t, "foo1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "bar1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "foo4", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.False(t, r.Next())
	assert.Nil(t, r.CurrLine())
}

func TestTwoResourcesFirstSmaller(t *testing.T) {
	r := createTwoResourcesFirstSmaller()
	assert.True(t, r.Next())
	assert.Equal(t, "foo1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp1", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "bar1", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "bar2", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "bar3", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	assert.True(t, r.Next())
	assert.Equal(t, "bar4", firstWord(r.CurrLine()))
	assert.Equal(t, "corp2", r.CurrRscName())
	assert.False(t, r.Next())
	assert.Nil(t, r.CurrLine())
}

func TestSingleResourceWithNoLines(t *testing.T) {
	r := createSingleResourceEmptyResult()
	assert.False(t, r.Next())
}
