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
	"net/http"

	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"

	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV20) explain(ctx *gin.Context, fcsResponse *FCSResponse) int {
	// prepare response data
	fcsResponse.Explain = &FCSExplain{
		ServerName:          a.serverInfo.ServerHost,
		ServerPort:          a.serverInfo.ServerPort,
		Database:            a.serverInfo.Database,
		DatabaseTitle:       a.serverInfo.DatabaseTitle,
		DatabaseDescription: a.serverInfo.DatabaseDescription,
		DatabaseAuthor:      a.serverInfo.DatabaseAuthor,
		PrimaryLanguage:     a.serverInfo.PrimaryLanguage,
		MaximumRecords:      a.corporaConf.MaximumRecords,
		NumberOfRecords:     corpus.ExplainOpNumberOfRecords,
		PosAttrs:            a.corporaConf.Resources.GetCommonPosAttrs2(),
	}

	// check if all parameters are supported
	for key, _ := range ctx.Request.URL.Query() {
		if err := ExplainArg(key).Validate(); err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCUnsupportedParameter,
				Ident:   key,
				Message: err.Error(),
			})
			return general.ConformantStatusBadRequest
		}
	}

	// get resources
	if ctx.Query(ExplainArgFCSEndpointDescription.String()) == "true" {
		fcsResponse.Explain.ExtraResponseData = true
		for _, corpusConf := range a.corporaConf.Resources {
			fcsResponse.Explain.Resources = append(
				fcsResponse.Explain.Resources,
				FCSResourceInfo{
					PID:             corpusConf.PID,
					Title:           corpusConf.FullName,
					Description:     corpusConf.Description,
					URI:             corpusConf.URI,
					Languages:       corpusConf.Languages,
					AvailableLayers: corpusConf.GetDefinedLayersAsRefString(),
				},
			)
		}
	}
	return http.StatusOK
}
