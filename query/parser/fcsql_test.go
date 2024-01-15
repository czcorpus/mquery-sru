// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2023 Martin Zimandl <martin.zimandl@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

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
	}

	for i, q := range queries {
		ans, err := Parse(fmt.Sprintf("test_%d", i), []byte(q)) // Debug(true))
		if ans != nil {
			fmt.Printf("ans = %#v\n", ans.(*query).String())
		}
		assert.NoError(t, err)
	}
}
