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
	"encoding/json"
	"fcs/general"
	"fcs/query/compiler"
	"fcs/query/parser/simple"
	"fcs/rdb"
	"fcs/results"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV12) translateQuery(
	corpusName, query string,
) (compiler.AST, *general.FCSError) {
	var fcsErr *general.FCSError
	ast, err := simple.ParseQuery(
		query,
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
	return ast, fcsErr
}

func (a *FCSSubHandlerV12) searchRetrieve(ctx *gin.Context, fcsResponse *FCSResponse) int {
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

	// make searches
	waits := make([]<-chan *rdb.WorkerResult, len(corpora))
	for i, corpusName := range corpora {

		ast, fcsErr := a.translateQuery(corpusName, fcsQuery)
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
	for i, r := range results {
		for _, l := range r.Lines {
			row := FCSSearchRow{
				Position: len(fcsResponse.SearchRetrieve.Results) + 1,
				PID:      corpora[i],
				Web:      "TODO",
				Ref:      "TODO",
			}
			for _, t := range l.Text {
				token := Token{
					Text: t.Word,
					Hit:  t.Strong,
				}
				row.Tokens = append(row.Tokens, token)

			}
			fcsResponse.SearchRetrieve.Results = append(fcsResponse.SearchRetrieve.Results, row)
		}
	}
	return http.StatusOK
}
