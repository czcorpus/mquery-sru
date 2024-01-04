// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package v20

import "fcs/general"

type FCSResourceInfo struct {
	PID             string
	Title           string
	Description     string
	URI             string
	Languages       []string
	AvailableLayers string
}

type Segment struct {
	ID    string
	Start int
	End   int
}

type Token struct {
	Text    string
	Hit     bool
	Segment Segment
	Layers  map[string]string
}

type FCSSearchRow struct {
	Position   int
	PID        string
	Web        string
	Ref        string
	LayerAttrs []string
	Tokens     []Token
}

type FCSExplain struct {
	ServerName          string
	ServerPort          string
	Database            string
	DatabaseTitle       string
	DatabaseDescription string
	Layers              map[string]string
}

type FCSSearchRetrieve struct {
	QueryType string
	Results   []FCSSearchRow
}

type FCSResponse struct {
	General       general.FCSGeneralResponse
	RecordPacking string
	Operation     string

	Explain        FCSExplain
	Resources      []FCSResourceInfo
	SearchRetrieve FCSSearchRetrieve
}
