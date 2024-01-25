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

package results

import (
	"fmt"

	"github.com/czcorpus/mquery-sru/corpus/conc"
	"github.com/czcorpus/mquery-sru/mango"
)

type item struct {
	Name     string
	Depleted bool
	Started  bool
	CurrLine int
	Err      error
	Lines    ConcExample
}

type RoundRobinLineSel struct {
	items   []item
	currIdx int
}

func (r *RoundRobinLineSel) DescribeCurr() string {
	return fmt.Sprintf(
		"resource [%s][%d] (depleted: %t)",
		r.items[r.currIdx].Name,
		r.items[r.currIdx].CurrLine,
		r.items[r.currIdx].Depleted,
	)
}

func (r *RoundRobinLineSel) CurrLine() conc.ConcordanceLine {
	if !r.items[r.currIdx].Started {
		panic(fmt.Sprintf("iterator not initialized for %s", r.items[r.currIdx].Name))
	}
	if r.items[r.currIdx].Depleted {
		panic("accessing depleted resource")
	}
	return r.items[r.currIdx].Lines.Lines[r.items[r.currIdx].CurrLine]
}

func (r *RoundRobinLineSel) CurrRscName() string {
	if r.items[r.currIdx].Depleted {
		panic("accessing depleted resource")
	}
	return r.items[r.currIdx].Name
}

func (r *RoundRobinLineSel) SetRscLines(rsc string, c ConcExample) {
	for i, item := range r.items {
		if item.Name == rsc {
			item.Lines = c
			r.items[i] = item
			return
		}
	}
	panic("unknown resource")
}

func (r *RoundRobinLineSel) RscSetErrorAt(idx int, err error) {
	r.items[idx].Err = err
	r.items[idx].Depleted = true
}

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
func (r *RoundRobinLineSel) Next() bool {
	if len(r.items) == 0 || r.IsEmpty() {
		return false
	}
	if !r.items[r.currIdx].Started {
		r.items[r.currIdx].Started = true
		return true
	}

	foundNext := r.nextRsc()
	if !foundNext {
		return false
	}
	if r.items[r.currIdx].Started {
		r.items[r.currIdx].CurrLine++

	} else {
		r.items[r.currIdx].Started = true
	}
	if r.items[r.currIdx].CurrLine >= len(r.items[r.currIdx].Lines.Lines) {
		r.setCurrDepleted()
		r.Next()
	}
	return true
}

func (r *RoundRobinLineSel) initEmpty() {
	if r.items == nil {
		r.items = []item{}
	}
}

func (r *RoundRobinLineSel) nextRsc() bool {
	r.initEmpty()
	if r.AllDepleted() {
		return false
	}
	var avail bool
	for i := (r.currIdx + 1) % len(r.items); avail == false; i = ((i + 1) % len(r.items)) {
		avail = !r.items[i].Depleted
		if avail {
			r.currIdx = i
			break
		}
	}
	return true
}

func (r *RoundRobinLineSel) AllDepleted() bool {
	for _, v := range r.items {
		if v.Depleted == false {
			return false
		}
	}
	return true
}

func (r *RoundRobinLineSel) setCurrDepleted() {
	r.initEmpty()
	r.items[r.currIdx].Depleted = true

}

func NewRoundRobinLineSel(items ...string) *RoundRobinLineSel {
	ans := &RoundRobinLineSel{
		items: make([]item, len(items)),
	}
	for i, v := range items {
		ans.items[i] = item{Name: v}
	}
	return ans
}
