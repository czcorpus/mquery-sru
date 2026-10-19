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

// cStrings keeps track of C strings allocated for a C function call
// so they can be released at once (typically via `defer cs.free()`).
type cStrings []*C.char

func (cs *cStrings) new(s string) *C.char {
	ans := C.CString(s)
	*cs = append(*cs, ans)
	return ans
}

func (cs *cStrings) free() {
	for _, s := range *cs {
		C.free(unsafe.Pointer(s))
	}
	*cs = nil
}

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
	if fromLine < 0 {
		panic("GetConcordance - invalid fromLine value")
	}
	if maxItems < 0 {
		panic("GetConcordance - invalid maxItems value")
	}
	if maxContext < 0 {
		panic("GetConcordance - invalid maxContext value")
	}
	if !collections.SliceContains(refs, "#") {
		refs = append([]string{"#"}, refs...)
	}
	var cs cStrings
	defer cs.free()
	ans := C.conc_examples(
		cs.new(corpusPath),
		cs.new(query),
		cs.new(strings.Join(attrs, ",")),
		cs.new(strings.Join(structs, ",")),
		cs.new(strings.Join(refs, ",")),
		cs.new(concordance.RefsEndMark),
		C.longlong(fromLine),
		C.longlong(maxItems),
		C.longlong(maxContext),
		cs.new(viewContextStruct))
	var ret GoConcordance
	ret.Lines = make([]string, 0, maxItems)
	ret.ConcSize = int(ans.concSize)
	if ans.err != nil {
		err := errors.New(C.GoString(ans.err))
		defer C.free(unsafe.Pointer(ans.err))
		if ans.errorCode == 1 {
			return ret, ErrRowsRangeOutOfConc
		}
		return ret, err

	} else if ans.value == nil {
		return ret, nil

	} else {
		defer C.conc_examples_free(ans.value, C.int(ans.size))
	}
	for _, cstr := range unsafe.Slice((**C.char)(unsafe.Pointer(ans.value)), int(ans.size)) {
		str := C.GoString(cstr)
		// we must test str len as our c++ wrapper may return it
		// e.g. in case our offset is higher than actual num of lines
		if len(str) > 0 {
			ret.Lines = append(ret.Lines, str)
		}
	}
	return ret, nil
}
