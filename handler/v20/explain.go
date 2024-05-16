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
	"encoding/xml"
	"net/http"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/mquery-sru/corpus"
	"github.com/czcorpus/mquery-sru/general"
	"github.com/czcorpus/mquery-sru/handler/v20/schema"

	"github.com/gin-gonic/gin"
)

func (a *FCSSubHandlerV20) explain(ctx *gin.Context, fcsResponse *FCSResponse) (schema.XMLExplainResponse, int) {
	ans := schema.XMLExplainResponse{
		XMLNSSRUResponse: "http://docs.oasis-open.org/ns/search-ws/sruResponse",
		Version:          "2.0",
		ExplainRecord: &schema.XMLExplainRecord{
			Schema:      "http://explain.z3950.org/dtd/2.0/",
			XMLEscaping: string(fcsResponse.RecordXMLEscaping),
			Data: schema.XMLExplainData{
				XMLNSZR: "http://explain.z3950.org/dtd/2.0/",
				ServerInfo: schema.XMLExplainServerInfo{
					Protocol:  "SRU",
					Version:   "2.0",
					Transport: "http",
					Host:      a.serverInfo.ServerHost,
					Port:      a.serverInfo.ServerPort,
					Database:  a.serverInfo.Database,
				},
				DatabaseInfo: schema.XMLExplainDatabaseInfo{
					Titles: general.MapItems(
						a.serverInfo.DatabaseTitle,
						func(k string, v string) schema.XMLMultilingual {
							return schema.XMLMultilingual{Language: k, Primary: a.serverInfo.PrimaryLanguage == k, Value: v}
						},
					),
					Descriptions: general.MapItems(
						a.serverInfo.DatabaseDescription,
						func(k string, v string) schema.XMLMultilingual {
							return schema.XMLMultilingual{Language: k, Primary: a.serverInfo.PrimaryLanguage == k, Value: v}
						},
					),
					Authors: general.MapItems(
						a.serverInfo.DatabaseAuthor,
						func(k string, v string) schema.XMLMultilingual {
							return schema.XMLMultilingual{Language: k, Primary: a.serverInfo.PrimaryLanguage == k, Value: v}
						},
					),
				},
				IndexInfo: schema.XMLExplainIndexInfo{
					Set: schema.XMLExplainDefinition{
						Identifier: "http://clarin.eu/fcs/resource",
						Name:       "fcs",
						Titles: []schema.XMLMultilingual{
							{Language: "se", Value: "Clarins innehållssökning"},
							{Language: "en", Value: "CLARIN Content Search", Primary: true},
						},
					},
					Index: schema.XMLExplainIndexInfoIndex{
						Search: true, Scan: false, Sort: false,
						Titles: []schema.XMLMultilingual{
							{Language: "en", Value: "Words", Primary: true},
						},
						Maps: []schema.XMLExplainIndexInfoIndexMap{
							{Primary: true, Name: schema.XMLExplainIndexInfoIndexMapName{Set: "fcs", Value: "words"}},
						},
					},
				},
				SchemaInfo: schema.XMLExplainSchemaInfo{
					Schema: schema.XMLExplainDefinition{
						Identifier: "http://clarin.eu/fcs/resource",
						Name:       "fcs",
						Titles: []schema.XMLMultilingual{
							{Language: "en", Value: "CLARIN Federated Content Search", Primary: true},
						},
					},
				},
				ConfigInfo: schema.XMLExplainConfigInfo{Values: []schema.XMLExplainConfig{
					schema.XMLExplainConfig{
						XMLName: xml.Name{Local: "zr:default"},
						Type:    "numberOfRecords",
						Value:   corpus.ExplainOpNumberOfRecords,
					},
					schema.XMLExplainConfig{
						XMLName: xml.Name{Local: "zr:setting"},
						Type:    "maximumRecords",
						Value:   a.corporaConf.MaximumRecords,
					},
				}},
			},
		},
		EchoedRequest: &schema.XMLExplainEchoedRequest{
			Version: "2.0",
		},
	}

	// check if all parameters are supported
	for key, _ := range ctx.Request.URL.Query() {
		if err := ExplainArg(key).Validate(); err != nil {
			ans.Diagnostics = schema.NewXMLDiagnostics()
			ans.Diagnostics.AddDiagnostic(general.DCUnsupportedParameter, 0, key, err.Error())
			return ans, general.ConformantStatusBadRequest
		}
	}

	// extra data
	if ctx.Query(ExplainArgFCSEndpointDescription.String()) == "true" {
		ans.EndpointDescription = &schema.XMLExplainEndpointDescription{
			XMLNSED: "http://clarin.eu/fcs/endpoint-description",
			Version: "2",

			Capabilities: []string{
				"http://clarin.eu/fcs/capability/basic-search",
				"http://clarin.eu/fcs/capability/advanced-search",
			},
			SupportedDataViews: []schema.XMLExplainSupportedDataView{
				{ID: "hits", DeliveryPolicy: "send-by-default", Value: "application/x-clarin-fcs-hits+xml"},
				{ID: "adv", DeliveryPolicy: "send-by-default", Value: "application/x-clarin-fcs-adv+xml"},
			},
			SupportedLayers: collections.SliceMap(
				a.corporaConf.Resources.GetCommonPosAttrs2(),
				func(posAttr corpus.PosAttr, i int) schema.XMLExplainSupportedLayer {
					return schema.XMLExplainSupportedLayer{
						ID:        posAttr.ID,
						Qualifier: posAttr.Name,
						ResultID:  posAttr.Layer.GetResultID(),
						Value:     string(posAttr.Layer),
					}
				},
			),
			Resources: collections.SliceMap(
				a.corporaConf.Resources,
				func(corpusConf *corpus.CorpusSetup, i int) schema.XMLExplainResource {
					return schema.XMLExplainResource{
						PID:             corpusConf.PID,
						LandingPage:     corpusConf.URI,
						Languages:       corpusConf.Languages,
						AvailableLayers: schema.XMLExplainAvailableValues{Values: corpusConf.GetDefinedLayersAsRefString()},
						Titles: general.MapItems(
							corpusConf.FullName, func(lang, title string) schema.XMLMultilingual2 {
								return schema.XMLMultilingual2{Language: lang, Value: title}
							},
						),
						Descriptions: general.MapItems(
							corpusConf.Description, func(lang, title string) schema.XMLMultilingual2 {
								return schema.XMLMultilingual2{Language: lang, Value: title}
							},
						),
					}
				},
			),
		}
	}
	return ans, http.StatusOK
}
