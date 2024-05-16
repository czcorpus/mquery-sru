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

import (
	"fmt"

	"github.com/czcorpus/cnc-gokit/strutil"
	"github.com/czcorpus/mquery-sru/general"
)

type XMLDiagnostic struct {
	URI     []string `xml:"diag:uri,omitempty"`
	Details string   `xml:"diag:details"`
	Message string   `xml:"diag:message"`
}

type XMLDiagnostics struct {
	XMLNSDiag   string          `xml:"xmlns:diag,attr"`
	Diagnostics []XMLDiagnostic `xml:"diag:diagnostic"`
}

func (d *XMLDiagnostics) AddDiagnostic(code general.DiagnosticCode, typ general.DiagnosticType, ident string, message string) {
	uri := []string{}
	if code > 0 {
		uri = append(uri, fmt.Sprintf("info:srw/diagnostic/1/%d", code))
	}
	if typ > 0 {
		uri = append(uri, fmt.Sprintf("info:srw/diagnostic/%d", typ))
	}
	d.Diagnostics = append(d.Diagnostics, XMLDiagnostic{
		URI:     uri,
		Details: ident,
		Message: strutil.SmartTruncate(message, 200),
	})
}

func NewXMLDiagnostics() *XMLDiagnostics {
	return &XMLDiagnostics{
		XMLNSDiag: "http://docs.oasis-open.org/ns/search-ws/diagnostic",
	}
}
