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

import "fmt"

func MapItems[K string, V any, T any](data map[K]V, mapFn func(k K, v V) T) []T {
	ans := make([]T, len(data))
	i := 0
	for k, v := range data {
		ans[i] = mapFn(k, v)
		i++
	}
	return ans
}

func ReturnIf[T any](cond bool, ifTrue T, ifFalse T) T {
	if cond {
		return ifTrue
	}
	return ifFalse
}

func GetXSLTHeader(xslt string) string {
	if xslt != "" {
		return fmt.Sprintf("<?xml-stylesheet type=\"text/xsl\" href=\"%s\"?>\n", xslt)
	}
	return ""
}
