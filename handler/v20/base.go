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
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/mquery-sru/cnf"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/handler/v20/schema"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

type FCSSubHandlerV20 struct {
	serverInfo  *cnf.ServerInfo
	corporaConf *corpus.CorporaSetup
	radapter    *rdb.Adapter
}

func (a *FCSSubHandlerV20) produceXMLResponse(ctx *gin.Context, code int, xslt string, data any) {
	xmlAns, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Err(err).Msg("failed to encode a result to XML")
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.Writer.WriteHeader(code)
	_, err = ctx.Writer.Write([]byte(xml.Header + general.GetXSLTHeader(xslt) + string(xmlAns)))
	if err != nil {
		log.Err(err).Msg("failed to write XML to response")
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
	}
	ctx.Writer.Header().Set("Content-Type", "application/xml")
}

func (a *FCSSubHandlerV20) produceErrorResponse(ctx *gin.Context, code int, xslt string, fcsErrors []general.FCSError) {
	ans := schema.XMLExplainResponse{
		XMLNSSRUResponse: "http://docs.oasis-open.org/ns/search-ws/sruResponse",
		Version:          "2.0",
		Diagnostics:      schema.NewXMLDiagnostics(),
	}
	for _, fcsErr := range fcsErrors {
		ans.Diagnostics.AddDiagnostic(fcsErr.Code, fcsErr.Type, fcsErr.Ident, fcsErr.Message)
	}
	a.produceXMLResponse(ctx, code, xslt, ans)
}

func (a *FCSSubHandlerV20) Handle(
	ctx *gin.Context,
	fcsGeneralRequest general.FCSGeneralRequest,
	xslt map[string]string,
) {
	fcsRequest := &FCSRequest{
		General:           fcsGeneralRequest,
		RecordXMLEscaping: RecordXMLEscapingXML,
		Operation:         OperationExplain,
	}

	if fcsRequest.General.HasFatalError() {
		a.produceErrorResponse(ctx, general.ConformantStatusBadRequest, fcsGeneralRequest.XSLT, fcsGeneralRequest.Errors)
		return
	}

	var operation Operation = OperationExplain
	if ctx.Request.URL.Query().Has("operation") {
		operation = getTypedArg(ctx, "operation", fcsRequest.Operation)

	} else if ctx.Request.URL.Query().Has(SearchRetrArgQuery.String()) {
		operation = OperationSearchRetrive

	} else if ctx.Request.URL.Query().Has(ScanArgScanClause.String()) {
		operation = OperationScan
	}
	if err := operation.Validate(); err != nil {
		fcsRequest.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedOperation,
			Ident:   "operation",
			Message: fmt.Sprintf("Unsupported operation: %s", operation),
		})
		a.produceErrorResponse(ctx, general.ConformantStatusBadRequest, fcsGeneralRequest.XSLT, fcsGeneralRequest.Errors)
		return
	}
	fcsRequest.Operation = operation
	fcsRequest.General.XSLT = xslt[operation.String()]
	logging.AddLogEvent(ctx, "operation", operation)

	recordXMLEscaping := getTypedArg(ctx, "recordXMLEscaping", fcsRequest.RecordXMLEscaping)
	if err := recordXMLEscaping.Validate(); err != nil {
		fcsRequest.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedRecordPacking,
			Ident:   "recordXMLEscaping",
			Message: err.Error(),
		})
		a.produceErrorResponse(ctx, general.ConformantStatusBadRequest, fcsGeneralRequest.XSLT, fcsGeneralRequest.Errors)
		return
	}
	fcsRequest.RecordXMLEscaping = recordXMLEscaping
	logging.AddLogEvent(ctx, "recordXMLEscaping", recordXMLEscaping)

	var response any
	var code int
	switch fcsRequest.Operation {
	case OperationExplain:
		response, code = a.explain(ctx, fcsRequest)
	case OperationSearchRetrive:
		response, code = a.searchRetrieve(ctx, fcsRequest)
	case OperationScan:
		response, code = a.scan(ctx, fcsRequest)
	}
	a.produceXMLResponse(ctx, code, fcsGeneralRequest.XSLT, response)
}

func NewFCSSubHandlerV20(
	generalConf *cnf.ServerInfo,
	corporaConf *corpus.CorporaSetup,
	radapter *rdb.Adapter,
) *FCSSubHandlerV20 {
	return &FCSSubHandlerV20{
		serverInfo:  generalConf,
		corporaConf: corporaConf,
		radapter:    radapter,
	}
}
