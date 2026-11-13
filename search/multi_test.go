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

package search

import (
	"errors"
	"fmt"
	"testing"

	"github.com/czcorpus/mquery-common/concordance"
	"github.com/czcorpus/mquery-sru/mango"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/czcorpus/mquery-sru/result"
	"github.com/stretchr/testify/assert"
)

// fakePublisher simulates workers searching in corpora
// of given sizes. Lines are identified by their Ref ("corpus:lineIdx").
type fakePublisher struct {
	sizes map[string]int

	// failing maps corpus to the call number (1-based) from which
	// the corpus starts to fail (1 = always)
	failing map[string]int

	calls map[string][]rdb.ConcQueryArgs
}

func (fp *fakePublisher) PublishQuery(q rdb.Query) (<-chan result.ConcResult, error) {
	if fp.calls == nil {
		fp.calls = make(map[string][]rdb.ConcQueryArgs)
	}
	corp := q.Args.CorpusPath
	fp.calls[corp] = append(fp.calls[corp], q.Args)
	ch := make(chan result.ConcResult, 1)
	if from, ok := fp.failing[corp]; ok && len(fp.calls[corp]) >= from {
		ch <- result.ConcResult{Error: errors.New("broken corpus")}
		return ch, nil
	}
	size := fp.sizes[corp]
	if q.Args.StartLine >= size && !(size == 0 && q.Args.StartLine == 0) {
		ch <- result.ConcResult{ConcSize: size, Error: mango.ErrRowsRangeOutOfConc}
		return ch, nil
	}
	ans := result.ConcResult{ConcSize: size, Query: q.Args.Query}
	for i := q.Args.StartLine; i < min(size, q.Args.StartLine+q.Args.MaxItems); i++ {
		ans.Lines = append(ans.Lines, concordance.Line{Ref: fmt.Sprintf("%s:%d", corp, i)})
	}
	ch <- ans
	return ch, nil
}

func (fp *fakePublisher) numCalls(corp string) int {
	return len(fp.calls[corp])
}

func resources(names ...string) []ResourceQuery {
	ans := make([]ResourceQuery, len(names))
	for i, n := range names {
		ans[i] = ResourceQuery{ID: n, PID: "pid/" + n, Args: rdb.ConcQueryArgs{CorpusPath: n}}
	}
	return ans
}

func refs(res *Result) []string {
	ans := make([]string, len(res.Lines))
	for i, ln := range res.Lines {
		ans[i] = ln.Line.Ref
	}
	return ans
}

func TestFirstPageNeedsSinglePhase(t *testing.T) {
	fp := &fakePublisher{sizes: map[string]int{"a": 2, "b": 0, "c": 50}}
	res, err := Search(fp, resources("a", "b", "c"), 0, 5)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a:0", "c:0", "a:1", "c:1", "c:2"}, refs(res))
	assert.Equal(t, 52, res.Total)
	assert.Empty(t, res.Failed)
	assert.Equal(t, 1, fp.numCalls("a"))
	assert.Equal(t, 1, fp.numCalls("b"))
	assert.Equal(t, 1, fp.numCalls("c"))
}

func TestUnevenSizesSecondPageRefetches(t *testing.T) {
	// estimated start for both corpora is 50 (=100/2) but the exact
	// page 2 consists of "big" lines 90..189
	fp := &fakePublisher{sizes: map[string]int{"small": 10, "big": 1000}}
	res, err := Search(fp, resources("small", "big"), 100, 100)
	assert.NoError(t, err)
	assert.Len(t, res.Lines, 100)
	assert.Equal(t, "big:90", res.Lines[0].Line.Ref)
	assert.Equal(t, "big:189", res.Lines[99].Line.Ref)
	assert.Equal(t, 1010, res.Total)
	// the small corpus is not searched again, the big one is
	assert.Equal(t, 1, fp.numCalls("small"))
	assert.Equal(t, 2, fp.numCalls("big"))
	assert.Equal(t, 90, fp.calls["big"][1].StartLine)
	assert.Equal(t, 100, fp.calls["big"][1].MaxItems)
}

func TestConsecutivePagesFormWholeResult(t *testing.T) {
	sizes := map[string]int{"a": 7, "b": 0, "c": 33, "d": 1, "e": 120}
	names := []string{"a", "b", "c", "d", "e"}
	var total int
	for _, s := range sizes {
		total += s
	}
	seen := make(map[string]bool)
	for offset := 0; offset < total; offset += 10 {
		res, err := Search(&fakePublisher{sizes: sizes}, resources(names...), offset, 10)
		assert.NoError(t, err)
		assert.Equal(t, total, res.Total)
		assert.Len(t, res.Lines, min(10, total-offset))
		for _, r := range refs(res) {
			assert.False(t, seen[r], "duplicate line %s at offset %d", r, offset)
			seen[r] = true
		}
	}
	assert.Len(t, seen, total)
}

func TestFailedResourceIsSkipped(t *testing.T) {
	fp := &fakePublisher{
		sizes:   map[string]int{"a": 3, "broken": 100, "c": 3},
		failing: map[string]int{"broken": 1},
	}
	res, err := Search(fp, resources("a", "broken", "c"), 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a:0", "c:0", "a:1", "c:1", "a:2", "c:2"}, refs(res))
	assert.Equal(t, 6, res.Total)
	assert.Len(t, res.Failed, 1)
	assert.Equal(t, "pid/broken", res.Failed[0].PID)
	assert.False(t, res.AllFailed(3))
}

func TestFailureInSecondPhase(t *testing.T) {
	fp := &fakePublisher{
		sizes:   map[string]int{"small": 10, "big": 1000},
		failing: map[string]int{"big": 2},
	}
	res, err := Search(fp, resources("small", "big"), 100, 100)
	assert.NoError(t, err)
	assert.Empty(t, res.Lines)
	assert.Len(t, res.Failed, 1)
	assert.Equal(t, "big", res.Failed[0].ID)
}

func TestAllFailed(t *testing.T) {
	fp := &fakePublisher{
		sizes:   map[string]int{"a": 3, "b": 3},
		failing: map[string]int{"a": 1, "b": 1},
	}
	res, err := Search(fp, resources("a", "b"), 0, 10)
	assert.NoError(t, err)
	assert.True(t, res.AllFailed(2))
	assert.Empty(t, res.Lines)
}

func TestOffsetBeyondResult(t *testing.T) {
	fp := &fakePublisher{sizes: map[string]int{"a": 3, "b": 4}}
	res, err := Search(fp, resources("a", "b"), 50, 10)
	assert.NoError(t, err)
	assert.Empty(t, res.Lines)
	assert.Equal(t, 7, res.Total)
	assert.Empty(t, res.Failed)
}
