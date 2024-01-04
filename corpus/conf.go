// Copyright 2019 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright 2019 Institute of the Czech National Corpus,
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

package corpus

import (
	"fmt"
	"path/filepath"

	"github.com/czcorpus/cnc-gokit/fs"
)

type PosAttrProps struct {
	Name string `json:"name"`
}

type CorpusSetup struct {
	DefaultSearchAttr string       `json:"defaultSearchAttr"`
	AvailableLayers   []string     `json:"availableLayers"`
	SyntaxParentAttr  PosAttrProps `json:"syntaxParentAttr"`
}

type LayersSetup struct {
	Text     string `json:"text"`
	Lemma    string `json:"lemma"`
	POS      string `json:"pos"`
	Orth     string `json:"orth"`
	Norm     string `json:"norm"`
	Phonetic string `json:"phonetic"`
}

func (ls *LayersSetup) ToDict() map[string]string {
	layers := make(map[string]string)
	layers["text"] = ls.Text
	if ls.Lemma != "" {
		layers["lemma"] = ls.Lemma
	}
	if ls.POS != "" {
		layers["pos"] = ls.POS
	}
	if ls.Orth != "" {
		layers["orth"] = ls.Orth
	}
	if ls.Norm != "" {
		layers["norm"] = ls.Norm
	}
	if ls.Phonetic != "" {
		layers["phonetic"] = ls.Phonetic
	}
	return layers
}

func (ls *LayersSetup) Validate(confContext string) error {
	if ls == nil {
		return fmt.Errorf("missing configuration section `%s.layers`", confContext)
	}
	if ls.Text == "" {
		return fmt.Errorf("missing `%s.layers.text`", confContext)
	}
	return nil
}

// CorporaSetup defines mquery application configuration related
// to a corpus
type CorporaSetup struct {
	RegistryDir string                  `json:"registryDir"`
	Layers      *LayersSetup            `json:"layers"`
	Resources   map[string]*CorpusSetup `json:"resources"`
}

func (cs *CorporaSetup) GetRegistryPath(corpusID string) string {
	return filepath.Join(cs.RegistryDir, corpusID)
}

func (cs *CorporaSetup) ValidateAndDefaults(confContext string) error {
	if cs == nil {
		return fmt.Errorf("missing configuration section `%s`", confContext)
	}
	if cs.RegistryDir == "" {
		return fmt.Errorf("missing `%s.registryDir`", confContext)
	}
	isDir, err := fs.IsDir(cs.RegistryDir)
	if err != nil {
		return fmt.Errorf("failed to test `%s.registryDir`: %w", confContext, err)
	}
	if !isDir {
		return fmt.Errorf("`%s.registryDir` is not a directory", confContext)
	}
	return cs.Layers.Validate(confContext)
}
