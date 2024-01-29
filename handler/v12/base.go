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

type FCSSubHandlerV12 struct {
	serverInfo  *cnf.ServerInfo
	corporaConf *corpus.CorporaSetup
	radapter    *rdb.Adapter
	tmpl        *template.Template
}

func (a *FCSSubHandlerV12) produceResponse(ctx *gin.Context, fcsResponse *FCSResponse, code int) {
	ctx.Writer.WriteHeader(code)
	// TODO in case an error occurs in executing of the template, how can we rewrite
	// an already written status header? (see docs for ctx.Write)
	if err := a.tmpl.ExecuteTemplate(ctx.Writer, "fcs-1.2.xml", fcsResponse); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.Writer.Header().Set("Content-Type", "application/xml")
}

func (a *FCSSubHandlerV12) Handle(
	ctx *gin.Context,
	fcsGeneralResponse general.FCSGeneralResponse,
	xslt map[string]string,
) {
	fcsResponse := &FCSResponse{
		General:       fcsGeneralResponse,
		RecordPacking: RecordPackingXML,
		Operation:     OperationExplain,
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

	recordPacking := getTypedArg(ctx, "recordPacking", fcsResponse.RecordPacking)
	if err := recordPacking.Validate(); err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedRecordPacking,
			Ident:   "recordPacking",
			Message: err.Error(),
		})
		a.produceResponse(ctx, fcsResponse, general.ConformantStatusBadRequest)
		return
	}
	fcsResponse.RecordPacking = recordPacking

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

func NewFCSSubHandlerV12(
	generalConf *cnf.ServerInfo,
	corporaConf *corpus.CorporaSetup,
	radapter *rdb.Adapter,
	projectRootDir string,
) *FCSSubHandlerV12 {
	path := filepath.Join(projectRootDir, "handler", "v12", "templates")
	return &FCSSubHandlerV12{
		serverInfo:  generalConf,
		corporaConf: corporaConf,
		radapter:    radapter,
		tmpl: template.Must(
			template.New("").
				Funcs(common.GetTemplateFunctions()).
				ParseGlob(path + "/*")),
	}
}
