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

type XMLSRResponse struct {
	XMLName          xml.Name `xml:"sruResponse:searchRetrieveResponse"`
	XMLNSSRUResponse string   `xml:"xmlns:sruResponse,attr"`
	Version          string   `xml:"sruResponse:version"`

	NumberOfRecords int `xml:"sruResponse:numberOfRecords"`

	// Records
	// note: we need a pointer here to allow the marshaler skip the 'records' parent
	// in case there are no 'record' children
	Records              *[]XMLSRRecord      `xml:"sruResponse:records>sruResponse:record,omitempty"`
	NextRecordPosition   int                 `xml:"sruResponse:nextRecordPosition,omitempty"`
	EchoedRequest        *XMLSREchoedRequest `xml:"sruResponse:echoedSearchRetrieveRequest,omitempty"`
	Diagnostics          *XMLDiagnostics     `xml:"sruResponse:diagnostics,omitempty"`
	ResultCountPrecision string              `xml:"sruResponse:resultCountPrecision"`
}

func NewXMLSRResponse() XMLSRResponse {
	return XMLSRResponse{
		XMLNSSRUResponse:     "http://docs.oasis-open.org/ns/search-ws/sruResponse",
		Version:              "2.0",
		ResultCountPrecision: "info:srw/vocabulary/resultCountPrecision/1/exact",
		EchoedRequest:        &XMLSREchoedRequest{Version: "2.0"},
	}
}

func NewMinimalXMLSRResponse() XMLSRResponse {
	return XMLSRResponse{
		XMLNSSRUResponse:     "http://docs.oasis-open.org/ns/search-ws/sruResponse",
		ResultCountPrecision: "info:srw/vocabulary/resultCountPrecision/1/exact",
		Version:              "2.0",
	}
}

// --------------------- Search Retrieve Record ---------------------

type XMLSRRecord struct {
	Schema         string        `xml:"sruResponse:recordSchema"`
	XMLEscaping    string        `xml:"sruResponse:recordXMLEscaping"`
	Data           XMLSRResource `xml:"sruResponse:recordData>fcs:Resource"`
	RecordPosition int           `xml:"sruResponse:recordPosition"`
}

type XMLSRResource struct {
	XMLNSFCS         string                `xml:"xmlns:fcs,attr"`
	PID              string                `xml:"pid,attr"`
	ResourceFragment XMLSRResourceFragment `xml:"fcs:ResourceFragment"`
}

type XMLSRResourceFragment struct {
	Ref       string           `xml:"ref,attr,omitempty"`
	DataViews []*XMLSRDataView `xml:"fcs:DataView"`
}

type XMLSRDataView struct {
	Type   string `xml:"type,attr"`
	Result any
}

type XMLSRBasicDataViewResult struct {
	XMLName   xml.Name `xml:"hits:Result"`
	XMLNSHits string   `xml:"xmlns:hits,attr"`
	Data      string   `xml:",innerxml"`
}

type XMLSRAdvancedDataViewResult struct {
	XMLName  xml.Name          `xml:"adv:Advanced"`
	Unit     string            `xml:"unit,attr"`
	XMLNSAdv string            `xml:"xmlns:adv,attr"`
	Segments []XMLSRAdvSegment `xml:"adv:Segments>adv:Segment"`
	Layers   []XMLSRAdvLayer   `xml:"adv:Layers>adv:Layer"`
}

type XMLSRAdvSegment struct {
	ID    string `xml:"id,attr"`
	Start int    `xml:"start,attr"`
	End   int    `xml:"end,attr"`
}

type XMLSRAdvLayer struct {
	ID     string          `xml:"id,attr"`
	Values []XMLSRAdvValue `xml:"adv:Span"`
}

type XMLSRAdvValue struct {
	Ref       string `xml:"ref,attr"`
	Highlight string `xml:"highlight,attr,omitempty"`
	Value     string `xml:",chardata"`
}

// --------------------- Echoed Search Retrieve Request ---------------------

type XMLSREchoedRequest struct {
	Version     string `xml:"sruResponse:version"`
	Query       string `xml:"sruResponse:query"`
	StartRecord int    `xml:"sruResponse:startRecord"`
}
