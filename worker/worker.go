// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
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

package worker

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/czcorpus/mquery-common/concordance"
	"github.com/czcorpus/mquery-sru/mango"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/czcorpus/mquery-sru/result"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	DefaultTickerInterval = 2 * time.Second
	MaxFreqResultItems    = 100
)

type jobLogger interface {
	Log(rec result.JobLog)
}

type Worker struct {
	ID         string
	messages   <-chan *redis.Message
	radapter   *rdb.Adapter
	ctx        context.Context
	ticker     *time.Ticker
	jobLogger  jobLogger
	currJobLog *result.JobLog
}

func (w *Worker) publishResult(res *result.ConcResult, channel string) error {
	w.currJobLog.End = time.Now()
	w.currJobLog.Err = res.Error
	w.jobLogger.Log(*w.currJobLog)
	w.currJobLog = nil
	return w.radapter.PublishResult(channel, res)
}

func (w *Worker) tryNextQuery() error {
	time.Sleep(time.Duration(rand.Intn(40)) * time.Millisecond)
	query, err := w.radapter.DequeueQuery()
	if err == rdb.ErrorEmptyQueue {
		return nil

	} else if err != nil {
		return err
	}
	log.Debug().
		Str("channel", query.Channel).
		Str("func", query.Func).
		Any("args", query.Args).
		Msg("received query")

	isActive, err := w.radapter.SomeoneListens(query)
	if err != nil {
		return err
	}
	if !isActive {
		log.Warn().
			Str("func", query.Func).
			Str("channel", query.Channel).
			Any("args", query.Args).
			Msg("worker found an inactive query")
		return nil
	}

	w.currJobLog = &result.JobLog{
		WorkerID: w.ID,
		Func:     query.Func,
		Begin:    time.Now(),
	}
	ans := w.ConcResult(query.Args)
	if err := w.publishResult(ans, query.Channel); err != nil {
		return fmt.Errorf("failed to publish result: %w", err)
	}
	return nil
}

func (w *Worker) Listen() {
	for {
		select {
		case <-w.ticker.C:
			if err := w.tryNextQuery(); err != nil {
				log.Error().
					Err(err).
					Msg("failed to process query")
			}
		case <-w.ctx.Done():
			w.ticker.Stop()
			log.Info().Msg("worker exiting due to cancellation")
			return
		case msg := <-w.messages:
			if msg.Payload == rdb.MsgNewQuery {
				if err := w.tryNextQuery(); err != nil {
					log.Error().
						Err(err).
						Msg("failed to process query")
				}
			}
		}
	}
}

func (w *Worker) ConcResult(args rdb.ConcQueryArgs) (ans *result.ConcResult) {
	ans = &result.ConcResult{Query: args.Query}
	defer func() {
		if r := recover(); r != nil {
			ans = &result.ConcResult{
				Error: fmt.Errorf("%v", r),
				Lines: make([]concordance.Line, 0),
			}
		}
	}()
	concEx, err := mango.GetConcordance(
		args.CorpusPath,
		args.Query,
		args.Attrs,
		[]string{},
		[]string{},
		args.StartLine,
		args.MaxItems,
		args.MaxContext,
		args.ViewContextStruct,
	)
	log.Debug().
		Str("query", args.Query).
		Int("concSize", concEx.ConcSize).
		Err(err).
		Msg("obtained concordance result")
	if err != nil {
		ans.Error = err
		return
	}
	parser := concordance.NewLineParser(args.Attrs)
	ans.Lines = parser.Parse(concEx.Lines)
	ans.ConcSize = concEx.ConcSize
	return
}

func NewWorker(
	ctx context.Context,
	workerID string,
	radapter *rdb.Adapter,
	messages <-chan *redis.Message,
	jobLogger jobLogger,
) *Worker {
	return &Worker{
		ID:        workerID,
		radapter:  radapter,
		messages:  messages,
		ctx:       ctx,
		ticker:    time.NewTicker(DefaultTickerInterval),
		jobLogger: jobLogger,
	}
}
