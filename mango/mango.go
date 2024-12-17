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

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/mquery-common/concordance"
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

type GoConcordance struct {
	Lines    []string
	ConcSize int
}

func GetConcordance(
	corpusPath, query string,
	attrs []string,
	structs []string,
	refs []string,
	fromLine, maxItems, maxContext int,
	viewContextStruct string,
) (GoConcordance, error) {
	if !collections.SliceContains(refs, "#") {
		refs = append([]string{"#"}, refs...)
	}
	ans := C.conc_examples(
		C.CString(corpusPath),
		C.CString(query),
		C.CString(strings.Join(attrs, ",")),
		C.CString(strings.Join(structs, ",")),
		C.CString(strings.Join(refs, ",")),
		C.CString(concordance.RefsEndMark),
		C.longlong(fromLine),
		C.longlong(maxItems),
		C.longlong(maxContext),
		C.CString(viewContextStruct))
	var ret GoConcordance
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
		str := C.GoString(tmp[i])
		// we must test str len as our c++ wrapper may return it
		// e.g. in case our offset is higher than actual num of lines
		if len(str) > 0 {
			ret.Lines = append(ret.Lines, str)
		}
	}
	return ret, nil
}
