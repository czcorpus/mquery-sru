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
	"fcs/corpus"
	"fcs/rdb"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
)

const DefaultVersion = "1.2"

type FCSSubHandler interface {
	Handle(ctx *gin.Context, fcsResponse *FCSResponse)
}

type FCSHandler struct {
	conf     *corpus.CorporaSetup
	radapter *rdb.Adapter
	tmpl     *template.Template

	supportedRecordPackings []string
	supportedOperations     []string

	queryAllow          []string
	queryExplain        []string
	querySearchRetrieve []string

	versions map[string]FCSSubHandler
}

type FCSResourceInfo struct {
	PID         string
	Title       string
	Description string
	URI         string
	Languages   []string
}

type FCSSearchRow struct {
	Position int
	PID      string
	Left     string
	KWIC     string
	Right    string
	Web      string
	Ref      string
}

type FCSExplain struct {
	ServerName          string
	ServerPort          string
	Database            string
	DatabaseTitle       string
	DatabaseDescription string
}

type FCSSearchRetrieve struct {
	Results []FCSSearchRow
}

type FCSResponse struct {
	Version       string
	RecordPacking string
	Operation     string

	MaximumRecords int
	MaximumTerms   int

	Explain        FCSExplain
	Resources      []FCSResourceInfo
	SearchRetrieve FCSSearchRetrieve
	Error          *FCSError
}

func (a *FCSHandler) FCSHandler(ctx *gin.Context) {
	fcsResponse := &FCSResponse{
		Version:        DefaultVersion,
		RecordPacking:  "xml",
		Operation:      "explain",
		MaximumRecords: 250,
		MaximumTerms:   100,
	}

	version := ctx.DefaultQuery("version", DefaultVersion)
	handler, ok := a.versions[version]
	if !ok {
		fcsResponse.Error = &FCSError{
			Code:    CodeUnsupportedVersion,
			Ident:   DefaultVersion,
			Message: "Unsupported version " + version,
		}
		if err := a.tmpl.ExecuteTemplate(ctx.Writer, fmt.Sprintf("fcs-%s.xml", DefaultVersion), fcsResponse); err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	fcsResponse.Version = version
	handler.Handle(ctx, fcsResponse)
}

func NewFCSHandler(
	conf *corpus.CorporaSetup,
	radapter *rdb.Adapter,
) *FCSHandler {
	tmpl := template.Must(template.ParseGlob("templates/*"))
	return &FCSHandler{
		conf:     conf,
		radapter: radapter,
		tmpl:     tmpl,
		versions: map[string]FCSSubHandler{
			"1.2": NewFCSSubHandlerV12(conf, radapter, tmpl),
		},
	}
}
