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

package query

type lineRange struct {
	From int
	To   int
}

// CalculatePartialRanges calculates ranges for individual resources (corpora)
// in case we know that:
//  1. we need global offset and limit
//  2. we create result by round robin selection from individual records.
//
// So e.g. for two corpora and offset 100, we set offset 50 for each corpus.
// But because we don't know whether each of the corpora will be able to provide
// enough records, we have to set the global limit for each individual resource so
// in case all but one corpora results are empty, we can still provide the required
// number of items.
func CalculatePartialRanges(rscList []string, offset, limit int) map[string]lineRange {

	numRsc := len(rscList)
	commonStart := offset / numRsc
	remaind := offset % numRsc
	ans := make(map[string]lineRange)
	for i := 0; i < len(rscList); i++ {
		ans[rscList[i]] = lineRange{commonStart, commonStart + limit}
	}
	for i := 0; i < int(remaind); i++ {
		v := ans[rscList[i]]
		v.From++
		v.To++
		ans[rscList[i]] = v
	}
	return ans
}
