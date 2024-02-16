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
	"encoding/xml"
	"io"
	"net/http"
	"net/url"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	uri *url.URL
}

func (suite *IntegrationTestSuite) SetupTest() {
}

func (suite *IntegrationTestSuite) makeRequest(uri *url.URL, data any) {
	req, err := http.NewRequest("GET", uri.String(), nil)
	suite.NoError(err)
	client := http.Client{}
	response, err := client.Do(req)
	suite.NoError(err)
	suite.Equal(http.StatusOK, response.StatusCode, "Unexpected response code")

	var body []byte
	body, err = io.ReadAll(response.Body)
	suite.NoError(err)
	err = response.Body.Close()
	suite.NoError(err)

	err = xml.Unmarshal(body, &data)
	suite.NoError(err, body)
}

func NewIntegrationTestSuite(endpoint string) (*IntegrationTestSuite, error) {
	uri, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &IntegrationTestSuite{uri: uri}, nil
}
