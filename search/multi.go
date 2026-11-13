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

// Package search provides searching in multiple resources (corpora)
// with results merged into a single result with stable paging.
package search

import (
	"errors"
	"fmt"

	"github.com/czcorpus/mquery-common/concordance"
	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/mango"
	"github.com/czcorpus/mquery-sru/query"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/czcorpus/mquery-sru/result"
	"github.com/rs/zerolog/log"
)

// QueryPublisher publishes queries for workers
// (typically implemented by rdb.Adapter)
type QueryPublisher interface {
	PublishQuery(query rdb.Query) (<-chan result.ConcResult, error)
}

// ResourceQuery describes a search in a single resource.
// The StartLine and MaxItems arguments are set by the search itself.
type ResourceQuery struct {
	ID   string
	PID  string
	Args rdb.ConcQueryArgs
}

// Line is a single line of a merged result
type Line struct {

	// ResourceIdx is an index of the resource (within the searched
	// resources) the line comes from
	ResourceIdx int

	Line concordance.Line
}

// Result is a single page of a merged multi-resource result
type Result struct {

	// Lines contains the requested page of the merged result
	Lines []Line

	// Total is the size of the merged result (failed resources
	// are not counted in)
	Total int

	// Failed contains resources which failed to provide their results.
	// Such resources are skipped.
	Failed []general.ResourceError
}

// AllFailed tells whether there was no resource able to provide its result
func (r *Result) AllFailed(numResources int) bool {
	return numResources > 0 && len(r.Failed) == numResources
}

type fetchedLines struct {
	rng   query.RscRange
	lines []concordance.Line
}

// get returns a line with the `idx` index within the whole resource result
func (fl fetchedLines) get(idx int) (concordance.Line, bool) {
	i := idx - fl.rng.From
	if i < 0 || i >= len(fl.lines) {
		return concordance.Line{}, false
	}
	return fl.lines[i], true
}

// publishAll publishes queries for resources with non-empty ranges.
// The returned slice contains nil for resources with no query.
func publishAll(
	publisher QueryPublisher,
	resources []ResourceQuery,
	ranges []query.RscRange,
) ([]<-chan result.ConcResult, error) {
	waits := make([]<-chan result.ConcResult, len(resources))
	for i, rsc := range resources {
		if ranges[i].Count == 0 {
			continue
		}
		args := rsc.Args
		args.StartLine = ranges[i].From
		args.MaxItems = ranges[i].Count
		wait, err := publisher.PublishQuery(rdb.Query{Func: "concExample", Args: args})
		if err != nil {
			return nil, fmt.Errorf("failed to publish query for %s: %w", rsc.ID, err)
		}
		waits[i] = wait
	}
	return waits, nil
}

// Search searches in all the `resources` and returns a page (specified by
// a zero-based `offset` and `limit`) of the merged result. The merged result
// is defined by query.PlanMerge.
//
// The search works in two phases. In the first one, all the resources are
// searched with estimated line ranges (query.EstimateRanges), which also
// provides the result sizes. With the sizes, the exact ranges are calculated
// and resources whose estimated lines do not cover the exact ones are searched
// again (this happens only for pages other than the first one and typically
// only for resources with results much larger than the others).
//
// Failures of individual resources are not fatal - they are reported via
// Result.Failed. The returned error means that the search could not be performed
// at all.
func Search(publisher QueryPublisher, resources []ResourceQuery, offset, limit int) (*Result, error) {
	ans := &Result{Failed: make([]general.ResourceError, 0)}
	failed := make([]bool, len(resources))
	setFailed := func(i int, err error) {
		log.Error().
			Err(err).
			Str("resource", resources[i].ID).
			Msg("failed to search resource, skipping it")
		failed[i] = true
		ans.Failed = append(
			ans.Failed,
			general.ResourceError{ID: resources[i].ID, PID: resources[i].PID, Err: err},
		)
	}

	// phase 1 - estimated ranges + result sizes
	estimated := query.EstimateRanges(len(resources), offset, limit)
	waits, err := publishAll(publisher, resources, estimated)
	if err != nil {
		return nil, err
	}
	sizes := make([]int, len(resources))
	fetched := make([]fetchedLines, len(resources))
	for i, wait := range waits {
		res := <-wait
		if res.Error != nil && !errors.Is(res.Error, mango.ErrRowsRangeOutOfConc) {
			setFailed(i, res.Error)
			continue
		}
		sizes[i] = res.ConcSize
		fetched[i] = fetchedLines{
			rng:   query.RscRange{From: estimated[i].From, Count: len(res.Lines)},
			lines: res.Lines,
		}
	}
	ans.Total = query.TotalSize(sizes)

	// phase 2 - exact ranges
	plan := query.PlanMerge(sizes, offset, limit)
	missing := make([]query.RscRange, len(resources))
	var numMissing int
	for i, rng := range plan.Ranges {
		if !fetched[i].rng.Contains(rng) {
			missing[i] = rng
			numMissing++
		}
	}
	if numMissing > 0 {
		log.Debug().
			Int("numResources", len(resources)).
			Int("numRefetched", numMissing).
			Int("offset", offset).
			Msg("some resources must be searched again to obtain the requested page")
		waits, err := publishAll(publisher, resources, missing)
		if err != nil {
			return nil, err
		}
		for i, wait := range waits {
			if wait == nil {
				continue
			}
			res := <-wait
			if res.Error != nil {
				setFailed(i, res.Error)
				continue
			}
			fetched[i] = fetchedLines{
				rng:   query.RscRange{From: missing[i].From, Count: len(res.Lines)},
				lines: res.Lines,
			}
		}
	}

	ans.Lines = make([]Line, 0, len(plan.Slots))
	for _, slot := range plan.Slots {
		if failed[slot.Rsc] {
			continue
		}
		line, ok := fetched[slot.Rsc].get(slot.Line)
		if !ok {
			// should not happen unless the resource changed between the phases
			log.Warn().
				Str("resource", resources[slot.Rsc].ID).
				Int("line", slot.Line).
				Msg("missing line in a resource result, skipping")
			continue
		}
		ans.Lines = append(ans.Lines, Line{ResourceIdx: slot.Rsc, Line: line})
	}
	return ans, nil
}
