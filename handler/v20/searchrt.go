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
	"fmt"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/mango"
	"github.com/czcorpus/mquery-sru/query"
	"github.com/czcorpus/mquery-sru/query/compiler"
	"github.com/czcorpus/mquery-sru/query/parser/basic"
	"github.com/czcorpus/mquery-sru/query/parser/fcsql"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/czcorpus/mquery-sru/result"

	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV20) translateQuery(
	corpusName, query string,
	queryType QueryType,
) (compiler.AST, *general.FCSError) {
	var ast compiler.AST
	var fcsErr *general.FCSError
	res, err := a.corporaConf.Resources.GetResource(corpusName)
	if err != nil {
		fcsErr = &general.FCSError{
			Code:    general.DCGeneralSystemError,
			Ident:   err.Error(),
			Message: general.DCGeneralSystemError.AsMessage(),
		}
		return nil, fcsErr
	}
	switch queryType {
	case QueryTypeCQL:
		var err error
		ast, err = basic.ParseQuery(
			query,
			res.PosAttrs,
			res.StructureMapping,
		)
		if err != nil {
			fcsErr = &general.FCSError{
				Code:    general.DCQuerySyntaxError,
				Ident:   query,
				Message: fmt.Sprintf("Invalid query syntax: %s", err),
			}
		}
	case QueryTypeFCS:
		var err error
		ast, err = fcsql.ParseQuery(
			query,
			res.PosAttrs,
			res.StructureMapping,
		)
		if err != nil {
			fcsErr = &general.FCSError{
				Code:    general.DCQuerySyntaxError,
				Ident:   query,
				Message: fmt.Sprintf("Invalid query syntax: %s", err),
			}
		}

	default:
		fcsErr = &general.FCSError{
			Code:    general.DCUnsupportedParameterValue,
			Ident:   queryType.String(),
			Message: general.DCUnsupportedParameterValue.AsMessage(),
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
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCUnsupportedParameter,
				Ident:   key,
				Message: err.Error(),
			})
			return general.ConformantStatusBadRequest
		}
	}

	fcsQuery := ctx.Query("query")
	if len(fcsQuery) == 0 {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCMandatoryParameterNotSupplied,
			Ident:   "fcs_query",
			Message: general.DCMandatoryParameterNotSupplied.AsMessage(),
		})
		return general.ConformantStatusBadRequest
	}
	fcsResponse.SearchRetrieve.EchoedSRRequest.Query = fcsQuery

	xStartRecord := ctx.DefaultQuery(SearchRetrStartRecord.String(), "1")
	startRecord, err := strconv.Atoi(xStartRecord)
	if err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedParameterValue,
			Ident:   SearchRetrStartRecord.String(),
			Message: general.DCUnsupportedParameterValue.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}
	if startRecord < 1 {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedParameterValue,
			Ident:   SearchRetrStartRecord.String(),
			Message: general.DCUnsupportedParameterValue.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}
	fcsResponse.SearchRetrieve.EchoedSRRequest.StartRecord = startRecord

	recordSchema := ctx.DefaultQuery(SearchRetrArgRecordSchema.String(), general.RecordSchema)
	if recordSchema != general.RecordSchema {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnknownSchemaForRetrieval,
			Ident:   SearchMaximumRecords.String(),
			Message: general.DCUnknownSchemaForRetrieval.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}

	maximumRecords := a.corporaConf.MaximumRecords
	if xMaximumRecords := ctx.Query(SearchMaximumRecords.String()); len(xMaximumRecords) > 0 {
		maximumRecords, err = strconv.Atoi(xMaximumRecords)
		if err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCUnsupportedParameterValue,
				Ident:   SearchMaximumRecords.String(),
				Message: general.DCUnsupportedParameterValue.AsMessage(),
			})
			return general.ConformantUnprocessableEntity
		}
	}
	if maximumRecords < 1 {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedParameterValue,
			Ident:   SearchMaximumRecords.String(),
			Message: general.DCUnsupportedParameterValue.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}
	if maximumRecords > mango.MaxRecordsInternalLimit {
		// TODO the error type is not probably very accurate
		// as the actual result can be very small. But we still
		// have to limit max. number of records...
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCTooManyMatchingRecords,
			Ident:   fmt.Sprintf("%d", mango.MaxRecordsInternalLimit),
			Message: general.DCTooManyMatchingRecords.AsMessage(),
		})
		return general.ConformantUnprocessableEntity
	}

	corporaPids := fetchContext(ctx)
	corpora := make([]string, 0, len(corporaPids))
	if len(corporaPids) > 0 {
		for _, pid := range corporaPids {
			res, err := a.corporaConf.Resources.GetResourceByPID(pid)
			if err != corpus.ErrResourceNotFound {
				fcsResponse.SearchRetrieve.Results = []FCSSearchRow{}
				return http.StatusOK
			}
			corpora = append(corpora, res.ID)
		}

	} else {
		corpora = a.corporaConf.Resources.GetCorpora()
	}

	// get searchable corpora and attrs
	if len(corpora) == 0 {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCUnsupportedContextSet,
			Ident:   SearchRetrArgFCSContext.String(),
			Message: general.DCUnsupportedContextSet.AsMessage(),
		})
		return general.ConformantStatusBadRequest
	}

	retrieveAttrs, err := a.corporaConf.Resources.GetCommonPosAttrNames(corpora...)
	if err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCGeneralSystemError,
			Ident:   err.Error(),
			Message: general.DCGeneralSystemError.AsMessage(),
		})
		return http.StatusInternalServerError
	}

	queryType := getTypedArg[QueryType](ctx, "queryType", DefaultQueryType)
	fcsResponse.SearchRetrieve.QueryType = queryType
	ranges := query.CalculatePartialRanges(corpora, startRecord-1, maximumRecords)

	// make searches
	waits := make([]<-chan *rdb.WorkerResult, len(corpora))
	for i, rng := range ranges {

		ast, fcsErr := a.translateQuery(rng.Rsc, fcsQuery, queryType)
		if fcsErr != nil {
			fcsResponse.General.AddError(*fcsErr)
			return general.ConformantUnprocessableEntity
		}

		query := ast.Generate()
		if len(ast.Errors()) > 0 {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCQueryCannotProcess,
				Ident:   SearchRetrArgQuery.String(),
				Message: ast.Errors()[0].Error(),
			})
			return general.ConformantUnprocessableEntity
		}
		args, err := sonic.Marshal(rdb.ConcExampleArgs{
			CorpusPath: a.corporaConf.GetRegistryPath(rng.Rsc),
			Query:      query,
			Attrs:      retrieveAttrs,
			StartLine:  rng.From,
			MaxItems:   maximumRecords,
			MaxContext: a.corporaConf.MaximumContext,
		})
		if err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCGeneralSystemError,
				Ident:   err.Error(),
				Message: general.DCGeneralSystemError.AsMessage(),
			})
			return http.StatusInternalServerError
		}
		wait, err := a.radapter.PublishQuery(rdb.Query{
			Func: "concExample",
			Args: args,
		})
		if err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCGeneralSystemError,
				Ident:   err.Error(),
				Message: general.DCGeneralSystemError.AsMessage(),
			})
			return http.StatusInternalServerError
		}
		waits[i] = wait
	}

	// using fromResource, we will cycle through available resources' results and their lines
	fromResource := result.NewRoundRobinLineSel(maximumRecords, corpora...)
	var totalConcSize int
	for i, wait := range waits {
		rawResult := <-wait
		result, err := rdb.DeserializeConcExampleResult(rawResult)
		if err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCGeneralSystemError,
				Ident:   err.Error(),
				Message: general.DCGeneralSystemError.AsMessage(),
			})
			return http.StatusInternalServerError
		}
		if err := result.Err(); err != nil {
			if err.Error() == mango.ErrRowsRangeOutOfConc.Error() {
				fromResource.RscSetErrorAt(i, err)

			} else {
				fcsResponse.General.AddError(general.FCSError{
					Code:    general.DCQueryCannotProcess,
					Ident:   err.Error(),
					Message: general.DCQueryCannotProcess.AsMessage(),
				})
				return http.StatusInternalServerError
			}
		}
		fromResource.SetRscLines(corpora[i], result)
		totalConcSize += result.ConcSize
	}
	fcsResponse.SearchRetrieve.NumberOfRecords = totalConcSize
	if fromResource.AllHasOutOfRangeError() {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCFirstRecordPosOutOfRange,
			Ident:   fromResource.GetFirstError().Error(),
			Message: general.DCFirstRecordPosOutOfRange.AsMessage(),
		})
		return general.ConformantUnprocessableEntity

	} else if fromResource.HasFatalError() {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCQueryCannotProcess,
			Ident:   fromResource.GetFirstError().Error(),
			Message: general.DCQueryCannotProcess.AsMessage(),
		})
		return general.ConformandGeneralServerError
	}

	// transform results
	fcsResponse.SearchRetrieve.Results = make([]FCSSearchRow, 0, maximumRecords)
	commonLayers := a.corporaConf.Resources.GetCommonLayers()
	commonPosAttrs, err := a.corporaConf.Resources.GetCommonPosAttrs(corpora...)
	if err != nil {
		fcsResponse.General.AddError(general.FCSError{
			Code:    general.DCGeneralSystemError,
			Ident:   err.Error(),
			Message: general.DCGeneralSystemError.AsMessage(),
		})
		return http.StatusInternalServerError
	}

	for len(fcsResponse.SearchRetrieve.Results) < maximumRecords && fromResource.Next() {
		segmentPos := 1
		res, err := a.corporaConf.Resources.GetResource(fromResource.CurrRscName())
		if err != nil {
			fcsResponse.General.AddError(general.FCSError{
				Code:    general.DCGeneralSystemError,
				Ident:   err.Error(),
				Message: general.DCGeneralSystemError.AsMessage(),
			})
			return http.StatusInternalServerError
		}
		row := FCSSearchRow{
			LayerAttrs: res.GetDefinedLayers().ToOrderedSlice(),
			Position:   len(fcsResponse.SearchRetrieve.Results) + startRecord,
			PID:        fromResource.CurrRscName(),
			Ref:        res.URI,
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
