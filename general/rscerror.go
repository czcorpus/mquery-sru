// Copyright 2026 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2026 Institute of the Czech National Corpus,
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

import "fmt"

// ResourceError describes a failure of a search in a single resource
// (corpus) within a multi-resource search. Such a failure is not
// fatal as long as there are other resources able to provide data.
type ResourceError struct {
	ID  string
	PID string
	Err error
}

// DiagnosticMessage returns a message suitable for a non-fatal
// SRU diagnostic
func (re ResourceError) DiagnosticMessage() string {
	return fmt.Sprintf("Resource skipped due to an error: %s", re.Err)
}
