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
	"fcs/corpus"
	"fmt"
)

// ParseQuery parses FCS-QL and returns an abstract syntax
// tree which can be used to generate CQL.
func ParseQuery(
	q string,
	posAttrs []corpus.PosAttr,
	smapping corpus.StructureMapping,
) (*Query, error) {
	ans, err := Parse("query", []byte(q)) // Debug(true))
	if err != nil {
		return nil, err
	}
	tAns, ok := ans.(*Query)
	if !ok {
		return nil, fmt.Errorf("invalid AST type produced by parser")
	}
	tAns.
		SetStructureMapping(smapping).
		SetPosAttrs(posAttrs)
	return tAns, nil
}
