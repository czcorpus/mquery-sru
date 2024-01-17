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
	"fcs/rdb"
	"net/http"
	"text/template"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/gin-gonic/gin"
)

type FCSSubHandlerV20 struct {
	serverInfo  *cnf.ServerInfo
	corporaConf *corpus.CorporaSetup
	radapter    *rdb.Adapter
	tmpl        *template.Template

	supportedRecordPackings []string
	supportedOperations     []string
	supportedQueryTypes     []string

	queryGeneral        []string
	queryExplain        []string
	querySearchRetrieve []string
}

func (a *FCSSubHandlerV20) produceResponse(ctx *gin.Context, fcsResponse *FCSResponse, code int) {
	if err := a.tmpl.ExecuteTemplate(ctx.Writer, "fcs-2.0.xml", fcsResponse); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.Writer.WriteHeader(code)
}

func (a *FCSSubHandlerV20) Handle(ctx *gin.Context, fcsGeneralResponse general.FCSGeneralResponse) {
	fcsResponse := &FCSResponse{
		General:       fcsGeneralResponse,
		RecordPacking: "xml",
		Operation:     "explain",
	}
	if fcsResponse.General.Error != nil {
		a.produceResponse(ctx, fcsResponse, http.StatusBadRequest)
		return
	}

	recordPacking := ctx.DefaultQuery("recordPacking", fcsResponse.RecordPacking)
	if !collections.SliceContains(a.supportedRecordPackings, recordPacking) {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedRecordPacking,
			Ident:   "recordPacking",
			Message: "Unsupported record packing",
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

	operation := ctx.DefaultQuery("operation", fcsResponse.Operation)
	if !collections.SliceContains(a.supportedOperations, operation) {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedOperation,
			Ident:   "",
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
) *FCSSubHandlerV20 {
	return &FCSSubHandlerV20{
		serverInfo:              generalConf,
		corporaConf:             corporaConf,
		radapter:                radapter,
		tmpl:                    template.Must(template.New("").Funcs(general.GetTemplateFuncMap()).ParseGlob("handler/v20/templates/*")),
		supportedOperations:     []string{"explain", "scan", "searchRetrieve"},
		supportedQueryTypes:     []string{"cql", "fcs"},
		supportedRecordPackings: []string{"xml", "string"},
		queryGeneral:            []string{"version", "recordPacking", "operation"},
		queryExplain:            []string{"x-fcs-endpoint-description"},
		querySearchRetrieve:     []string{"query", "queryType", "x-fcs-context", "x-fcs-dataviews", "x-fcs-rewrites-allowed"},
	}
}
