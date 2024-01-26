// Copyright 2024 Martin Zimandl <martin.zimandl@gmail.com>
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

package handler

import (
	"path"

	"github.com/gin-gonic/gin"
)

type ViewHandler struct {
	fcsHandler    *FCSHandler
	assetsURLPath string
}

func (handler *ViewHandler) Handle(ctx *gin.Context) {
	handler.fcsHandler.handleWithXSLT(
		ctx,
		map[string]string{
			"explain":        path.Join(handler.assetsURLPath, "ui/assets/xslt/explain.xslt"),
			"searchRetrieve": path.Join(handler.assetsURLPath, "ui/assets/xslt/searchRetrieve.xslt"),
			"scan":           path.Join(handler.assetsURLPath, "ui/assets/xslt/scan.xslt"),
		},
	)
}

func NewViewHandler(fcsHandler *FCSHandler, assetsURLPath string) *ViewHandler {
	return &ViewHandler{
		fcsHandler:    fcsHandler,
		assetsURLPath: assetsURLPath,
	}
}
