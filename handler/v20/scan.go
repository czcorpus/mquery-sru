// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2024 Institute of the Czech National Corpus,
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
	"strconv"

	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/handler/v20/schema"
	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV20) scan(ctx *gin.Context, _ *FCSRequest) (schema.XMLScanResponse, int) {
	ans := schema.NewXMLScanResponse()
	for key, _ := range ctx.Request.URL.Query() {
		if err := ScanArg(key).Validate(); err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDiagnostic(
				general.DCUnsupportedParameter, 0, key, err.Error())
			return ans, general.ConformantStatusBadRequest
		}
	}

	xMaxTerms := ctx.DefaultQuery(ScanArgMaximumTerms.String(), "1000")
	_, err := strconv.Atoi(xMaxTerms)
	if err != nil {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnsupportedParameterValue, 0, ScanArgMaximumTerms.String())
		return ans, general.ConformantUnprocessableEntity
	}

	xResponsePos := ctx.DefaultQuery(ScanArgResponsePosition.String(), "1")
	_, err = strconv.Atoi(xResponsePos)
	if err != nil {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnsupportedParameterValue, 0, ScanArgResponsePosition.String())
		return ans, general.ConformantUnprocessableEntity
	}

	scanClause := ctx.Query(ScanArgScanClause.String())
	if scanClause == "" {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCMandatoryParameterNotSupplied, 0, ScanArgScanClause.String())
		return ans, general.ConformantUnprocessableEntity
	}

	ans.Diagnostics = schema.NewXMLDiagnostics()
	ans.Diagnostics.AddDfltMsgDiagnostic(
		general.DCUnsupportedIndex, 0, ScanArgScanClause.String())
	return ans, general.ConformantUnprocessableEntity
}
