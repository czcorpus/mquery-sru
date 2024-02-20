// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
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

package result

import (
	"errors"

	"github.com/czcorpus/mquery-sru/corpus/conc"
)

const (
	ResultTypeFx           = "Fx"
	ResultTypeFy           = "Fy"
	ResultTypeFxy          = "Fxy"
	ResultTypeCollocations = "Collocations"
	ResultTypeCollFreqData = "collFreqData"
	ResultTypeError        = "Error"
)

type ResultType string

func (rt ResultType) IsValid() bool {
	return rt == ResultTypeFx || rt == ResultTypeFy || rt == ResultTypeFxy
}

func (rt ResultType) String() string {
	return string(rt)
}

type SerializableResult interface {
	Type() ResultType
	Err() error
}

// ----

type ConcExample struct {
	Lines      []conc.ConcordanceLine `json:"lines"`
	ConcSize   int                    `json:"concSize"`
	ResultType ResultType             `json:"resultType"`
	Error      string                 `json:"error"`
}

func (res *ConcExample) Err() error {
	if res.Error != "" {
		return errors.New(res.Error)
	}
	return nil
}

func (res *ConcExample) Type() ResultType {
	return res.ResultType
}

func (res *ConcExample) NumLines() int {
	return len(res.Lines)
}
