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
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/bytedance/sonic"
	"github.com/czcorpus/mquery-sru/corpus/conc"
	"github.com/czcorpus/mquery-sru/mango"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/czcorpus/mquery-sru/results"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	DefaultTickerInterval = 2 * time.Second
	MaxFreqResultItems    = 100
)

type jobLogger interface {
	Log(rec results.JobLog)
}

type Worker struct {
	ID         string
	messages   <-chan *redis.Message
	radapter   *rdb.Adapter
	exitEvent  chan os.Signal
	ticker     time.Ticker
	jobLogger  jobLogger
	currJobLog *results.JobLog
}

func (w *Worker) publishResult(res results.SerializableResult, channel string) error {
	ans, err := rdb.CreateWorkerResult(res)
	if err != nil {
		return err
	}

	w.currJobLog.End = time.Now()
	w.currJobLog.Err = res.Err()
	w.jobLogger.Log(*w.currJobLog)
	w.currJobLog = nil
	return w.radapter.PublishResult(channel, ans)
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

	w.currJobLog = &results.JobLog{
		WorkerID: w.ID,
		Func:     query.Func,
		Begin:    time.Now(),
	}

	switch query.Func {
	case "concExample":
		var args rdb.ConcExampleArgs
		if err := sonic.Unmarshal(query.Args, &args); err != nil {
			return err
		}
		ans := w.concExample(args)
		ans.ResultType = query.ResultType
		if err := w.publishResult(ans, query.Channel); err != nil {
			return err
		}
	default:
		ans := &results.ErrorResult{Error: fmt.Sprintf("unknown query function: %s", query.Func)}
		if err = w.publishResult(ans, query.Channel); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) Listen() {
	for {
		select {
		case <-w.ticker.C:
			w.tryNextQuery()
		case <-w.exitEvent:
			log.Info().Msg("worker exiting")
			return
		case msg := <-w.messages:
			if msg.Payload == rdb.MsgNewQuery {
				w.tryNextQuery()
			}
		}
	}
}

func (w *Worker) tokenCoverage(mktokencovPath, subcPath, corpusPath, structure string) error {
	cmd := exec.Command(mktokencovPath, corpusPath, structure, "-s", subcPath)
	return cmd.Run()
}

func (w *Worker) concExample(args rdb.ConcExampleArgs) *results.ConcExample {
	var ans results.ConcExample
	concEx, err := mango.GetConcExamples(
		args.CorpusPath, args.Query, args.Attrs, args.StartLine, args.MaxItems)
	if err != nil {
		ans.Error = err.Error()
		return &ans
	}
	parser := conc.NewLineParser(args.Attrs)
	ans.Lines = parser.Parse(concEx)
	ans.ConcSize = concEx.ConcSize
	return &ans
}

func NewWorker(
	workerID string,
	radapter *rdb.Adapter,
	messages <-chan *redis.Message,
	exitEvent chan os.Signal,
	jobLogger jobLogger,
) *Worker {
	return &Worker{
		ID:        workerID,
		radapter:  radapter,
		messages:  messages,
		exitEvent: exitEvent,
		ticker:    *time.NewTicker(DefaultTickerInterval),
		jobLogger: jobLogger,
	}
}
