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

package v12

import (
	"fcs/corpus"
	"fcs/general"
)

type FCSResourceInfo struct {
	PID         string
	Title       map[string]string
	Description map[string]string
	URI         string
	Languages   []string
}

type Token struct {
	Text string
	Hit  bool
}

type FCSSearchRow struct {
	Position int
	PID      string
	Web      string
	Ref      string
	Tokens   []Token
}

type FCSExplain struct {
	ServerName          string
	ServerPort          string
	Database            string
	DatabaseTitle       map[string]string
	DatabaseDescription map[string]string
	PrimaryLanguage     string
	PosAttrs            []corpus.PosAttr
	Resources           []FCSResourceInfo
	ExtraResponseData   bool
}

type FCSSearchRetrieve struct {
	Results []FCSSearchRow
}

type FCSResponse struct {
	General       general.FCSGeneralResponse
	RecordPacking RecordPacking
	Operation     Operation

	Explain        FCSExplain
	SearchRetrieve FCSSearchRetrieve
}
