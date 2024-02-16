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

package tests

import (
	"net/url"
)

type Data struct {
	Total int      `xml:"numberOfRecords"`
	Hits  []string `xml:"records>record>recordData>Resource>ResourceFragment>DataView>Result>Hit"`
}

func (suite *IntegrationTestSuite) TestBasicQueryResponse() {
	param := make(url.Values)
	param.Set("version", "2.0")
	param.Set("operation", "searchRetrieve")
	param.Set("queryType", "cql")
	param.Set("query", "word_A923_tag2 OR word_B923_tag2")

	uri := suite.uri
	uri.RawQuery = param.Encode()

	var data Data
	suite.makeRequest(uri, &data)

	// check number of results
	suite.Equal(2, data.Total)
	// check number of hits on first page
	suite.Len(data.Hits, 2)
	// check hits
	suite.Exactly(
		[]string{
			"word_A923_tag2",
			"word_B923_tag2",
		},
		data.Hits,
	)
}

func (suite *IntegrationTestSuite) TestAdvancedQueryResponse() {
	param := make(url.Values)
	param.Set("version", "2.0")
	param.Set("operation", "searchRetrieve")
	param.Set("queryType", "fcs")
	param.Set("query", "[text=\"word.*\"]")

	uri := suite.uri
	uri.RawQuery = param.Encode()

	var data Data
	suite.makeRequest(uri, &data)

	// check number of results
	suite.Equal(2000, data.Total)
	// check number of hits on first page
	suite.Len(data.Hits, 10)
	// check hits
	expectedHits := []string{
		"word_A923_tag2",
		"word_B923_tag2",
		"word_A754_tag1",
		"word_B754_tag1",
		"word_A555_tag3",
		"word_B555_tag3",
		"word_A195_tag3",
		"word_B195_tag3",
		"word_A944_tag2",
		"word_B944_tag2",
	}
	suite.Exactly(expectedHits, data.Hits)
}
