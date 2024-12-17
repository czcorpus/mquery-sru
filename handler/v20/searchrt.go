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
	"strings"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/logging"
	"github.com/czcorpus/mquery-common/concordance"
	"github.com/czcorpus/mquery-sru/backlink"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/handler/v20/schema"
	"github.com/czcorpus/mquery-sru/mango"
	"github.com/czcorpus/mquery-sru/query"
	"github.com/czcorpus/mquery-sru/query/compiler"
	"github.com/czcorpus/mquery-sru/query/parser/basic"
	"github.com/czcorpus/mquery-sru/query/parser/fcsql"
	"github.com/czcorpus/mquery-sru/rdb"
	"github.com/czcorpus/mquery-sru/result"
	"github.com/rs/zerolog/log"

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

func (a *FCSSubHandlerV20) getAttrByLayers(
	commonPosAttrs []corpus.PosAttr,
	layer corpus.LayerType,
	token concordance.Token,
) string {
	for _, posAttr := range commonPosAttrs {
		if posAttr.Layer == layer {
			if v, ok := token.Attrs[posAttr.Name]; ok {
				return v
			}
		}
	}
	return "??"
}

func (a *FCSSubHandlerV20) searchRetrieve(ctx *gin.Context, fcsResponse *FCSRequest) (schema.XMLSRResponse, int) {
	logArgs := make(map[string]interface{})
	logging.AddLogEvent(ctx, "args", logArgs)
	ans := schema.NewXMLSRResponse()
	// check if all parameters are supported
	for key := range ctx.Request.URL.Query() {
		if err := SearchRetrArg(key).Validate(); err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDiagnostic(general.DCUnsupportedParameter, 0, key, err.Error())
			return ans, general.ConformantStatusBadRequest
		}
	}

	// handle query parameter
	fcsQuery := ctx.Query(SearchRetrArgQuery.String())
	if len(fcsQuery) == 0 {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCMandatoryParameterNotSupplied, 0, "fcs_query")
		return ans, general.ConformantStatusBadRequest
	}
	ans.EchoedRequest.Query = fcsQuery
	logArgs[SearchRetrArgQuery.String()] = fcsQuery
	// handle start record parameter
	xStartRecord := ctx.DefaultQuery(SearchRetrStartRecord.String(), "1")
	startRecord, err := strconv.Atoi(xStartRecord)
	if err != nil {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnsupportedParameterValue, 0, SearchRetrStartRecord.String())
		return ans, general.ConformantUnprocessableEntity
	}
	if startRecord < 1 {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnsupportedParameterValue, 0, SearchRetrStartRecord.String())
		return ans, general.ConformantUnprocessableEntity
	}
	ans.EchoedRequest.StartRecord = startRecord
	logArgs[SearchRetrStartRecord.String()] = startRecord

	// handle record schema parameter
	recordSchema := ctx.DefaultQuery(SearchRetrArgRecordSchema.String(), general.RecordSchema)
	if recordSchema != general.RecordSchema {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnknownSchemaForRetrieval, 0, SearchMaximumRecords.String())
		return ans, general.ConformantUnprocessableEntity
	}

	// handle max records parameter
	maximumRecords := a.corporaConf.MaximumRecords
	if xMaximumRecords := ctx.Query(SearchMaximumRecords.String()); len(xMaximumRecords) > 0 {
		maximumRecords, err = strconv.Atoi(xMaximumRecords)
		if err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDfltMsgDiagnostic(
				general.DCUnsupportedParameterValue, 0, SearchMaximumRecords.String())
			return ans, general.ConformantUnprocessableEntity
		}
	}
	if maximumRecords < 1 {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnsupportedParameterValue, 0, SearchMaximumRecords.String())
		return ans, general.ConformantUnprocessableEntity

	}
	if maximumRecords > mango.MaxRecordsInternalLimit {
		// TODO the error type is not probably very accurate
		// as the actual result can be very small. But we still
		// have to limit max. number of records...
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCTooManyMatchingRecords, 0, fmt.Sprintf("%d", mango.MaxRecordsInternalLimit))
		return ans, general.ConformantUnprocessableEntity
	}
	logArgs[SearchMaximumRecords.String()] = maximumRecords

	// handle requested sources
	corporaPids := fetchContext(ctx)
	corpora := make([]string, 0, len(corporaPids))
	if len(corporaPids) > 0 {
		for _, pid := range corporaPids {
			res, err := a.corporaConf.Resources.GetResourceByPID(pid)
			if err == corpus.ErrResourceNotFound {
				ans.Records = nil
				return ans, http.StatusOK
			}
			corpora = append(corpora, res.ID)
		}

	} else {
		corpora = a.corporaConf.Resources.GetCorpora()
	}

	// get searchable corpora and attrs
	if len(corpora) == 0 {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCUnsupportedContextSet, 0, SearchRetrArgFCSContext.String())
		return ans, general.ConformantStatusBadRequest
	}
	retrieveAttrs, err := a.corporaConf.Resources.GetCommonPosAttrNames(corpora...)
	if err != nil {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCGeneralSystemError, 0, err.Error())
		return ans, http.StatusInternalServerError
	}
	// add text layer as another attr, otherwise we won't be able to parse it due to Manatee output formatting
	retrieveAttrs = append(retrieveAttrs, retrieveAttrs[0])

	logArgs["corpus"] = a.serverInfo.Database
	logArgs["sources"] = corpora
	logArgs[SearchRetrArgFCSContext.String()] = ctx.Query(SearchRetrArgFCSContext.String())
	log.Warn().Msg("Data views are not implemented yet!")
	logArgs[SearchRetrArgFCSDataViews.String()] = ctx.Query(SearchRetrArgFCSDataViews.String())

	queryType := getTypedArg[QueryType](ctx, SearchRetrArgQueryType.String(), DefaultQueryType)
	logArgs[SearchRetrArgQueryType.String()] = queryType

	ranges := query.CalculatePartialRanges(corpora, startRecord-1, maximumRecords)

	// make searches
	waits := make([]<-chan result.ConcResult, len(ranges))
	for i, rng := range ranges {

		ast, fcsErr := a.translateQuery(rng.Rsc, fcsQuery, queryType)
		if fcsErr != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDiagnostic(fcsErr.Code, fcsErr.Type, fcsErr.Ident, fcsErr.Message)
			return ans, general.ConformantUnprocessableEntity
		}

		query := ast.Generate()
		if len(ast.Errors()) > 0 {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDiagnostic(
				general.DCQueryCannotProcess, 0, SearchRetrArgQuery.String(), ast.Errors()[0].Error())
			return ans, general.ConformantUnprocessableEntity
		}
		rscConf, err := a.corporaConf.Resources.GetResource(rng.Rsc)
		if err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDfltMsgDiagnostic(
				general.DCGeneralSystemError, 0, err.Error())
			return ans, general.ConformandGeneralServerError
		}
		wait, err := a.radapter.PublishQuery(rdb.Query{
			Func: "concExample",
			Args: rdb.ConcQueryArgs{
				CorpusPath:        a.corporaConf.GetRegistryPath(rng.Rsc),
				Query:             query,
				Attrs:             retrieveAttrs,
				StartLine:         rng.From,
				MaxItems:          maximumRecords,
				MaxContext:        a.corporaConf.MaximumContext,
				ViewContextStruct: rscConf.ViewContextStruct,
			},
		})
		if err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDfltMsgDiagnostic(
				general.DCGeneralSystemError, 0, err.Error())
			return ans, http.StatusInternalServerError
		}
		waits[i] = wait
	}
	// using fromResource, we will cycle through available resources' results and their lines
	fromResource := result.NewRoundRobinLineSel(maximumRecords, ranges.PIDList()...)
	usedQueries := make(map[string]string) // maps resource ID to Manatee CQL query
	var totalConcSize int
	for i, wait := range waits {
		result := <-wait
		if result.Error == mango.ErrRowsRangeOutOfConc {
			fromResource.RscSetErrorAt(i, err)

		} else if result.Error != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDfltMsgDiagnostic(
				general.DCQueryCannotProcess, 0, result.Error.Error())
			return ans, http.StatusInternalServerError
		}
		fromResource.SetRscLines(ranges[i].Rsc, result)
		usedQueries[ranges[i].Rsc] = result.Query
		totalConcSize += result.ConcSize
	}

	ans.NumberOfRecords = totalConcSize
	if fromResource.AllHasOutOfRangeError() {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCFirstRecordPosOutOfRange, 0, fromResource.GetFirstError().Error())
		return ans, general.ConformantUnprocessableEntity

	} else if fromResource.HasFatalError() {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCQueryCannotProcess, 0, fromResource.GetFirstError().Error())
		return ans, general.ConformandGeneralServerError
	}

	// transform results
	commonLayers := a.corporaConf.Resources.GetCommonLayers()
	commonPosAttrs, err := a.corporaConf.Resources.GetCommonPosAttrs(corpora...)
	if err != nil {
		ans.Diagnostics = schema.NewXMLDiagnostics()
		ans.Diagnostics.AddDfltMsgDiagnostic(
			general.DCGeneralSystemError, 0, err.Error())
		return ans, http.StatusInternalServerError
	}

	records := make([]schema.XMLSRRecord, 0, maximumRecords)
	for len(records) < maximumRecords && fromResource.Next() {
		res, err := a.corporaConf.Resources.GetResource(fromResource.CurrRscName())
		if err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDfltMsgDiagnostic(
				general.DCGeneralSystemError, 0, err.Error())
			return ans, http.StatusInternalServerError
		}
		item := fromResource.CurrLine()
		var refURL string
		if res.KontextBacklinkRootURL != "" {
			var err error
			refURL, err = backlink.GenerateForKonText(
				res.KontextBacklinkRootURL, res.ID, usedQueries[res.ID], item.Ref)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate ResourceFragment URL")
			}
		}
		segmentPos := 1
		records = append(records, schema.XMLSRRecord{
			Schema:      "http://clarin.eu/fcs/resource",
			XMLEscaping: string(fcsResponse.RecordXMLEscaping),
			Data: schema.XMLSRResource{
				XMLNSFCS: "http://clarin.eu/fcs/resource",
				PID:      res.PID,
				ResourceFragment: schema.XMLSRResourceFragment{
					Ref: refURL,
					DataViews: []*schema.XMLSRDataView{
						// basic data view
						{
							Type: "application/x-clarin-fcs-hits+xml",
							Result: schema.XMLSRBasicDataViewResult{
								XMLNSHits: "http://clarin.eu/fcs/dataview/hits",
								Data: strings.Join(
									collections.SliceMap(
										item.Text.Tokens(),
										func(token *concordance.Token, i int) string {
											if token.Strong {
												return "<hits:Hit>" + token.Word + "</hits:Hit>"
											}
											return token.Word
										},
									),
									" ",
								),
							},
						},
						// advanced data view if requested
						general.ReturnIf(
							queryType == QueryTypeFCS,
							&schema.XMLSRDataView{
								Type: "application/x-clarin-fcs-adv+xml",
								Result: schema.XMLSRAdvancedDataViewResult{
									Unit:     "item",
									XMLNSAdv: "http://clarin.eu/fcs/dataview/advanced",
									Segments: collections.SliceMap(
										item.Text.Tokens(),
										func(token *concordance.Token, i int) schema.XMLSRAdvSegment {
											segment := schema.XMLSRAdvSegment{
												ID:    fmt.Sprintf("s%d", i),
												Start: segmentPos,
												End:   segmentPos + len(token.Word) - 1,
											}
											segmentPos += len(token.Word) + 1 // with space between words
											return segment
										},
									),
									Layers: collections.SliceMap(
										commonLayers,
										func(layer corpus.LayerType, j int) schema.XMLSRAdvLayer {
											return schema.XMLSRAdvLayer{
												ID: layer.GetResultID(),
												Values: collections.SliceMap(
													item.Text.Tokens(),
													func(token *concordance.Token, i int) schema.XMLSRAdvValue {
														return schema.XMLSRAdvValue{
															Ref:       fmt.Sprintf("s%d", i),
															Highlight: general.ReturnIf(token.Strong, fmt.Sprintf("s%d", i), ""),
															Value:     a.getAttrByLayers(commonPosAttrs, layer, *token),
														}
													},
												),
											}
										},
									),
								},
							},
							nil,
						),
					},
				},
			},
			RecordPosition: len(records) + startRecord,
		})
	}
	if len(records) > 0 {
		ans.Records = &records
	}
	if len(records)+startRecord-1 < ans.NumberOfRecords {
		ans.NextRecordPosition = len(records) + startRecord
	}
	return ans, http.StatusOK
}
