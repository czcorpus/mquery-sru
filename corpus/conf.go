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
	"sort"
	"strings"

	"github.com/czcorpus/cnc-gokit/collections"
	"github.com/czcorpus/cnc-gokit/fs"
)

const (
	LayerTypeText     LayerType = "text"
	LayerTypeLemma    LayerType = "lemma"
	LayerTypePOS      LayerType = "pos"
	LayerTypeOrth     LayerType = "orth"
	LayerTypeNorm     LayerType = "norm"
	LayerTypePhonetic LayerType = "phonetic"

	DefaultLayerType = LayerTypeText
)

// LayerType is a layer above positional attributes
// combining similar attribute types (e.g. annotation).
// In Manatee, no such thing is defined but it is
// nevertheless supported via configuration of corpora
// in MQuery-FCS where each positional attribute belongs
// to a specific layer.
type LayerType string

func (name LayerType) Validate() error {
	if name == LayerTypeText ||
		name == LayerTypeLemma ||
		name == LayerTypePOS ||
		name == LayerTypeOrth ||
		name == LayerTypeNorm ||
		name == LayerTypePhonetic {
		return nil
	}
	return fmt.Errorf("invalid layer name `%s`", name)
}

func (name LayerType) GetResultID() string {
	switch name {
	case LayerTypeText:
		return "http://clarin.dk/ns/fcs/layer/word"
	case LayerTypeLemma:
		return "http://clarin.dk/ns/fcs/layer/lemma"
	case LayerTypePOS:
		return "http://clarin.dk/ns/fcs/layer/pos"
	case LayerTypeOrth:
		return "http://clarin.dk/ns/fcs/layer/orth"
	case LayerTypeNorm:
		return "http://clarin.dk/ns/fcs/layer/norm"
	case LayerTypePhonetic:
		return "http://clarin.dk/ns/fcs/layer/phonetic"
	}
	return ""
}

// PosAttr represents a corpus positional attribute
type PosAttr struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Layer defines a layer the attribute is attached to
	// (this is not supported directly by Manatee so it
	// is configured and supported in MQuery-FCS)
	Layer LayerType `json:"layer"`

	// IsLayerDefault defines whether the attribute is
	// used as a default one when querying its layer.
	// (e.g. the `word` attribute is typically set as
	// the default for the `text` layer)
	IsLayerDefault bool `json:"isLayerDefault"`
}

// StructureMapping provides mapping between custom
// corpus structures and FCS-QL generic structures
// (paragraph, sentence, utterance,...)
type StructureMapping struct {
	SentenceStruct  string `json:"sentenceStruct"`
	UtteranceStruct string `json:"utteranceStruct"`
	ParagraphStruct string `json:"paragraphStruct"`
	TurnStruct      string `json:"turnStruct"`
	TextStruct      string `json:"textStruct"`
	SessionStruct   string `json:"sessionStruct"`
}

// CorpusSetup is a complete corpus configuration
// (it is part of MQuery-FCS configuration)
type CorpusSetup struct {
	PID              string           `json:"pid"`
	FullName         string           `json:"fullName"`
	PosAttrs         []PosAttr        `json:"posAttrs"`
	StructureMapping StructureMapping `json:"structureMapping"`
}

// GetLayerDefault provides default positional
// attribute for a specified layer.
func (cs *CorpusSetup) GetLayerDefault(ln LayerType) PosAttr {
	for _, item := range cs.PosAttrs {
		if item.IsLayerDefault {
			return item
		}
	}
	return PosAttr{}
}

// GetDefinedLayers returns all the layers defined for the corpus
func (cs *CorpusSetup) GetDefinedLayers() *collections.Set[LayerType] {
	ans := collections.NewSet[LayerType]()
	for _, item := range cs.PosAttrs {
		ans.Add(item.Layer)
	}
	return ans
}

// GetDefinedLayersAsString provides all the layers
// defined for the corpus formatted as a single string
// (this is required in SRU XML)
func (cs *CorpusSetup) GetDefinedLayersAsString() string {
	layers := cs.GetDefinedLayers().ToOrderedSlice()
	ans := make([]string, len(layers))
	for i, v := range layers {
		ans[i] = string(v)
	}
	return strings.Join(ans, " ")
}

// Validate validates corpus setup. This should be run
// as part of server startup (i.e. before any requests start)
func (ls *CorpusSetup) Validate(confContext string) error {
	if ls == nil {
		return fmt.Errorf("missing configuration section `%s.layers`", confContext)
	}
	layerDefaults := make(map[LayerType]int)
	for _, attr := range ls.PosAttrs {
		if err := attr.Layer.Validate(); err != nil {
			return err
		}
		if attr.IsLayerDefault {
			layerDefaults[attr.Layer]++
		}
	}
	for layer, num := range layerDefaults {
		if num != 1 {
			return fmt.Errorf(
				"invalid number of isLayerDefault items for layer %s: %d (must be 1)",
				layer,
				num,
			)
		}
	}

	return nil
}

// -----

// SrchResources is a configuration of all the enabled
// corpora.
type SrchResources map[string]*CorpusSetup

func (sr SrchResources) GetCommonLayers() []LayerType {
	var ans *collections.Set[LayerType]
	for _, corp := range sr {
		if ans == nil {
			ans = corp.GetDefinedLayers()

		} else {
			ans = ans.Intersect(corp.GetDefinedLayers())
		}
	}
	return ans.ToOrderedSlice()
}

func (sr SrchResources) GetCorpora() []string {
	ans := make([]string, len(sr))
	i := 0
	for k := range sr {
		ans[i] = k
		i++
	}
	return ans
}

// GetCommonPosAttrs returns positional attributes common
// to provided corpora. The attribute of the text layer which
// is set as default will be listed always first, the rest
// is sorted alphabetically.
func (sr SrchResources) GetCommonPosAttrs(corpusNames ...string) []PosAttr {
	collect := make(map[string]PosAttr)
	for _, corp := range corpusNames {
		for _, pa := range sr[corp].PosAttrs {
			collect[pa.Name] = pa
		}
	}
	i := 0
	ans := make([]PosAttr, len(collect))
	for _, v := range collect {
		ans[i] = v
		i++
	}
	sort.SliceStable(ans, func(i, j int) bool {
		if ans[i].Layer == DefaultLayerType && ans[i].IsLayerDefault {
			return true
		}
		if ans[j].Layer == DefaultLayerType && ans[j].IsLayerDefault {
			return false
		}
		return strings.Compare(ans[i].Name, ans[j].Name) < 0
	})
	return ans
}

// GetCommonPosAttrNames is the same as GetCommonPosAttrs
// but it returns just a list of attribute names.
func (sr SrchResources) GetCommonPosAttrNames(corpusName ...string) []string {
	pa := sr.GetCommonPosAttrs(corpusName...)
	ans := make([]string, len(pa))
	for i, pa := range pa {
		ans[i] = pa.Name
	}
	return ans
}

// Validate validates all the corpora configurations.
// This should be run during server startup.
func (sr SrchResources) Validate(confContext string) error {
	for name, corp := range sr {
		if err := corp.Validate(fmt.Sprintf("%s[%s]", confContext, name)); err != nil {
			return err
		}
	}
	return nil
}

// ---

// CorporaSetup defines mquery application configuration related
// to a corpus
type CorporaSetup struct {
	RegistryDir string        `json:"registryDir"`
	Resources   SrchResources `json:"resources"`
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
	return cs.Resources.Validate("resources")
}
