// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
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

package general

const (

	// ConformantStatusBadRequest
	// Note: we want to keep awareness about proper
	// states but to keep in line with the SRU specification,
	// 200 is expected
	ConformantStatusBadRequest = 200

	// ConformantUnprocessableEntity
	// Note: we want to keep awareness about proper
	// states but to keep in line with the SRU specification,
	// 200 is expected
	ConformantUnprocessableEntity = 200

	// ConformandGeneralServerError
	// Note: we want to keep awareness about proper
	// states but to keep in line with the SRU specification,
	// 200 is expected
	ConformandGeneralServerError = 200

	RecordSchema = "http://clarin.eu/fcs/resource"
)

type FCSGeneralRequest struct {
	Version string
	Errors  []FCSError
	Fatal   bool

	// XSLT is an optional path of a XSL template
	// for outputting formatted (typically HTML) result
	XSLT string
}

func (r *FCSGeneralRequest) AddError(fcsError FCSError) {
	if fcsError.IsFatal() {
		r.Fatal = true
		if fcsError.Overthrow() {
			r.Errors = r.Errors[0:0]
		}
	}
	r.Errors = append(r.Errors, fcsError)
}

func (r *FCSGeneralRequest) HasFatalError() bool {
	return len(r.Errors) > 0 && r.Fatal
}
