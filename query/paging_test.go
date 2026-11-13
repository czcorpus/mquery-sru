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

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// referenceMerge creates the whole merged result in a naive way
func referenceMerge(sizes []int) []Slot {
	ans := make([]Slot, 0)
	for round := 0; len(ans) < TotalSize(sizes); round++ {
		for rsc, s := range sizes {
			if s > round {
				ans = append(ans, Slot{Rsc: rsc, Line: round})
			}
		}
	}
	return ans
}

func TestPlanMergeDocExample(t *testing.T) {
	// A (2 lines), B (0 lines), C (4 lines) => A0 C0 A1 C1 C2 C3
	plan := PlanMerge([]int{2, 0, 4}, 0, 10)
	assert.Equal(
		t,
		[]Slot{{0, 0}, {2, 0}, {0, 1}, {2, 1}, {2, 2}, {2, 3}},
		plan.Slots,
	)
	assert.Equal(t, []RscRange{{0, 2}, {0, 0}, {0, 4}}, plan.Ranges)
}

func TestPlanMergeEmptyResourceDoesNotShortenPage(t *testing.T) {
	// formerly, an empty resource "consumed" its share of a page
	plan := PlanMerge([]int{4, 0}, 0, 4)
	assert.Equal(t, []Slot{{0, 0}, {0, 1}, {0, 2}, {0, 3}}, plan.Slots)
}

func TestPlanMergeUnevenSizesSecondPage(t *testing.T) {
	// A has 10 lines, B has 1000 lines; page size 100
	// page 1: A0 B0 ... A9 B9 B10 ... B89
	// page 2 must continue with B90 (formerly it started with B50)
	plan := PlanMerge([]int{10, 1000}, 100, 100)
	assert.Equal(t, Slot{Rsc: 1, Line: 90}, plan.Slots[0])
	assert.Equal(t, Slot{Rsc: 1, Line: 189}, plan.Slots[99])
	assert.Equal(t, []RscRange{{0, 0}, {90, 100}}, plan.Ranges)
}

func TestPlanMergeOffsetOutOfRange(t *testing.T) {
	plan := PlanMerge([]int{3, 2}, 5, 10)
	assert.Empty(t, plan.Slots)
	assert.Equal(t, []RscRange{{0, 0}, {0, 0}}, plan.Ranges)
}

func TestPlanMergeNoResults(t *testing.T) {
	plan := PlanMerge([]int{0, 0, 0}, 0, 10)
	assert.Empty(t, plan.Slots)
}

func TestPlanMergeNoResources(t *testing.T) {
	plan := PlanMerge([]int{}, 0, 10)
	assert.Empty(t, plan.Slots)
	assert.Empty(t, plan.Ranges)
}

func TestPlanMergeLastPartialPage(t *testing.T) {
	plan := PlanMerge([]int{3, 3}, 4, 10)
	assert.Equal(t, []Slot{{0, 2}, {1, 2}}, plan.Slots)
}

// TestPlanMergeMatchesReference tests many random setups against the naive
// implementation, including the fact that consecutive pages form the whole
// result without gaps and duplicates
func TestPlanMergeMatchesReference(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	for setup := 0; setup < 300; setup++ {
		sizes := make([]int, 1+rnd.Intn(12))
		for i := range sizes {
			switch rnd.Intn(4) {
			case 0:
				sizes[i] = 0
			case 1:
				sizes[i] = rnd.Intn(5)
			default:
				sizes[i] = rnd.Intn(80)
			}
		}
		expected := referenceMerge(sizes)
		total := len(expected)

		// arbitrary pages
		for i := 0; i < 20; i++ {
			offset := rnd.Intn(total + 5)
			limit := 1 + rnd.Intn(40)
			plan := PlanMerge(sizes, offset, limit)
			if offset >= total {
				assert.Empty(t, plan.Slots)
				continue
			}
			end := min(offset+limit, total)
			if !assert.Equal(t, expected[offset:end], plan.Slots, "sizes %v, offset %d, limit %d", sizes, offset, limit) {
				return
			}
			// ranges must exactly describe the slots
			for rsc, rng := range plan.Ranges {
				var lines []int
				for _, s := range plan.Slots {
					if s.Rsc == rsc {
						lines = append(lines, s.Line)
					}
				}
				assert.Equal(t, len(lines), rng.Count)
				for j, ln := range lines {
					assert.Equal(t, rng.From+j, ln)
				}
			}
		}

		// consecutive pages
		pageSize := 1 + rnd.Intn(15)
		joined := make([]Slot, 0, total)
		for offset := 0; offset < total; offset += pageSize {
			joined = append(joined, PlanMerge(sizes, offset, pageSize).Slots...)
		}
		assert.Equal(t, expected, joined)
	}
}

func TestEstimateRangesCoversPlanForLargeResults(t *testing.T) {
	sizes := []int{10000, 10000, 10000, 10000, 10000, 10000, 10000}
	for _, offset := range []int{0, 1, 6, 7, 38, 100, 1234} {
		for _, limit := range []int{1, 5, 10, 50} {
			est := EstimateRanges(len(sizes), offset, limit)
			plan := PlanMerge(sizes, offset, limit)
			for i := range sizes {
				assert.True(t, est[i].Contains(plan.Ranges[i]), "offset %d, limit %d, rsc %d", offset, limit, i)
			}
		}
	}
}

func TestEstimateRangesFirstPageAlwaysCovered(t *testing.T) {
	// on the first page, the estimation is always sufficient
	sizes := []int{0, 3, 1000, 1, 0, 25}
	est := EstimateRanges(len(sizes), 0, 20)
	plan := PlanMerge(sizes, 0, 20)
	for i := range sizes {
		assert.True(t, est[i].Contains(plan.Ranges[i]))
	}
}

func TestRscRangeContains(t *testing.T) {
	r := RscRange{From: 10, Count: 5}
	assert.True(t, r.Contains(RscRange{From: 10, Count: 5}))
	assert.True(t, r.Contains(RscRange{From: 12, Count: 2}))
	assert.True(t, r.Contains(RscRange{From: 100, Count: 0}))
	assert.False(t, r.Contains(RscRange{From: 9, Count: 2}))
	assert.False(t, r.Contains(RscRange{From: 14, Count: 2}))
}
