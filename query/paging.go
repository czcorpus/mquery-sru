// Copyright 2026 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2026 Institute of the Czech National Corpus,
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

// The merged result of a multi-resource search is defined as follows:
// lines are taken in "rounds" - in round t, each resource (in the order
// of the resource list) having more than t lines contributes its t-th line.
// Resources which run out of lines are simply skipped. E.g. for resources
// A (2 lines), B (0 lines) and C (4 lines), the merged result is:
//
//	A0 C0 A1 C1 C2 C3
//
// Because the definition depends only on the sizes of the individual
// results, any page (offset, limit) of the merged result can be mapped
// exactly to lines of individual resources once the sizes are known.

// Slot identifies a single line of a merged multi-resource result.
type Slot struct {

	// Rsc is an index of a resource
	Rsc int

	// Line is a line index within the resource's result
	Line int
}

// RscRange is a contiguous range of lines of a resource's result.
type RscRange struct {
	From  int
	Count int
}

// Contains tests whether the other range lies within the range.
// An empty other range is contained in any range.
func (r RscRange) Contains(other RscRange) bool {
	if other.Count == 0 {
		return true
	}
	return other.From >= r.From && other.From+other.Count <= r.From+r.Count
}

// MergePlan describes a single page of a merged multi-resource result.
type MergePlan struct {

	// Slots contains the page lines in their final order
	Slots []Slot

	// Ranges contains, for each resource, the lines the resource
	// has to provide for the page (Count == 0 means "nothing").
	Ranges []RscRange
}

// TotalSize returns the size of a merged result
func TotalSize(sizes []int) int {
	var ans int
	for _, s := range sizes {
		ans += s
	}
	return ans
}

// linesInRounds returns the number of lines in the first `numRounds`
// rounds of a merged result
func linesInRounds(sizes []int, numRounds int) int {
	var ans int
	for _, s := range sizes {
		ans += min(s, numRounds)
	}
	return ans
}

// PlanMerge calculates which lines of individual resources (with result
// sizes `sizes`) form the page of the merged result specified
// by a zero-based `offset` and `limit`.
func PlanMerge(sizes []int, offset, limit int) MergePlan {
	ans := MergePlan{Ranges: make([]RscRange, len(sizes))}
	total := TotalSize(sizes)
	if offset < 0 || limit <= 0 || offset >= total {
		return ans
	}
	numLines := min(limit, total-offset)
	ans.Slots = make([]Slot, 0, numLines)

	// find the round containing the offset - i.e. the largest `round`
	// such that the lines of all the preceding rounds fit into the offset
	var maxSize int
	for _, s := range sizes {
		maxSize = max(maxSize, s)
	}
	round, hi := 0, maxSize
	for round < hi {
		mid := (round + hi + 1) / 2
		if linesInRounds(sizes, mid) <= offset {
			round = mid

		} else {
			hi = mid - 1
		}
	}

	// within the round, skip lines up to the offset
	toSkip := offset - linesInRounds(sizes, round)
	rsc := 0
	for ; rsc < len(sizes); rsc++ {
		if sizes[rsc] > round {
			if toSkip == 0 {
				break
			}
			toSkip--
		}
	}

	// now collect the lines
	for len(ans.Slots) < numLines {
		if sizes[rsc] > round {
			ans.Slots = append(ans.Slots, Slot{Rsc: rsc, Line: round})
			if ans.Ranges[rsc].Count == 0 {
				ans.Ranges[rsc].From = round
			}
			ans.Ranges[rsc].Count++
		}
		rsc++
		if rsc == len(sizes) {
			rsc = 0
			round++
		}
	}
	return ans
}

// EstimateRanges provides line ranges for individual resources
// in case their result sizes are not known yet. The estimation
// expects the sizes to be large enough. In such case, the ranges
// will contain everything PlanMerge requires for the page.
// In other cases (some sizes are small), resources with larger
// results will likely need to provide lines from a higher offset.
func EstimateRanges(numResources, offset, limit int) []RscRange {
	ans := make([]RscRange, numResources)
	if numResources == 0 {
		return ans
	}
	round := offset / numResources
	remainder := offset % numResources
	for i := range ans {
		ans[i] = RscRange{From: round, Count: limit}
		if i < remainder {
			ans[i].From++
		}
	}
	return ans
}
