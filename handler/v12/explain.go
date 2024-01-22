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
	"fcs/general"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV12) explain(ctx *gin.Context, fcsResponse *FCSResponse) int {
	// check if all parameters are supported
	for key, _ := range ctx.Request.URL.Query() {
		if err := ExplainArg(key).Validate(); err != nil {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeUnsupportedParameter,
				Ident:   key,
				Message: err.Error(),
			}
			return http.StatusBadRequest
		}
	}

	// prepare response data
	fcsResponse.Explain = FCSExplain{
		ServerName:          a.serverInfo.ServerName,
		ServerPort:          a.serverInfo.ServerPort,
		Database:            a.serverInfo.Database,
		DatabaseTitle:       a.serverInfo.DatabaseTitle,
		DatabaseDescription: a.serverInfo.DatabaseDescription,
		PrimaryLanguage:     a.serverInfo.PrimaryLanguage,
	}
	if ctx.Query(ExplainArgFCSEndpointDescription.String()) == "true" {
		fcsResponse.Explain.ExtraResponseData = true
		for _, corpusConf := range a.corporaConf.Resources {
			fcsResponse.Explain.Resources = append(
				fcsResponse.Explain.Resources,
				FCSResourceInfo{
					PID:         corpusConf.PID,
					Title:       corpusConf.FullName,
					Description: corpusConf.Description,
					URI:         corpusConf.URI,
					Languages:   corpusConf.Languages,
				},
			)
		}
	}
	return http.StatusOK
}
