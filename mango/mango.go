// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package mango

// #cgo LDFLAGS:  -lmanatee -L${SRCDIR} -Wl,-rpath='$ORIGIN'
// #include <stdlib.h>
// #include "mango.h"
import "C"

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

const (
	MaxRecordsInternalLimit = 1000
)

var (
	ErrRowsRangeOutOfConc = errors.New("rows range is out of concordance size")
)

// ---

type GoConcSize struct {
	Value      int64
	CorpusSize int64
}

type GoConcExamples struct {
	Lines    []string
	ConcSize int
}

func GetConcExamples(corpusPath, query string, attrs []string, fromLine, maxItems int, maxContext int) (GoConcExamples, error) {
	ans := C.conc_examples(
		C.CString(corpusPath), C.CString(query), C.CString(strings.Join(attrs, ",")),
		C.longlong(fromLine), C.longlong(maxItems), C.longlong(maxContext))
	var ret GoConcExamples
	ret.Lines = make([]string, 0, maxItems)
	ret.ConcSize = int(ans.concSize)
	if ans.err != nil {
		err := fmt.Errorf(C.GoString(ans.err))
		defer C.free(unsafe.Pointer(ans.err))
		if ans.errorCode == 1 {
			return ret, ErrRowsRangeOutOfConc
		}
		return ret, err

	} else {
		defer C.conc_examples_free(ans.value, C.int(ans.size))
	}
	tmp := (*[MaxRecordsInternalLimit]*C.char)(unsafe.Pointer(ans.value))
	for i := 0; i < int(ans.size); i++ {
		ret.Lines = append(ret.Lines, C.GoString(tmp[i]))
	}
	return ret, nil
}
