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

import (
	"fmt"
)

type DiagnosticType int

type DiagnosticCode int

func (dc DiagnosticCode) AsMessage() string {
	switch dc {
	case DCGeneralSystemError:
		return "General system error"
	case DCSystemTemporarilyUnavailable:
		return "System temporarily unavailable"
	case DCAuthenticationError:
		return "Authentication error"
	case DCUnsupportedOperation:
		return "Unsupported operation"
	case DCUnsupportedVersion:
		return "Unsupported version"
	case DCUnsupportedParameterValue:
		return "Unsupported parameter value"
	case DCMandatoryParameterNotSupplied:
		return "Mandatory parameter not supplied"
	case DCUnsupportedParameter:
		return "Unsupported Parameter"
	case DCUnsupportedContextSet:
		return "Unsupported context set"
	case DCUnsupportedIndex:
		return "Unsupported index"
	case DCDatabaseDoesNotExist:
		return "Database does not exist"
	case DCQuerySyntaxError:
		return "Query syntax error"
	case DCQueryCannotProcess:
		return "Cannot process query; reason unknown"
	case DCQueryFeatureUnsupported:
		return "Query feature unsupported"
	case DCTooManyMatchingRecords:
		return "Result set not created: too many matching records"
	case DCFirstRecordPosOutOfRange:
		return "First record position out of range"
	case DCUnknownSchemaForRetrieval:
		return "Unknown schema for retrieval"
	case DCUnsupportedRecordPacking:
		return "Unsupported record packing"
	}
	return "??"
}

// from appendix A FCS 2.0 documentation
const (
	DTPersistent                            DiagnosticType = 1  // generally non-fatal, with code fatal?
	DTResourceSetTooLarge                   DiagnosticType = 2  // non-fatal
	DTResourceSetTooLargeCannotPerformQuery DiagnosticType = 3  // fatal
	DTRequestedDataViewNotValid             DiagnosticType = 4  // non-fatal
	DTGeneralQuerySyntaxError               DiagnosticType = 10 // fatal, return only this one
	DTQueryTooComplex                       DiagnosticType = 11 // fatal, return only this one
	DTQueryWasRewritten                     DiagnosticType = 12 // non-fatal, only advanced query with `x-fcs-rewrites-allowed`
	DTGeneralProcessingHint                 DiagnosticType = 14 // non-fatal, only advanced query
)

// https://www.loc.gov/standards/sru/diagnostics/diagnosticsList.html
// used with diagnostic type 1
const (
	// General diagnostics
	DCGeneralSystemError            DiagnosticCode = 1
	DCSystemTemporarilyUnavailable  DiagnosticCode = 2
	DCAuthenticationError           DiagnosticCode = 3
	DCUnsupportedOperation          DiagnosticCode = 4
	DCUnsupportedVersion            DiagnosticCode = 5
	DCUnsupportedParameterValue     DiagnosticCode = 6
	DCMandatoryParameterNotSupplied DiagnosticCode = 7
	DCUnsupportedParameter          DiagnosticCode = 8
	DCDatabaseDoesNotExist          DiagnosticCode = 235
	// CQL related diagnostics
	DCQuerySyntaxError        DiagnosticCode = 10
	DCUnsupportedContextSet   DiagnosticCode = 15
	DCUnsupportedIndex        DiagnosticCode = 16
	DCQueryCannotProcess      DiagnosticCode = 47
	DCQueryFeatureUnsupported DiagnosticCode = 48
	// Diagnostics Relating to Records
	DCTooManyMatchingRecords    DiagnosticCode = 60
	DCFirstRecordPosOutOfRange  DiagnosticCode = 61
	DCUnknownSchemaForRetrieval DiagnosticCode = 66
	// Records related diagnostics
	DCUnsupportedRecordPacking DiagnosticCode = 71
)

type FCSError struct {
	Type    DiagnosticType
	Code    DiagnosticCode
	Ident   string
	Message string
}

func (fe FCSError) Error() string {
	return fmt.Sprintf("%d: %s (%s)", fe.Code, fe.Message, fe.Ident)
}

func (fe FCSError) IsFatal() bool {
	return fe.Type == DTResourceSetTooLargeCannotPerformQuery || fe.Type == DTGeneralQuerySyntaxError || fe.Type == DTQueryTooComplex || fe.Code > 0
}

func (fe FCSError) Overthrow() bool {
	return fe.Type == DTGeneralQuerySyntaxError || fe.Type == DTQueryTooComplex
}
