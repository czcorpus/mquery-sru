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

package basic

import (
	"fcs/general"
	"fmt"
)

type node interface {
	transform(attr string) (string, *general.FCSError)
	isLeaf() bool
}

// ------------- root node ----------------

type rootNode struct {
	Child node
}

func (r *rootNode) transform(attr string) (string, *general.FCSError) {
	return r.Child.transform(attr)
}

func (r *rootNode) isLeaf() bool {
	return r.Child == nil
}

// ------------- term node ----------------

type termNode struct {
	Value string
}

func (t *termNode) transform(attr string) (string, *general.FCSError) {
	return fmt.Sprintf("[%s=\"%s\"]", attr, t.Value), nil
}

func (t *termNode) isLeaf() bool {
	return true
}

// ------------- unary node ----------------

type unaryNode struct {
	Op    string
	Child node
}

func (u *unaryNode) transform(attr string) (string, *general.FCSError) {
	child, err := u.Child.transform(attr)
	if err != nil {
		return "", err
	}

	switch u.Op {
	case "NOT":
		return fmt.Sprintf("!%s", child), nil
	}
	return "", &general.FCSError{
		Code:    general.CodeQueryFeatureUnsupported,
		Ident:   u.Op,
		Message: "Query feature unsupported",
	}
}

func (u *unaryNode) isLeaf() bool {
	return false
}

// ------------- binary node ----------------

type binaryNode struct {
	Op    string
	Left  node
	Right node
}

func (b *binaryNode) transform(attr string) (string, *general.FCSError) {
	left, err := b.Left.transform(attr)
	if err != nil {
		return "", err
	}
	right, err := b.Right.transform(attr)
	if err != nil {
		return "", err
	}

	if !b.Left.isLeaf() {
		left = fmt.Sprintf("(%s)", left)
	}
	if !b.Right.isLeaf() {
		right = fmt.Sprintf("(%s)", right)
	}

	switch b.Op {
	case "AND":
		return fmt.Sprintf("%s&%s", left, right), nil
	case "OR":
		return fmt.Sprintf("%s|%s", left, right), nil
	}
	return "", &general.FCSError{
		Code:    general.CodeQueryFeatureUnsupported,
		Ident:   b.Op,
		Message: "Query feature unsupported",
	}
}

func (b *binaryNode) isLeaf() bool {
	return false
}

// ------------- paren node ----------------

type parenNode struct {
	Child node
}

func (p *parenNode) transform(attr string) (string, *general.FCSError) {
	child, err := p.Child.transform(attr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s)", child), nil
}

func (p *parenNode) isLeaf() bool {
	return true
}
