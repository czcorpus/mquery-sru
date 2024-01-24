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
	"fcs/mango"
	"fcs/query/compiler"
	"fcs/query/parser/basic"
	"fcs/query/parser/fcsql"
	"fcs/rdb"
	"fcs/results"
	"fmt"
	"net/http"
	"strconv"
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
		ast, err = basic.ParseQuery(
			query,
			a.corporaConf.Resources[corpusName].PosAttrs,
			a.corporaConf.Resources[corpusName].StructureMapping,
		)
		if err != nil {
			fcsErr = &general.FCSError{
				Code:    general.CodeQuerySyntaxError,
				Ident:   query,
				Message: fmt.Sprintf("Invalid query syntax: %s", err),
			}
		}
	case QueryTypeFCS:
		var err error
		ast, err = fcsql.ParseQuery(
			query,
			a.corporaConf.Resources[corpusName].PosAttrs,
			a.corporaConf.Resources[corpusName].StructureMapping,
		)
		if err != nil {
			fcsErr = &general.FCSError{
				Code:    general.CodeQuerySyntaxError,
				Ident:   query,
				Message: fmt.Sprintf("Invalid query syntax: %s", err),
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
			return general.ConformantStatusBadRequest
		}
	}

	fcsQuery := ctx.Query("query")
	if len(fcsQuery) == 0 {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeMandatoryParameterNotSupplied,
			Ident:   "fcs_query",
			Message: "Mandatory parameter not supplied",
		}
		return general.ConformantStatusBadRequest
	}
	fcsResponse.SearchRetrieve.EchoedSRRequest.Query = fcsQuery

	xStartRecord := ctx.DefaultQuery(SearchRetrStartRecord.String(), "1")
	startRecord, err := strconv.Atoi(xStartRecord)
	if err != nil {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   SearchRetrStartRecord.String(),
			Message: "Invalid parameter value",
		}
		return general.ConformantUnprocessableEntity
	}
	if startRecord < 1 {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   SearchRetrStartRecord.String(),
			Message: "Invalid parameter value",
		}
		return general.ConformantUnprocessableEntity
	}
	fcsResponse.SearchRetrieve.EchoedSRRequest.StartRecord = startRecord

	xMaximumRecords := ctx.DefaultQuery(SearchMaximumRecords.String(), "100")
	maximumRecords, err := strconv.Atoi(xMaximumRecords)
	if err != nil {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   SearchMaximumRecords.String(),
			Message: "Invalid parameter value",
		}
		return general.ConformantUnprocessableEntity
	}
	if maximumRecords < 1 {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   SearchMaximumRecords.String(),
			Message: "Invalid parameter value",
		}
		return general.ConformantUnprocessableEntity
	}

	corpora := strings.Split(ctx.DefaultQuery(SearchRetrArgFCSContext.String(), ""), ",")
	if len(corpora) == 0 || len(corpora) == 1 && corpora[0] == "" {
		corpora = a.corporaConf.Resources.GetCorpora()
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
				return general.ConformantUnprocessableEntity
			}
		}

	} else {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeUnsupportedParameterValue,
			Ident:   SearchRetrArgFCSContext.String(),
			Message: "Empty context",
		}
		return general.ConformantStatusBadRequest
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
			return general.ConformantUnprocessableEntity
		}

		query := ast.Generate()
		if len(ast.Errors()) > 0 {
			fcsResponse.General.Error = &general.FCSError{
				Code:    general.CodeQueryCannotProcess,
				Ident:   SearchRetrArgQuery.String(),
				Message: ast.Errors()[0].Error(),
			}
			return general.ConformantUnprocessableEntity
		}
		args, err := json.Marshal(rdb.ConcExampleArgs{
			CorpusPath: a.corporaConf.GetRegistryPath(corpusName),
			Query:      query,
			Attrs:      retrieveAttrs,
			StartLine:  startRecord - 1,
			MaxItems:   maximumRecords,
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

	// using fromResource, we will cycle through available resources' results and their lines
	fromResource := results.NewRoundRobinLineSel(corpora...)

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
			if err.Error() == mango.ErrRowsRangeOutOfConc.Error() {
				fromResource.RscSetErrorAt(i, err)

			} else {
				fcsResponse.General.Error = &general.FCSError{
					Code:    general.CodeGeneralSystemError,
					Ident:   err.Error(),
					Message: "General system error",
				}
				return http.StatusInternalServerError
			}
		}
		fromResource.SetRscLines(corpora[i], result)
	}

	if fromResource.HasFatalError() {
		fcsResponse.General.Error = &general.FCSError{
			Code:    general.CodeGeneralSystemError,
			Ident:   fromResource.GetFirstError().Error(),
			Message: "General system error",
		}
		return general.ConformantUnprocessableEntity // TODO how can we infer the code from the error?
	}

	// transform results
	fcsResponse.SearchRetrieve.Results = make([]FCSSearchRow, 0, maximumRecords)
	commonLayers := a.corporaConf.Resources.GetCommonLayers()
	commonPosAttrs := a.corporaConf.Resources.GetCommonPosAttrs(corpora...)

	for len(fcsResponse.SearchRetrieve.Results) < maximumRecords && fromResource.Next() {
		segmentPos := 1
		row := FCSSearchRow{
			LayerAttrs: a.corporaConf.Resources[fromResource.CurrRscName()].GetDefinedLayers().ToOrderedSlice(),
			Position:   len(fcsResponse.SearchRetrieve.Results) + 1,
			PID:        fromResource.CurrRscName(),
			Web:        "TODO",
			Ref:        "TODO",
		}
		item := fromResource.CurrLine()
		for j, t := range item.Text {
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
	return http.StatusOK
}
