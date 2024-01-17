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
	"fcs/general"
	"net/http"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV20) explain(ctx *gin.Context, fcsResponse *FCSResponse) int {
	// check if all parameters are supported
	for key, _ := range ctx.Request.URL.Query() {
		if !collections.SliceContains(a.queryGeneral, key) && !collections.SliceContains(a.queryExplain, key) {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeUnsupportedParameter,
				Ident:   key,
				Message: "Unsupported parameter",
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
		PosAttrs:            a.corporaConf.Resources.GetCommonPosAttrs(a.corporaConf.Resources.GetCorpora()...),
	}
	if ctx.Query("x-fcs-endpoint-description") == "true" {
		fcsResponse.Explain.ExtraResponseData = true
		for corpusName, corpusConf := range a.corporaConf.Resources {
			fcsResponse.Explain.Resources = append(
				fcsResponse.Explain.Resources,
				FCSResourceInfo{
					PID:             corpusName,
					Title:           corpusName,
					Description:     "TODO",
					URI:             "TODO",
					Languages:       []string{"cs", "TODO"},
					AvailableLayers: corpusConf.GetDefinedLayersAsString(),
				},
			)
		}
	}
	return http.StatusOK
}
