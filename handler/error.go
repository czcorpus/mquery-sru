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

package handler

// https://www.loc.gov/standards/sru/diagnostics/diagnosticsList.html
const (
	// General diagnostics
	CodeGeneralSystemError            = 1
	CodeSystemTemporarilyUnavailable  = 2
	CodeAuthenticationError           = 3
	CodeUnsupportedOperation          = 4
	CodeUnsupportedVersion            = 5
	CodeUnsupportedParameterValue     = 6
	CodeMandatoryParameterNotSupplied = 7
	CodeUnsupportedParameter          = 8
	CodeDatabaseDoesNotExist          = 235
	// CQL related diagnostics
	CodeQuerySyntaxError = 10
	// Records related diagnostics
	CodeUnsupportedRecordPacking = 71
)

type FCSError struct {
	Code    int
	Ident   string
	Message string
}
