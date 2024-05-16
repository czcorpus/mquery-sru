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

package schema

import "encoding/xml"

type XMLExplainResponse struct {
	XMLName          xml.Name `xml:"sruResponse:explainResponse"`
	XMLNSSRUResponse string   `xml:"xmlns:sruResponse,attr"`
	Version          string   `xml:"sruResponse:version"`

	ExplainRecord       *XMLExplainRecord              `xml:"sruResponse:record,omitempty"`
	EchoedRequest       *XMLExplainEchoedRequest       `xml:"sruResponse:echoedExplainRequest,omitempty"`
	EndpointDescription *XMLExplainEndpointDescription `xml:"sruResponse:extraResponseData>ed:EndpointDescription,omitempty"`
	Diagnostics         *XMLDiagnostics                `xml:"sruResponse:diagnostics,omitempty"`
}

// --------------------- Explain Record ---------------------

type XMLExplainRecord struct {
	Schema      string         `xml:"sruResponse:recordSchema"`
	XMLEscaping string         `xml:"sruResponse:recordXMLEscaping"`
	Data        XMLExplainData `xml:"sruResponse:recordData>zr:explain"`
}

type XMLExplainData struct {
	XMLNSZR string `xml:"xmlns:zr,attr"`

	ServerInfo   XMLExplainServerInfo   `xml:"zr:serverInfo"`
	DatabaseInfo XMLExplainDatabaseInfo `xml:"zr:databaseInfo"`
	IndexInfo    XMLExplainIndexInfo    `xml:"zr:indexInfo"`
	SchemaInfo   XMLExplainSchemaInfo   `xml:"zr:schemaInfo"`
	ConfigInfo   XMLExplainConfigInfo   `xml:"zr:configInfo"`
}

type XMLExplainServerInfo struct {
	Protocol  string `xml:"protocol,attr"`
	Version   string `xml:"version,attr"`
	Transport string `xml:"transport,attr"`

	Host     string `xml:"zr:host"`
	Port     string `xml:"zr:port"`
	Database string `xml:"zr:database"`
}

type XMLExplainDatabaseInfo struct {
	Titles       []XMLMultilingual `xml:"zr:title"`
	Descriptions []XMLMultilingual `xml:"zr:description"`
	Authors      []XMLMultilingual `xml:"zr:author"`
}

type XMLExplainIndexInfo struct {
	Set   XMLExplainDefinition     `xml:"zr:set"`
	Index XMLExplainIndexInfoIndex `xml:"zr:index"`
}

type XMLExplainDefinition struct {
	Identifier string `xml:"identifier,attr"`
	Name       string `xml:"name,attr"`

	Titles []XMLMultilingual `xml:"zr:title"`
}

type XMLExplainIndexInfoIndex struct {
	Search bool `xml:"search,attr"`
	Scan   bool `xml:"scan,attr"`
	Sort   bool `xml:"sort,attr"`

	Titles []XMLMultilingual             `xml:"zr:title"`
	Maps   []XMLExplainIndexInfoIndexMap `xml:"zr:map"`
}

type XMLExplainIndexInfoIndexMap struct {
	Primary bool                            `xml:"primary,attr,omitempty"`
	Name    XMLExplainIndexInfoIndexMapName `xml:"zr:name"`
}

type XMLExplainIndexInfoIndexMapName struct {
	Set   string `xml:"set,attr"`
	Value string `xml:",chardata"`
}

type XMLExplainSchemaInfo struct {
	Schema XMLExplainDefinition `xml:"zr:schema"`
}

type XMLExplainConfigInfo struct {
	Values []XMLExplainConfig
}

func (c *XMLExplainConfigInfo) AddDefault(key string, value any) {
	c.Values = append(c.Values, XMLExplainConfig{
		XMLName: xml.Name{Local: "zr:default"},
		Type:    key,
		Value:   value,
	})
}

func (c *XMLExplainConfigInfo) AddSetting(typ string, value any) {
	c.Values = append(c.Values, XMLExplainConfig{
		XMLName: xml.Name{Local: "zr:setting"},
		Type:    typ,
		Value:   value,
	})
}

type XMLExplainConfig struct {
	XMLName xml.Name
	Type    string `xml:"type,attr"`
	Value   any    `xml:",chardata"`
}

// --------------------- Echoed Explain Request ---------------------

type XMLExplainEchoedRequest struct {
	Version string `xml:"sruResponse:version"`
}

// --------------------- Extra Response Data ---------------------

type XMLExplainEndpointDescription struct {
	XMLNSED string `xml:"xmlns:ed,attr"`
	Version string `xml:"version,attr"`

	Capabilities       []string                      `xml:"ed:Capabilities>ed:Capability"`
	SupportedDataViews []XMLExplainSupportedDataView `xml:"ed:SupportedDataViews>ed:SupportedDataView"`
	SupportedLayers    []XMLExplainSupportedLayer    `xml:"ed:SupportedLayers>ed:SupportedLayer"`
	Resources          []XMLExplainResource          `xml:"ed:Resources>ed:Resource"`
}

type XMLExplainSupportedDataView struct {
	ID             string `xml:"id,attr"`
	DeliveryPolicy string `xml:"delivery-policy,attr"`
	Value          string `xml:",chardata"`
}

type XMLExplainSupportedLayer struct {
	ID        string `xml:"id,attr"`
	Qualifier string `xml:"qualifier,attr"`
	ResultID  string `xml:"result-id,attr"`
	Value     string `xml:",chardata"`
}

type XMLExplainResource struct {
	PID                string                    `xml:"pid,attr"`
	Titles             []XMLMultilingual2        `xml:"ed:Title"`
	Descriptions       []XMLMultilingual2        `xml:"ed:Description"`
	LandingPage        string                    `xml:"ed:LandingPageURI,omitempty"`
	Languages          []string                  `xml:"ed:Languages>ed:Language"`
	AvailableDataViews XMLExplainAvailableValues `xml:"ed:AvailableDataViews"`
	AvailableLayers    XMLExplainAvailableValues `xml:"ed:AvailableLayers"`
}

type XMLExplainAvailableValues struct {
	Values string `xml:"ref,attr"`
}
