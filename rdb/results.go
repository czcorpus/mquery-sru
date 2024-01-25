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

package rdb

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/czcorpus/mquery-sru/results"
)

type WorkerResult struct {
	ID         string             `json:"id"`
	ResultType results.ResultType `json:"resultType"`
	Value      json.RawMessage    `json:"value"`
}

func (wr *WorkerResult) AttachValue(value results.SerializableResult) error {
	rawValue, err := sonic.Marshal(value)
	if err != nil {
		return err
	}
	wr.Value = rawValue
	return nil
}

func CreateWorkerResult(value results.SerializableResult) (*WorkerResult, error) {
	rawValue, err := sonic.Marshal(value)
	if err != nil {
		return nil, err
	}
	return &WorkerResult{Value: rawValue, ResultType: value.Type()}, nil
}

func DeserializeConcExampleResult(w *WorkerResult) (results.ConcExample, error) {
	var ans results.ConcExample
	err := sonic.Unmarshal(w.Value, &ans)
	if err != nil {
		return ans, fmt.Errorf("failed to deserialize ConcExample: %w", err)
	}
	return ans, nil
}
