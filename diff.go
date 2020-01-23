// Package diff calculates the differences between two sequences.
//
// It implements the algorithm from "An Algorithm for Differential File
// Comparison" by Hunt and McIlroy:
// https://www.cs.dartmouth.edu/~doug/diff.pdf
//
// For flexibility, the algorithm itself operates on a sequence of integers.
// This allows you to compare arbitrary sequences, as long as you can map their
// elements to a uint64.
//
// To generate a diff for text, the inputs need to be split and hashed.
// Splitting should be done to reduce algorithmic complexity (which is
// O(m•n•log(m)) in the worst case). It also creates diffs that are better
// suited for human consumption. Hashing means that collisions are possible,
// but they should be rare enough in practice to not matter. If they do happen,
// the resulting diff might be subpoptimal.
package diff

import (
	"bytes"
	"fmt"
	"sort"
)

type Op int

const (
	OpA, OpEq, OpB Op = -1, 0, 1
)

// Diff calculates a minimal diff between a and b as a series of operations.
// See the example for how to interpret the result.
func Uint64(a, b []uint64) []Op {
	var prefix, suffix []Op
	for len(a) > 0 && len(b) > 0 && a[0] == b[0] {
		prefix = append(prefix, OpEq)
		a, b = a[1:], b[1:]
	}
	for len(a) > 0 && len(b) > 0 && a[len(a)-1] == b[len(b)-1] {
		suffix = append(suffix, OpEq)
		a, b = a[:len(a)-1], b[:len(b)-1]
	}
	for i := 0; i < len(suffix)/2; i++ {
		a[i], a[len(a)-i-1] = a[len(a)-i-1], a[i]
	}
	if len(a) == 0 {
		out := make([]Op, len(b))
		for i := range out {
			out[i] = OpB
		}
		return append(append(prefix, out...), suffix...)
	}
	if len(b) == 0 {
		out := make([]Op, len(a))
		for i := range out {
			out[i] = OpA
		}
		return append(append(prefix, out...), suffix...)
	}

	swap := Op(1)
	if len(a) < len(b) {
		a, b = b, a
		swap = Op(-1)
	}

	l := makeClasses(b)
	p := make([]int, len(a))
	for i, v := range a {
		p[i] = l.findFirst(v)
	}

	K := []*candidate{
		&candidate{-1, -1, nil},
	}

	// i is an index in a, e is an index in l, j is the corresponding index in b.
	for i := range a {
		r, c := 0, K[0]
		for e := p[i]; e < len(l); e++ {
			j := l[e].i
			if j >= len(b) {
				break
			}
			s := sort.Search(len(K), func(s int) bool {
				return K[s].j > j
			})
			if K[s-1].j < j {
				if s == len(K) {
					K = append(K, nil)
				}
				K[r], r, c = c, s, &candidate{i, j, K[s-1]}
				break
			}
			if l[e].last {
				break
			}
		}
		K[r] = c
	}

	// k is the length of the longest common subsequence found. c is the list
	// representing that sequence (from the back).
	k, c := len(K)-1, K[len(K)-1]

	out := make([]Op, len(a)+len(b)-k)
	i, j := len(a)-1, len(b)-1
	for o := len(out) - 1; o >= 0; o-- {
		switch {
		case i > c.i:
			out[o] = OpA * swap
			i--
		case j > c.j:
			out[o] = OpB * swap
			j--
		default:
			out[o] = OpEq
			i--
			j--
			c = c.prev
		}
	}
	return append(append(prefix, out...), suffix...)
}

type element struct {
	i    int
	v    uint64
	last bool
}

type classes []element

func makeClasses(b []uint64) classes {
	l := make(classes, len(b))
	for i, v := range b {
		l[i] = element{i, v, false}
	}
	sort.Slice(l, func(i, j int) bool {
		if l[i].v == l[j].v {
			return l[i].i < l[j].i
		}
		return l[i].v < l[j].v
	})
	for i := 0; i < len(l)-1; i++ {
		if l[i].v != l[i+1].v {
			l[i].last = true
		}
	}
	l[len(l)-1].last = true
	return l
}

func (l classes) findFirst(v uint64) int {
	n := sort.Search(len(l), func(n int) bool {
		return l[n].v >= v
	})
	if n < len(l) && l[n].v != v {
		return len(l)
	}
	return n
}

type candidate struct {
	i    int
	j    int
	prev *candidate
}

// TextDelta describes a line of the resulting diff.
type TextDelta struct {
	Op
	Text []byte
}

// TextDiff calculates a diff between a and b. s is used to separate them into
// tokens and h is used to map those to integers. If s is nil, SplitLines is
// used. If h is nil, DefaultHash is used.
//
// The resulting diff will contain separate TextDelta values per token (even if
// consecutive elements use the same Op). See the example for how to use
// construct the diff from it.
//
// In case of an EqOp delta where the corresponding tokens of a and b differ
// (but hash to the same value), it is unspecified which of the two is
// returned.
func Text(a, b []byte, s SplitFunc, h HashFunc) []TextDelta {
	if s == nil {
		s = SplitLines
	}
	if h == nil {
		h = DefaultHash()
	}
	la, ha := tokenize(a, h, s)
	lb, hb := tokenize(b, h, s)
	diff := Uint64(ha, hb)
	var out []TextDelta
	for _, d := range diff {
		δ := TextDelta{Op: d}
		if d > OpA {
			δ.Text = lb[0]
			lb = lb[1:]
		}
		if d < OpB {
			δ.Text = la[0]
			la = la[1:]
		}
		out = append(out, δ)
	}
	return out
}

// LineDiff is equivalent to TextDiff(a, b, nil, nil).
func Lines(a, b []byte) []TextDelta {
	return Text(a, b, nil, nil)
}

func tokenize(a []byte, h HashFunc, s SplitFunc) ([][]byte, []uint64) {
	var (
		tokens [][]byte
		hashes []uint64
	)
	for len(a) > 0 {
		tok, skip := s(a)
		if tok < 0 || skip < 0 || tok+skip == 0 {
			panic(fmt.Errorf("invalid split (%d,%d)", tok, skip))
		}
		hashes = append(hashes, h(a[:tok]))
		tokens = append(tokens, a[:tok])
		a = a[tok+skip:]
	}
	return tokens, hashes
}

// A SplitFunc splits a token from b. tok specifies the length of the next
// token and skip specifies how many bytes to skip after the token. If neither
// of them is positive or either of them is negative, TextDiff will panic.
type SplitFunc func(b []byte) (tok, skip int)

// SplitLines splits at newlines, stripping them in the process.
func SplitLines(b []byte) (tok, skip int) {
	i := bytes.IndexByte(b, '\n')
	if i < 0 {
		return len(b), 0
	}
	// TODO: Handle \r\n?
	return i, 1
}

// A HashFunc maps a token to an integer.
type HashFunc func(b []byte) uint64

// DefaultHash returns a sensible (but unspecified) hash function for TextDiff.
func DefaultHash() HashFunc {
	return defaultHash()
}
