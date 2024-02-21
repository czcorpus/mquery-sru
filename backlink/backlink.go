// Copyright 2024 Tomas Machalek <tomas.machalek@gmail.com>
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

package backlink

import (
	"fmt"
	"net/url"
)

func GenerateForKonText(rootURL, corpus, mainQuery, tokenID string) (string, error) {
	rurl, err := url.Parse(rootURL)
	if err != nil {
		return "", err
	}
	rurl = rurl.JoinPath("create_view")
	q := make(url.Values)
	q.Add("corpname", corpus)
	q.Add("q", "aword,"+mainQuery)
	q.Add("q", fmt.Sprintf("p0 0 1 [%s]", tokenID))
	rurl.RawQuery = q.Encode()
	return rurl.String(), nil
}
