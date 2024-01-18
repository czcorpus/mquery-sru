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
	"encoding/json"
	"fcs/corpus"
	"fcs/general"
	"fcs/query/compiler"
	"fcs/query/parser/fcsql"
	"fcs/query/parser/simple"
	"fcs/rdb"
	"fcs/results"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV20) translateQuery(
	corpusName, query string,
	queryType QueryType,
) (compiler.AST, *general.FCSError) {
	var ast compiler.AST
	var fcsErr *general.FCSError
	switch queryType {
	case QueryTypeCQL:
		var err error
		ast, err = simple.ParseQuery(
			query,
			a.corporaConf.Resources[corpusName].GetSimpleSearchAttrs(),
			corpus.DefaultLayerType,
			a.corporaConf.Resources[corpusName].PosAttrs,
			a.corporaConf.Resources[corpusName].StructureMapping,
		)
		if err != nil {
			fcsErr = &general.FCSError{
				Code:    general.CodeQuerySyntaxError,
				Ident:   query,
				Message: "Invalid query syntax",
			}
		}
	case QueryTypeFCS:
		var err error
		ast, err = fcsql.ParseQuery(
			query,
			corpus.DefaultLayerType,
			a.corporaConf.Resources[corpusName].PosAttrs,
			a.corporaConf.Resources[corpusName].StructureMapping,
		)
		if err != nil {
			fcsErr = &general.FCSError{
				Code:    general.CodeQuerySyntaxError,
				Ident:   query,
				Message: "Invalid query syntax",
			}
		}

	default:
		fcsErr = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   queryType.String(),
			Message: "Unsupported queryType value",
		}
	}
	return ast, fcsErr
}

func (a *FCSSubHandlerV20) exportAttrsByLayers(
	word string,
	attrs map[string]string,
	layers []corpus.LayerType,
	posAttrs []corpus.PosAttr,
) map[corpus.LayerType]string {
	ans := make(map[corpus.LayerType]string)
	for _, layer := range layers {
		if layer == corpus.DefaultLayerType {
			ans[layer] = word
			// TODO this won't work for custom attributes requested from the 'text' layer
		} else {
			var found bool
			for _, posAttr := range posAttrs {
				if posAttr.Layer == layer {
					if v, ok := attrs[posAttr.Name]; ok {
						ans[layer] = v
						found = true
						break
					}
				}
			}
			if !found {
				ans[layer] = "??"
			}
		}
	}
	return ans
}

func (a *FCSSubHandlerV20) searchRetrieve(ctx *gin.Context, fcsResponse *FCSResponse) int {
	// check if all parameters are supported
	for key, _ := range ctx.Request.URL.Query() {
		if err := SearchRetrArg(key).Validate(); err != nil {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeUnsupportedParameter,
				Ident:   key,
				Message: err.Error(),
			}
			return http.StatusBadRequest
		}
	}

	fcsQuery := ctx.Query("query")
	if len(fcsQuery) == 0 {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeMandatoryParameterNotSupplied,
			Ident:   "fcs_query",
			Message: "Mandatory parameter not supplied",
		}
		return http.StatusBadRequest
	}

	corpora := a.corporaConf.Resources.GetCorpora()
	if ctx.Request.URL.Query().Has(ctx.Query(SearchRetrArgFCSContext.String())) {
		corpora = strings.Split(ctx.Query(SearchRetrArgFCSContext.String()), ",")
	}

	// get searchable corpora and attrs
	if len(corpora) > 0 {
		for _, v := range corpora {
			_, ok := a.corporaConf.Resources[v]
			if !ok {
				fcsResponse.General.Error = &general.FCSError{
					Code:    general.CodeUnsupportedParameterValue,
					Ident:   SearchRetrArgFCSContext.String(),
					Message: "Unknown context " + v,
				}
				return http.StatusBadRequest
			}
		}

	} else {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   SearchRetrArgFCSContext.String(),
			Message: "Empty context",
		}
		return http.StatusBadRequest
	}
	retrieveAttrs := a.corporaConf.Resources.GetCommonPosAttrNames(corpora...)

	queryType := getTypedArg[QueryType](ctx, "queryType", DefaultQueryType)
	fcsResponse.SearchRetrieve.QueryType = queryType

	// make searches
	waits := make([]<-chan *rdb.WorkerResult, len(corpora))
	for i, corpusName := range corpora {

		ast, fcsErr := a.translateQuery(corpusName, fcsQuery, queryType)
		if fcsErr != nil {
			fcsResponse.General.Error = fcsErr
			return http.StatusInternalServerError
		}

		query := ast.Generate()
		if len(ast.Errors()) > 0 {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeQueryCannotProcess,
				Ident:   SearchRetrArgQuery.String(),
				Message: ast.Errors()[0].Error(),
			}
			return http.StatusInternalServerError
		}
		args, err := json.Marshal(rdb.ConcExampleArgs{
			CorpusPath: a.corporaConf.GetRegistryPath(corpusName),
			QueryLemma: "",
			Query:      query,
			Attrs:      retrieveAttrs,
			MaxItems:   10,
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
	commonLayers := a.corporaConf.Resources.GetCommonLayers()
	commonPosAttrs := a.corporaConf.Resources.GetCommonPosAttrs(corpora...)
	for i, r := range results {
		for _, l := range r.Lines {
			segmentPos := 1
			row := FCSSearchRow{
				LayerAttrs: a.corporaConf.Resources[corpora[i]].GetDefinedLayers().ToOrderedSlice(),
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
					Layers: a.exportAttrsByLayers(
						t.Word,
						t.Attrs,
						commonLayers,
						commonPosAttrs,
					),
				}
				segmentPos += len(t.Word) + 1 // with space between words
				row.Tokens = append(row.Tokens, token)

			}
			fcsResponse.SearchRetrieve.Results = append(fcsResponse.SearchRetrieve.Results, row)
		}
	}
	return http.StatusOK
}
