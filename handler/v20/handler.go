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

package v20

import (
	"encoding/json"
	"fcs/corpus"
	"fcs/general"
	"fcs/rdb"
	"fcs/results"
	"fcs/transformers"
	"fcs/transformers/advanced"
	"fcs/transformers/basic"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/gin-gonic/gin"
)

type FCSSubHandlerV20 struct {
	conf     *corpus.CorporaSetup
	radapter *rdb.Adapter
	tmpl     *template.Template

	supportedRecordPackings []string
	supportedOperations     []string
	supportedQueryTypes     []string

	queryGeneral        []string
	queryExplain        []string
	querySearchRetrieve []string
}

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
		ServerName:          ctx.Request.URL.Host,   // TODO
		ServerPort:          ctx.Request.URL.Port(), // TODO
		Database:            ctx.Request.URL.Path,   // TODO
		DatabaseTitle:       "TODO",
		DatabaseDescription: "TODO",
		Layers:              a.conf.Layers,
	}
	if ctx.Query("x-fcs-endpoint-description") == "true" {
		for corpusName, _ := range a.conf.Resources {
			fcsResponse.Resources = append(
				fcsResponse.Resources,
				FCSResourceInfo{
					PID:             corpusName,
					Title:           corpusName,
					Description:     "TODO",
					URI:             "TODO",
					Languages:       []string{"cs", "TODO"},
					AvailableLayers: "text TODO",
				},
			)
		}
	}
	return http.StatusOK
}

func (a *FCSSubHandlerV20) searchRetrieve(ctx *gin.Context, fcsResponse *FCSResponse) int {
	// check if all parameters are supported
	for key, _ := range ctx.Request.URL.Query() {
		if !collections.SliceContains(a.queryGeneral, key) && !collections.SliceContains(a.querySearchRetrieve, key) {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeUnsupportedParameter,
				Ident:   key,
				Message: "Unsupported parameter",
			}
			return http.StatusBadRequest
		}
	}

	// prepare query
	fcsQuery := ctx.Query("query")
	if len(fcsQuery) == 0 {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeMandatoryParameterNotSupplied,
			Ident:   "fcs_query",
			Message: "Mandatory parameter not supplied",
		}
		return http.StatusBadRequest
	}

	var transformer transformers.Transformer
	var fcsErr *general.FCSError
	queryType := ctx.DefaultQuery("queryType", "cql")
	switch queryType {
	case "cql":
		transformer, fcsErr = basic.NewBasicTransformer(fcsQuery)
	case "fcs":
		transformer, fcsErr = advanced.NewAdvancedTransformer(fcsQuery)
	default:
		fcsErr = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   queryType,
			Message: "Unsupported queryType value",
		}
	}
	if fcsErr != nil {
		fcsResponse.General.Error = fcsErr
		return http.StatusInternalServerError
	}
	fcsResponse.SearchRetrieve.QueryType = queryType

	// get searchable corpora and attrs
	var corpora, searchAttrs []string
	if ctx.Request.URL.Query().Has("x-fcs-context") {
		for _, v := range strings.Split(ctx.Query("x-fcs-context"), ",") {
			resource, ok := a.conf.Resources[v]
			if !ok {
				fcsResponse.General.Error = &general.FCSError{
					Code:    general.CodeUnsupportedParameterValue,
					Ident:   "x-fcs-context",
					Message: "Unknown context " + v,
				}
				return http.StatusBadRequest
			}
			corpora = append(corpora, v)
			searchAttrs = append(searchAttrs, resource.DefaultAttr)
		}
	} else {
		for corpusName, resource := range a.conf.Resources {
			corpora = append(corpora, corpusName)
			searchAttrs = append(searchAttrs, resource.DefaultAttr)
		}
	}

	// make searches
	waits := make([]<-chan *rdb.WorkerResult, len(corpora))
	layerAttrs := []string{}
	for _, attr := range a.conf.Layers {
		layerAttrs = append(layerAttrs, attr)
	}
	for i, corpusName := range corpora {
		query, fcsErr := transformer.CreateCQL(searchAttrs[i])
		if fcsErr != nil {
			fcsResponse.General.Error = fcsErr
			return http.StatusInternalServerError
		}
		args, err := json.Marshal(rdb.ConcExampleArgs{
			CorpusPath:    a.conf.GetRegistryPath(corpusName),
			QueryLemma:    "",
			Query:         query,
			Attrs:         append([]string{a.conf.Layers["text"]}, layerAttrs...),
			MaxItems:      10,
			ParentIdxAttr: a.conf.Resources[corpusName].SyntaxParentAttr.Name,
		})
		if err != nil {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeGeneralSystemError,
				Ident:   err.Error(),
				Message: "General system error",
			}
			return http.StatusInternalServerError
		}
		wait, err := a.radapter.PublishQuery(rdb.Query{
			Func: "concExample",
			Args: args,
		})
		if err != nil {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeGeneralSystemError,
				Ident:   err.Error(),
				Message: "General system error",
			}
			return http.StatusInternalServerError
		}
		waits[i] = wait
	}

	// gather results
	results := make([]results.ConcExample, len(corpora))
	for i, wait := range waits {
		rawResult := <-wait
		result, err := rdb.DeserializeConcExampleResult(rawResult)
		if err != nil {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeGeneralSystemError,
				Ident:   err.Error(),
				Message: "General system error",
			}
			return http.StatusInternalServerError
		}
		if err := result.Err(); err != nil {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeGeneralSystemError,
				Ident:   err.Error(),
				Message: "General system error",
			}
			return http.StatusInternalServerError
		}
		results[i] = result
	}

	// transform results
	fcsResponse.SearchRetrieve.Results = make([]FCSSearchRow, 0, 100)
	for i, r := range results {
		for _, l := range r.Lines {
			segmentPos := 1
			row := FCSSearchRow{
				LayerAttrs: layerAttrs,
				Position:   len(fcsResponse.SearchRetrieve.Results) + 1,
				PID:        corpora[i],
				Web:        "TODO",
				Ref:        "TODO",
			}
			for j, t := range l.Text {
				token := Token{
					Text: t.Word,
					Hit:  t.Strong,
					Segment: Segment{
						ID:    fmt.Sprintf("s%d", j),
						Start: segmentPos,
						End:   segmentPos + len(t.Word) - 1,
					},
					Layers: t.Attrs,
				}
				segmentPos += len(t.Word) + 1 // with space between words
				row.Tokens = append(row.Tokens, token)

			}
			fcsResponse.SearchRetrieve.Results = append(fcsResponse.SearchRetrieve.Results, row)
		}
	}
	return http.StatusOK
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
		General:        fcsGeneralResponse,
		RecordPacking:  "xml",
		Operation:      "explain",
		MaximumRecords: 250,
		MaximumTerms:   100,
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
	conf *corpus.CorporaSetup,
	radapter *rdb.Adapter,
	tmpl *template.Template,
) *FCSSubHandlerV20 {
	return &FCSSubHandlerV20{
		conf:                    conf,
		radapter:                radapter,
		tmpl:                    tmpl,
		supportedOperations:     []string{"explain", "scan", "searchRetrieve"},
		supportedQueryTypes:     []string{"cql", "fcs"},
		supportedRecordPackings: []string{"xml", "string"},
		queryGeneral:            []string{"version", "recordPacking", "operation"},
		queryExplain:            []string{"x-fcs-endpoint-description"},
		querySearchRetrieve:     []string{"query", "queryType", "x-fcs-context", "x-fcs-dataviews", "x-fcs-rewrites-allowed"},
	}
}
