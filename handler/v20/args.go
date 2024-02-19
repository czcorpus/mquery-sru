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

package v20

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	OperationExplain        Operation         = "explain"
	OperationScan           Operation         = "scan"
	OperationSearchRetrive  Operation         = "searchRetrieve"
	QueryTypeCQL            QueryType         = "cql"
	QueryTypeFCS            QueryType         = "fcs"
	RecordXMLEscapingXML    RecordXMLEscaping = "xml"
	RecordXMLEscapingString RecordXMLEscaping = "string" // TODO for now unsupported

	SearchRetrArgVersion            SearchRetrArg = "version"
	SearchRetrStartRecord           SearchRetrArg = "startRecord"
	SearchMaximumRecords            SearchRetrArg = "maximumRecords"
	SearchRetrArgRecordXMLEscaping  SearchRetrArg = "recordXMLEscaping"
	SearchRetrArgOperation          SearchRetrArg = "operation"
	SearchRetrArgQuery              SearchRetrArg = "query"
	SearchRetrArgQueryType          SearchRetrArg = "queryType"
	SearchRetrArgRecordSchema       SearchRetrArg = "recordSchema"
	SearchRetrArgFCSContext         SearchRetrArg = "x-fcs-context"
	SearchRetrArgFCSDataViews       SearchRetrArg = "x-fcs-dataviews"
	SearchRetrArgFCSRewritesAllowed SearchRetrArg = "x-fcs-rewrites-allowed"

	ScanArgVersion           ScanArg = "version"
	ScanArgOperation         ScanArg = "operation"
	ScanArgRecordXMLEscaping ScanArg = "recordXMLEscaping"
	ScanArgScanClause        ScanArg = "scanClause"
	ScanArgMaximumTerms      ScanArg = "maximumTerms"
	ScanArgResponsePosition  ScanArg = "responsePosition"

	ExplainArgVersion                ExplainArg = "version"
	ExplainArgRecordXMLEscaping      ExplainArg = "recordXMLEscaping"
	ExplainArgOperation              ExplainArg = "operation"
	ExplainArgFCSEndpointDescription ExplainArg = "x-fcs-endpoint-description"

	DefaultQueryType QueryType = QueryTypeCQL
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

type QueryType string

func (qt QueryType) Validate() error {
	if qt == QueryTypeCQL || qt == QueryTypeFCS {
		return nil
	}
	return fmt.Errorf("unknown query type: %s", qt)
}

func (qt QueryType) String() string {
	return string(qt)
}

// ----

type RecordXMLEscaping string

func (rp RecordXMLEscaping) Validate() error {
	if rp == RecordXMLEscapingXML {
		return nil
	}
	return fmt.Errorf("unsupported record XML escaping: %s", rp)
}

// ----

type SearchRetrArg string

func (sra SearchRetrArg) Validate() error {
	if sra == SearchRetrArgVersion ||
		sra == SearchRetrStartRecord ||
		sra == SearchMaximumRecords ||
		sra == SearchRetrArgRecordXMLEscaping ||
		sra == SearchRetrArgOperation ||
		sra == SearchRetrArgQuery ||
		sra == SearchRetrArgQueryType ||
		sra == SearchRetrArgRecordSchema ||
		sra == SearchRetrArgFCSContext ||
		sra == SearchRetrArgFCSDataViews ||
		sra == SearchRetrArgFCSRewritesAllowed {
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
		sa == ScanArgRecordXMLEscaping ||
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
		arg == ExplainArgRecordXMLEscaping ||
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
