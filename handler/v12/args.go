// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package v12

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	OperationExplain       Operation     = "explain"
	OperationScan          Operation     = "scan"
	OperationSearchRetrive Operation     = "searchRetrieve"
	RecordPackingXML       RecordPacking = "xml"
	RecordPackingString    RecordPacking = "string" // TODO for now unsupported

	SearchRetrArgVersion       SearchRetrArg = "version"
	SearchRetrStartRecord      SearchRetrArg = "startRecord"
	SearchMaximumRecords       SearchRetrArg = "maximumRecords"
	SearchRetrArgRecordPacking SearchRetrArg = "recordPacking"
	SearchRetrArgOperation     SearchRetrArg = "operation"
	SearchRetrArgQuery         SearchRetrArg = "query"
	SearchRetrArgFCSContext    SearchRetrArg = "x-fcs-context"
	SearchRetrArgFCSDataViews  SearchRetrArg = "x-fcs-dataviews"
	SearchRetrArgRecordSchema  SearchRetrArg = "recordSchema"

	ScanArgVersion          ScanArg = "version"
	ScanArgOperation        ScanArg = "operation"
	ScanArgRecordPacking    ScanArg = "recordPacking"
	ScanArgScanClause       ScanArg = "scanClause"
	ScanArgMaximumTerms     ScanArg = "maximumTerms"
	ScanArgResponsePosition ScanArg = "responsePosition"

	ExplainArgVersion                ExplainArg = "version"
	ExplainArgRecordPacking          ExplainArg = "recordPacking"
	ExplainArgOperation              ExplainArg = "operation"
	ExplainArgFCSEndpointDescription ExplainArg = "x-fcs-endpoint-description"
)

type Operation string

func (op Operation) String() string {
	return string(op)
}

func (op Operation) Validate() error {
	if op == OperationExplain || op == OperationScan ||
		op == OperationSearchRetrive {
		return nil
	}
	return fmt.Errorf("unknown operation: %s", op)
}

// ----

type RecordPacking string

func (rp RecordPacking) Validate() error {
	if rp == RecordPackingXML {
		return nil
	}
	return fmt.Errorf("unsupported record packing: %s", rp)
}

// ----

type SearchRetrArg string

func (sra SearchRetrArg) Validate() error {
	if sra == SearchRetrArgVersion ||
		sra == SearchRetrStartRecord ||
		sra == SearchMaximumRecords ||
		sra == SearchRetrArgRecordPacking ||
		sra == SearchRetrArgOperation ||
		sra == SearchRetrArgQuery ||
		sra == SearchRetrArgFCSContext ||
		sra == SearchRetrArgRecordSchema ||
		sra == SearchRetrArgFCSDataViews {
		return nil
	}
	return fmt.Errorf("unknown searchRetrieve argument: %s", sra)
}

func (sra SearchRetrArg) String() string {
	return string(sra)
}

// -----

type ScanArg string

func (sa ScanArg) String() string {
	return string(sa)
}

func (sa ScanArg) Validate() error {
	if sa == ScanArgVersion ||
		sa == ScanArgOperation ||
		sa == ScanArgRecordPacking ||
		sa == ScanArgScanClause ||
		sa == ScanArgMaximumTerms ||
		sa == ScanArgResponsePosition {
		return nil
	}
	return fmt.Errorf("unknown scan argument: %s", sa)
}

// ----

type ExplainArg string

func (arg ExplainArg) Validate() error {
	if arg == ExplainArgVersion ||
		arg == ExplainArgRecordPacking ||
		arg == ExplainArgOperation ||
		arg == ExplainArgFCSEndpointDescription {
		return nil
	}
	return fmt.Errorf("unknown explain argument: %s", arg)
}

func (arg ExplainArg) String() string {
	return string(arg)
}

// ----

func getTypedArg[T ~string](ctx *gin.Context, name string, dflt T) T {
	v := ctx.DefaultQuery(name, string(dflt))
	return T(v)
}

// ----
