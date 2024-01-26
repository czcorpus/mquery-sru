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
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/czcorpus/mquery-sru/cnf"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/handler/common"
	"github.com/czcorpus/mquery-sru/rdb"

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

	ScanArgScanClause       ScanArg = "scanClause"
	ScanArgMaximumTerms     ScanArg = "maximumTerms"
	ScanArgResponsePosition ScanArg = "responsePosition"

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
	if sa == ScanArgScanClause ||
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

type FCSSubHandlerV20 struct {
	serverInfo  *cnf.ServerInfo
	corporaConf *corpus.CorporaSetup
	radapter    *rdb.Adapter
	tmpl        *template.Template
}

func (a *FCSSubHandlerV20) produceResponse(ctx *gin.Context, fcsResponse *FCSResponse, code int) {
	ctx.Writer.WriteHeader(code)
	// TODO in case an error occurs in executing of the template, how can we rewrite
	// an already written status header? (see docs for ctx.Write)
	if err := a.tmpl.ExecuteTemplate(ctx.Writer, "fcs-2.0.xml", fcsResponse); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.Writer.Header().Set("Content-Type", "application/xml")
}

func (a *FCSSubHandlerV20) Handle(
	ctx *gin.Context,
	fcsGeneralResponse general.FCSGeneralResponse,
	xslt map[string]string,
) {
	fcsResponse := &FCSResponse{
		General:           fcsGeneralResponse,
		RecordXMLEscaping: RecordXMLEscapingXML,
		Operation:         OperationExplain,
	}

	if fcsResponse.General.HasFatalError() {
		a.produceResponse(ctx, fcsResponse, general.ConformantStatusBadRequest)
		return
	}

	var operation Operation = OperationExplain
	if ctx.Request.URL.Query().Has("operation") {
		operation = getTypedArg(ctx, "operation", fcsResponse.Operation)

	} else if ctx.Request.URL.Query().Has(SearchRetrArgQuery.String()) {
		operation = OperationSearchRetrive

	} else if ctx.Request.URL.Query().Has(ScanArgScanClause.String()) {
		operation = OperationScan
	}
	if err := operation.Validate(); err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedOperation,
			Ident:   "operation",
			Message: fmt.Sprintf("Unsupported operation: %s", operation),
		})
		a.produceResponse(ctx, fcsResponse, general.ConformantStatusBadRequest)
		return
	}
	fcsResponse.Operation = operation
	fcsResponse.General.XSLT = xslt[operation.String()]

	recordXMLEscaping := getTypedArg(ctx, "recordXMLEscaping", fcsResponse.RecordXMLEscaping)
	if err := recordXMLEscaping.Validate(); err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedRecordPacking,
			Ident:   "recordXMLEscaping",
			Message: err.Error(),
		})
		a.produceResponse(ctx, fcsResponse, general.ConformantStatusBadRequest)
		return
	}
	fcsResponse.RecordXMLEscaping = recordXMLEscaping

	code := http.StatusOK
	switch fcsResponse.Operation {
	case OperationExplain:
		code = a.explain(ctx, fcsResponse)
	case OperationSearchRetrive:
		code = a.searchRetrieve(ctx, fcsResponse)
	case OperationScan:
		code = a.scan(ctx, fcsResponse)
	}
	a.produceResponse(ctx, fcsResponse, code)
}

func NewFCSSubHandlerV20(
	generalConf *cnf.ServerInfo,
	corporaConf *corpus.CorporaSetup,
	radapter *rdb.Adapter,
	projectRootDir string,
) *FCSSubHandlerV20 {
	path := filepath.Join(projectRootDir, "handler", "v20", "templates")
	return &FCSSubHandlerV20{
		serverInfo:  generalConf,
		corporaConf: corporaConf,
		radapter:    radapter,
		tmpl: template.Must(
			template.New("").
				Funcs(common.GetTemplateFunctions()).
				ParseGlob(path + "/*")),
	}
}
