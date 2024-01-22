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
	"fcs/cnf"
	"fcs/corpus"
	"fcs/general"
	"fcs/handler/common"
	"fcs/rdb"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/gin-gonic/gin"
)

const (
	OperationExplain       Operation     = "explain"
	OperationScan          Operation     = "scan"
	OperationSearchRetrive Operation     = "searchRetrieve"
	QueryTypeCQL           QueryType     = "cql"
	QueryTypeFCS           QueryType     = "fcs"
	RecordPackingXML       RecordPacking = "xml"
	RecordPackingString    RecordPacking = "string"

	SearchRetrArgVersion            SearchRetrArg = "version"
	SearchRetrStartRecord           SearchRetrArg = "startRecord"
	SearchMaximumRecords            SearchRetrArg = "maximumRecords"
	SearchRetrArgRecordPacking      SearchRetrArg = "recordPacking"
	SearchRetrArgOperation          SearchRetrArg = "operation"
	SearchRetrArgQuery              SearchRetrArg = "query"
	SearchRetrArgQueryType          SearchRetrArg = "queryType"
	SearchRetrArgFCSContext         SearchRetrArg = "x-fcs-context"
	SearchRetrArgFCSDataViews       SearchRetrArg = "x-fcs-dataviews"
	SearchRetrArgFCSRewritesAllowed SearchRetrArg = "x-fcs-rewrites-allowed"

	ExplainArgVersion                ExplainArg = "version"
	ExplainArgRecordPacking          ExplainArg = "recordPacking"
	ExplainArgOperation              ExplainArg = "operation"
	ExplainArgFCSEndpointDescription ExplainArg = "x-fcs-endpoint-description"

	DefaultQueryType QueryType = QueryTypeCQL
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

type RecordPacking string

func (rp RecordPacking) Validate() error {
	if rp == RecordPackingString || rp == RecordPackingXML {
		return nil
	}
	return fmt.Errorf("unknown record packing: %s", rp)
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
		sra == SearchRetrArgQueryType ||
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
}

func (a *FCSSubHandlerV20) Handle(ctx *gin.Context, fcsGeneralResponse general.FCSGeneralResponse) {
	fcsResponse := &FCSResponse{
		General:       fcsGeneralResponse,
		RecordPacking: RecordPackingXML,
		Operation:     OperationExplain,
	}

	if fcsResponse.General.Error != nil {
		a.produceResponse(ctx, fcsResponse, http.StatusBadRequest)
		return
	}

	recordPacking := getTypedArg(ctx, "recordPacking", fcsResponse.RecordPacking)
	if err := recordPacking.Validate(); err != nil {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedRecordPacking,
			Ident:   "recordPacking",
			Message: err.Error(),
		}
		a.produceResponse(ctx, fcsResponse, http.StatusBadRequest)
		return
	}
	if recordPacking == "xml" {
		ctx.Writer.Header().Set("Content-Type", "application/xml")
	} else if recordPacking == "string" {
		ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	fcsResponse.RecordPacking = recordPacking

	operation := getTypedArg(ctx, "operation", fcsResponse.Operation)
	if err := operation.Validate(); err != nil {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedOperation,
			Ident:   "operation",
			Message: "Unsupported operation",
		}
		a.produceResponse(ctx, fcsResponse, http.StatusBadRequest)
		return
	}
	fcsResponse.Operation = operation

	code := http.StatusOK
	switch fcsResponse.Operation {
	case "explain":
		code = a.explain(ctx, fcsResponse)
	case "searchRetrieve":
		code = a.searchRetrieve(ctx, fcsResponse)
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
