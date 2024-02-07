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

package common

import (
	"html"
	"text/template"

	"github.com/czcorpus/cnc-gokit/strutil"
)

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"add": func(i, j int) int {
			return i + j
		},
		"escape": html.EscapeString,
		"smartTruncate100": func(s string) string {
			return strutil.SmartTruncate(s, 100)
		},
		"smartTruncate200": func(s string) string {
			return strutil.SmartTruncate(s, 200)
		},
		"enMsgFrom": func(msg map[string]string) string {
			v, ok := msg["en"]
			if !ok {
				return "??"
			}
			return v
		},
	}
}
