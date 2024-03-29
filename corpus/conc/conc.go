// Copyright 2023 Tomas Machalek <tomas.machalek@gmail.com>
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

package conc

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/czcorpus/mquery-sru/mango"

	"github.com/rs/zerolog/log"
)

const (
	invalidParent = 1000000
)

var (
	splitPatt = regexp.MustCompile(`\s+`)
)

type TokenSlice []*Token

type Token struct {
	Word   string            `json:"word"`
	Strong bool              `json:"strong"`
	Attrs  map[string]string `json:"attrs"`
}

type ConcordanceLine struct {
	Text TokenSlice `json:"text"`
	Ref  string     `json:"ref"`
}

type ConcExamples struct {
	Lines []ConcordanceLine `json:"lines"`
}

type LineParser struct {
	attrs []string
}

func (lp *LineParser) parseTokenQuadruple(s []string) *Token {
	mAttrs := make(map[string]string)
	rawAttrs := strings.Split(s[2], "/")[1:]
	var token Token
	if len(rawAttrs) != len(lp.attrs)-1 {
		log.Warn().
			Str("value", s[2]).
			Int("expectedNumAttrs", len(lp.attrs)-1).
			Msg("cannot parse token quadruple")
		token.Word = s[0]
		for _, attr := range lp.attrs[1:] {
			mAttrs[attr] = "N/A"
		}

	} else {
		for i, attr := range lp.attrs[1:] {
			mAttrs[attr] = rawAttrs[i]
		}
		token.Word = s[0]
		token.Strong = len(s[1]) > 2
		token.Attrs = mAttrs
	}
	return &token
}

func (lp *LineParser) normalizeTokens(tokens []string) []string {
	ans := make([]string, 0, len(tokens))
	var parTok strings.Builder
	for _, tok := range tokens {
		tokLen := utf8.RuneCountInString(tok)
		if tok == "" {
			continue

		} else if tokLen == 1 {
			ans = append(ans, tok)

		} else if tok[0] == '{' {
			if tok[tokLen-1] != '}' {
				parTok.WriteString(tok)

			} else {
				ans = append(ans, tok)
			}

		} else if tok[tokLen-1] == '}' {
			parTok.WriteString(tok)
			ans = append(ans, parTok.String())
			parTok.Reset()

		} else {
			ans = append(ans, tok)
		}
	}
	return ans
}

// parseRawLine
func (lp *LineParser) parseRawLine(line string) ConcordanceLine {
	rtokens := splitPatt.Split(html.EscapeString(line), -1)
	items := lp.normalizeTokens(rtokens[1:])
	if len(items)%4 != 0 {
		log.Error().
			Str("origLine", line).
			Msg("unparseable Manatee KWIC line")
		return ConcordanceLine{
			Text: []*Token{{Word: "---- ERROR (unparseable) ----"}},
			Ref:  rtokens[0],
		}
	}
	tokens := make(TokenSlice, 0, len(items)/4)
	for i := 0; i < len(items); i += 4 {
		tokens = append(tokens, lp.parseTokenQuadruple(items[i:i+4]))
	}
	return ConcordanceLine{Text: tokens, Ref: rtokens[0]}
}

// Parse converts Manatee-encoded concordance lines into MQuery format.
// It also escapes strings to make them usable in XML documents.
func (lp *LineParser) Parse(data mango.GoConcExamples) []ConcordanceLine {
	pLines := make([]ConcordanceLine, len(data.Lines))
	for i, line := range data.Lines {
		pLines[i] = lp.parseRawLine(line)
	}
	return pLines
}

func NewLineParser(attrs []string) *LineParser {
	return &LineParser{
		attrs: attrs,
	}
}
