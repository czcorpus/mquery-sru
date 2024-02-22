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

package fcsql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFCSQLParser(t *testing.T) {
	queries := []string{
		`"walking"`,
		`[token = "walking"] within p`,
		`"Dog" /c`,
		`[word = "Dog" /c]`,
		`[pos = "NOUN"]`,
		`[pos != "NOUN"]`,
		`[lemma = "walk"]`,
		`"blaue|grüne" [pos = "NOUN"]`,
		`"dogs" []{3,} "cats" within s`,
		`[z:pos = "ADJ"]`,
		`[z:pos="ADJ" & q:pos="ADJ"]`,
		`[ (word="foo") ]`,
		`[( word="foo" )]`,
	}

	for i, q := range queries {
		ans, err := Parse(fmt.Sprintf("test_%d", i), []byte(q)) // Debug(true))
		assert.NoError(t, err)
		if ans != nil {
			fmt.Printf("ans = %#v\n", ans.(*Query).Generate())
		}

	}
}
