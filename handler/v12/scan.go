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

package v12

import (
	"strconv"

	"github.com/czcorpus/mquery-sru/general"
	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV12) scan(ctx *gin.Context, fcsResponse *FCSResponse) int {
	for key, _ := range ctx.Request.URL.Query() {
		if err := ScanArg(key).Validate(); err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCUnsupportedParameter,
				Ident:   key,
				Message: err.Error(),
			})
			return general.ConformantStatusBadRequest
		}
	}

	xMaxTerms := ctx.DefaultQuery(ScanArgMaximumTerms.String(), "1000")
	_, err := strconv.Atoi(xMaxTerms)
	if err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedParameterValue,
			Ident:   ScanArgMaximumTerms.String(),
			Message: general.DCUnsupportedParameterValue.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}

	xResponsePos := ctx.DefaultQuery(ScanArgResponsePosition.String(), "1")
	_, err = strconv.Atoi(xResponsePos)
	if err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedParameterValue,
			Ident:   ScanArgResponsePosition.String(),
			Message: general.DCUnsupportedParameterValue.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}

	scanClause := ctx.Query(ScanArgScanClause.String())
	if scanClause == "" {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCMandatoryParameterNotSupplied,
			Ident:   ScanArgScanClause.String(),
			Message: general.DCMandatoryParameterNotSupplied.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}

	fcsResponse.General.AddError(general.FCSError{
		Code:    general.DCUnsupportedIndex,
		Ident:   ScanArgScanClause.String(),
		Message: general.DCUnsupportedIndex.AsMessage(),
	})
	return general.ConformantUnprocessableEntity
}
