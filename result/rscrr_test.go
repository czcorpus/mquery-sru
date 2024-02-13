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

	"github.com/czcorpus/mquery-sru/corpus/conc"

	"github.com/stretchr/testify/assert"
)

func createResource() *RoundRobinLineSel {
	r := NewRoundRobinLineSel("corp1", "corp2", "corp3")
	r.SetRscLines("corp1", ConcExample{Lines: []conc.ConcordanceLine{
		{Text: conc.TokenSlice{&conc.Token{Word: "foo1"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "foo2"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "foo3"}}},
	}})
	r.SetRscLines("corp2", ConcExample{Lines: []conc.ConcordanceLine{
		{Text: conc.TokenSlice{&conc.Token{Word: "bar1"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "bar2"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "bar3"}}},
	}})
	r.SetRscLines("corp3", ConcExample{Lines: []conc.ConcordanceLine{
		{Text: conc.TokenSlice{&conc.Token{Word: "baz1"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "baz2"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "baz3"}}},
	}})
	return r
}

func createResourceWithSomeEmpty() *RoundRobinLineSel {
	r := NewRoundRobinLineSel("corp1", "corp2", "corp3")
	r.SetRscLines("corp1", ConcExample{Lines: []conc.ConcordanceLine{}})
	r.SetRscLines("corp2", ConcExample{Lines: []conc.ConcordanceLine{
		{Text: conc.TokenSlice{&conc.Token{Word: "bar1"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "bar2"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "bar3"}}},
	}})
	r.SetRscLines("corp3", ConcExample{Lines: []conc.ConcordanceLine{
		{Text: conc.TokenSlice{&conc.Token{Word: "baz1"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "baz2"}}},
		{Text: conc.TokenSlice{&conc.Token{Word: "baz3"}}},
	}})
	return r
}

func firstWord(line conc.ConcordanceLine) string {
	return line.Text[0].Word
}

func TestEmptyWithoutFactory(t *testing.T) {
	r := new(RoundRobinLineSel)
	hasNext := r.setNextAvailRsc()
	assert.False(t, hasNext)
}

func TestTypicalSetup(t *testing.T) {
	r := createResource()
	r.Next()
	assert.Equal(t, "foo1", firstWord(r.CurrLine()))

}

func TestAllDepletedWorks(t *testing.T) {
	r := createResource()
	for i := 0; i < 9; i++ {
		r.Next()
	}
	r.Next()
	assert.True(t, r.AllDepleted())
}

// TestWithSomeEmpty reflects problem reported
// in https://github.com/czcorpus/mquery-sru/issues/23
func TestWithSomeEmpty(t *testing.T) {
	r := createResourceWithSomeEmpty()
	hasNext := r.Next()
	assert.True(t, hasNext)
	ft := firstWord(r.items[r.currIdx].Lines.Lines[0])
	assert.Equal(t, "bar1", ft)
}

func TestSetRscLinesPanicsIfIterationStarted(t *testing.T) {
	r := createResource()
	r.Next()
	assert.Panics(t, func() {
		r.SetRscLines("corp1", ConcExample{Lines: []conc.ConcordanceLine{}})
	})
}
