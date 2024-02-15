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

type LineRange struct {
	Rsc  string
	From int
	To   int
}

type LineRangeList []LineRange

func (lrlist LineRangeList) Resources() []string {
	ans := make([]string, len(lrlist))
	for i, v := range lrlist {
		ans[i] = v.Rsc
	}
	return ans
}

// CalculatePartialRanges calculates ranges for individual resources (corpora)
// in case we know that:
//  1. we need global offset and limit
//  2. we create result by round robin selection from individual records.
//
// Please note that the offset is zero-based! Also, the order of resulting
// ranges may differ from the `rscList` so the iteration always starts from
// the correct resource. The order is changed only by rotating the resource list.
//
// E.g. for two corpora and offset 100, we set offset 50 for each corpus.
// But because we don't know whether each of the corpora will be able to provide
// enough records, we have to set the global limit for each individual resource so
// in case all but one corpora results are empty, we can still provide the required
// number of items.
func CalculatePartialRanges(rscList []string, offset, limit int) LineRangeList {

	numRsc := len(rscList)
	commonStart := offset / numRsc
	remaind := offset % numRsc
	ans := make([]LineRange, 0, len(rscList))
	for i := 0; i < len(rscList); i++ {
		ans = append(ans, LineRange{Rsc: rscList[i], From: commonStart, To: commonStart + limit})
	}
	for i := 0; i < remaind; i++ {
		ans[i].From++
		ans[i].To++
	}
	ans2 := make([]LineRange, 0, len(rscList))
	for i := 0; i < numRsc; i++ {
		ans2 = append(ans2, ans[(i+remaind)%numRsc])
	}
	return ans2
}
