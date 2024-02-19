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
	"fmt"

	"github.com/czcorpus/mquery-sru/corpus/conc"
	"github.com/czcorpus/mquery-sru/mango"
)

type item struct {
	Name     string
	CurrLine int
	Err      error
	Lines    ConcExample
	Started  bool
}

// RoundRobinLineSel allows for fetching data from
// multiple search results (= from different corpora)
// and taking them by "round robin" style. It is able
// to handle result sets of different sizes by cycling
// through less and less resources as they run out of items.
type RoundRobinLineSel struct {
	items       []item
	currIdx     int
	maxLines    int
	lineCounter int
}

func (r *RoundRobinLineSel) DescribeCurr() string {
	return fmt.Sprintf(
		"resource [%s][%d]",
		r.items[r.currIdx].Name,
		r.items[r.currIdx].CurrLine,
	)
}

// CurrLine returns the current line from a current resource
// during an iteration. It is intended to be called within a loop
// controlled by method `Next()`
func (r *RoundRobinLineSel) CurrLine() *conc.ConcordanceLine {
	if r.lineCounter >= r.maxLines+1 { // lineCounter is always ahead by 1 (that's why `+1`)
		return nil
	}
	lineNum := r.items[r.currIdx].CurrLine
	if lineNum < len(r.items[r.currIdx].Lines.Lines) {
		return &r.items[r.currIdx].Lines.Lines[lineNum]
	}
	return nil
}

// CurrRscName returns the currently set resource (corpus)
// during iteration.
func (r *RoundRobinLineSel) CurrRscName() string {
	return r.items[r.currIdx].Name
}

func (r *RoundRobinLineSel) iterationStarted() bool {
	for _, v := range r.items {
		if v.Started {
			return true
		}
	}
	return false
}

// SetRscLines sets concordance data for a resource (corpus).
// The method can be called only if the `Next()` method has not
// been called yet. Otherwise the call panics.
func (r *RoundRobinLineSel) SetRscLines(rsc string, c ConcExample) {
	if r.iterationStarted() {
		panic("cannot add resource lines to an already iterating RoundRobinLineSel")
	}
	for i, item := range r.items {
		if item.Name == rsc {
			item.Lines = c
			r.items[i] = item
			return
		}
	}
	panic("unknown resource")
}

// RscSetErrorAt sets and error for idx-th resource. With that,
// the iteration may continue, but the errored resource is skipped.
func (r *RoundRobinLineSel) RscSetErrorAt(idx int, err error) {
	r.items[idx].Err = err
}

// CurrRscGetError returns possible error for the current resource.
func (r *RoundRobinLineSel) CurrRscGetError() error {
	return r.items[r.currIdx].Err
}

// HasFatalError means that each configured resource (corpus)
// has an error and thus there is no source we can load
// lines from.
func (r *RoundRobinLineSel) HasFatalError() bool {
	for _, v := range r.items {
		if v.Err == nil {
			return false
		}
	}
	return true
}

// AllHasOutOfRangeError means that there was not a single
// resource able to provide lines with a required offset
func (r *RoundRobinLineSel) AllHasOutOfRangeError() bool {
	var numMatch int
	for _, v := range r.items {
		if v.Err != nil && v.Err.Error() == mango.ErrRowsRangeOutOfConc.Error() {
			numMatch++
		}
	}
	return numMatch == len(r.items)
}

func (r *RoundRobinLineSel) GetFirstError() error {
	for _, v := range r.items {
		if v.Err != nil {
			return v.Err
		}
	}
	return nil
}

// IsEmpty returns true if all the resources are emtpy
func (r *RoundRobinLineSel) IsEmpty() bool {
	for _, v := range r.items {
		if len(v.Lines.Lines) > 0 {
			return false
		}
	}
	return true
}

// Next prepares next line from the multi-resource result.
// Please note that to obtain the first item Next() must be
// called too.
// Also, once called for the first time, no new result sets
// can be added (this causes the call to panic)
func (r *RoundRobinLineSel) Next() bool {
	if r.lineCounter == r.maxLines {
		return false
	}
	if len(r.items) == 0 || r.IsEmpty() {
		return false
	}

	if !r.items[r.currIdx].Started {
		r.lineCounter++
		r.items[r.currIdx].Started = true
		if len(r.items[r.currIdx].Lines.Lines) > 0 {
			return true
		}
	}

	for i := 0; i < len(r.items); i++ {
		r.lineCounter++
		r.currIdx = (r.currIdx + 1) % len(r.items)
		if !r.items[r.currIdx].Started {
			r.items[r.currIdx].Started = true
			if len(r.items[r.currIdx].Lines.Lines) > 0 {
				return true
			}
			continue
		}
		r.items[r.currIdx].CurrLine++
		if r.items[r.currIdx].CurrLine < len(r.items[r.currIdx].Lines.Lines) {
			return true
		}
	}
	return false
}

// NewRoundRobinLineSel creates a new instance of NewRoundRobinLineSel
// with correctly initialized attributes.
func NewRoundRobinLineSel(maxLines int, items ...string) *RoundRobinLineSel {
	ans := &RoundRobinLineSel{
		items:    make([]item, len(items)),
		maxLines: maxLines,
	}
	for i, v := range items {
		ans.items[i] = item{Name: v}
	}
	return ans
}
