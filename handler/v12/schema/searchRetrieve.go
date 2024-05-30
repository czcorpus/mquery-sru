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
	XMLName          xml.Name `xml:"sru:searchRetrieveResponse"`
	XMLNSSRUResponse string   `xml:"xmlns:sru,attr"`
	Version          string   `xml:"sru:version"`

	NumberOfRecords int                `xml:"sru:numberOfRecords"`
	Records         *[]XMLSRRecord     `xml:"sru:records>sru:record,omitempty"`
	EchoedRequest   XMLSREchoedRequest `xml:"sru:echoedSearchRetrieveRequest"`
	Diagnostics     *XMLDiagnostics    `xml:"sru:diagnostics,omitempty"`
}

func NewXMLSRResponse() XMLSRResponse {
	return XMLSRResponse{
		XMLNSSRUResponse: "http://www.loc.gov/zing/srw/",
		Version:          "1.2",
		EchoedRequest:    XMLSREchoedRequest{Version: "1.2"},
	}
}

// --------------------- Search Retrieve Record ---------------------

type XMLSRRecord struct {
	Schema         string        `xml:"sru:recordSchema"`
	RecordPacking  string        `xml:"sru:recordPacking"`
	Data           XMLSRResource `xml:"sru:recordData>fcs:Resource"`
	RecordPosition int           `xml:"sru:recordPosition"`
}

type XMLSRResource struct {
	XMLNSFCS         string                `xml:"xmlns:fcs,attr"`
	PID              string                `xml:"pid,attr"`
	ResourceFragment XMLSRResourceFragment `xml:"fcs:ResourceFragment"`
}

type XMLSRResourceFragment struct {
	Ref       string        `xml:"ref,attr,omitempty"`
	DataViews XMLSRDataView `xml:"fcs:DataView"`
}

type XMLSRDataView struct {
	Type   string                   `xml:"type,attr"`
	Result XMLSRBasicDataViewResult `xml:"hits:Result"`
}

type XMLSRBasicDataViewResult struct {
	XMLNSHits string `xml:"xmlns:hits,attr"`
	Data      string `xml:",innerxml"`
}

// --------------------- Echoed Search Retrieve Request ---------------------

type XMLSREchoedRequest struct {
	Version     string `xml:"sru:version"`
	Query       string `xml:"sru:query"`
	StartRecord int    `xml:"sru:startRecord"`
}
