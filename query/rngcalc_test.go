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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreeResourcesOfDiffSizes(t *testing.T) {
	ans := CalculatePartialRanges([]string{"c1", "c2", "c3"}, 38, 10)
	assert.Equal(t, 12, ans[0].From)
	assert.Equal(t, 22, ans[0].To)
	assert.Equal(t, "c3", ans[0].Rsc)
	assert.Equal(t, 13, ans[1].From)
	assert.Equal(t, 23, ans[1].To)
	assert.Equal(t, "c1", ans[1].Rsc)
	assert.Equal(t, 13, ans[2].From)
	assert.Equal(t, 23, ans[2].To)
	assert.Equal(t, "c2", ans[2].Rsc)
}

func TestSingleResource(t *testing.T) {
	ans := CalculatePartialRanges([]string{"c1"}, 38, 10)
	assert.Equal(t, 38, ans[0].From)
	assert.Equal(t, 48, ans[0].To)
}
