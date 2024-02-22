// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
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

package monitoring

import (
	"time"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/mquery-sru/result"
	"github.com/rs/zerolog/log"
)

type loadInfo struct {
	start       time.Time
	end         time.Time
	timeRunning float64
}

type WorkerJobLogger struct {
	location    *time.Location
	inputStream chan result.JobLog
	totals      *collections.CircularList[result.JobLog]
}

func (w *WorkerJobLogger) Log(rec result.JobLog) {
	w.inputStream <- rec
}

func (w *WorkerJobLogger) RunLoadInfoSummary() {
	ticker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ticker.C:
			// log summary
			var startTime, stopTime time.Time
			var totalRun time.Duration
			w.totals.ForEach(func(i int, item result.JobLog) bool {
				if i == 0 {
					startTime = item.Begin
				}
				stopTime = item.End
				totalRun += item.End.Sub(item.Begin)
				return true
			})
			measuredTime := stopTime.Sub(startTime).Seconds()
			log.Info().
				Float64("load", totalRun.Seconds()/measuredTime).
				Float64("totalRunSeconds", totalRun.Seconds()).
				Msg("reporting worker load")
		case v := <-w.inputStream:
			w.totals.Append(v)
		}
	}
}

func NewWorkerJobLogger(location *time.Location) *WorkerJobLogger {
	return &WorkerJobLogger{
		location:    location,
		inputStream: make(chan result.JobLog, 100),
		totals:      collections.NewCircularList[result.JobLog](60),
	}
}
