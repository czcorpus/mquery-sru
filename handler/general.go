// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package handler

import (
	"github.com/czcorpus/mquery-sru/cnf"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"
	v12 "github.com/czcorpus/mquery-sru/handler/v12"
	v20 "github.com/czcorpus/mquery-sru/handler/v20"
	"github.com/czcorpus/mquery-sru/rdb"

	"github.com/gin-gonic/gin"
)

// supported versions
const (
	Version12 = "1.2"
	Version20 = "2.0"

	DefaultVersion = Version20
)

type FCSSubHandler interface {
	Handle(
		ctx *gin.Context,
		fcsResponse general.FCSGeneralResponse,
		xslt map[string]string,
	)
}

type FCSHandler struct {
	conf     *corpus.CorporaSetup
	radapter *rdb.Adapter

	versions map[string]FCSSubHandler
}

func (a *FCSHandler) FCSHandler(ctx *gin.Context) {
	a.handleWithXSLT(
		ctx,
		map[string]string{},
	)
}

func (a *FCSHandler) handleWithXSLT(ctx *gin.Context, xslt map[string]string) {
	resp := general.FCSGeneralResponse{
		Version: ctx.DefaultQuery("version", DefaultVersion),
		Fatal:   false,
		Errors:  make([]general.FCSError, 0, 10),
	}
	handler, ok := a.versions[resp.Version]
	if !ok {
		handler = a.versions[DefaultVersion]
		resp.Version = DefaultVersion
		resp.AddError(general.FCSError{
			Code:    general.DCUnsupportedVersion,
			Ident:   DefaultVersion,
			Message: "Unsupported version " + resp.Version,
		})
	}
	ctx.Set("logEvent_version", resp.Version)
	handler.Handle(ctx, resp, xslt)
}

func NewFCSHandler(
	serverInfo *cnf.ServerInfo,
	corporaConf *corpus.CorporaSetup,
	radapter *rdb.Adapter,
	projectRootDir string,
) *FCSHandler {
	return &FCSHandler{
		conf:     corporaConf,
		radapter: radapter,
		versions: map[string]FCSSubHandler{
			Version12: v12.NewFCSSubHandlerV12(
				serverInfo, corporaConf, radapter, projectRootDir),
			Version20: v20.NewFCSSubHandlerV20(
				serverInfo, corporaConf, radapter, projectRootDir),
		},
	}
}
