package diff_test

import (
	"fmt"
	"testing"

	"github.com/Merovius/diff"
)

func TestDiff(t *testing.T) {
	tcs := []struct {
		a       []uint64
		b       []uint64
		wantLen int
	}{
		{[]uint64{}, []uint64{}, 0},
		{[]uint64{}, []uint64{0}, 0},
		{[]uint64{0}, []uint64{}, 0},
		{[]uint64{0}, []uint64{0}, 1},
		{[]uint64{0}, []uint64{1}, 0},
		{[]uint64{0}, []uint64{0, 1}, 1},
		{[]uint64{10, 20, 30, 40}, []uint64{1, 10, 20, 25, 40, 45}, 3},
		{[]uint64{2, 4, 6}, []uint64{1, 2, 3, 4, 5}, 2},
		{
			[]uint64{0x7d774ae01bb778e, 0xf7cbe5773314f049, 0x876e85dc2a33ae69},
			[]uint64{0xc1b3cc1b6eb8bf8d, 0x7d774ae01bb778e, 0x9e88d894119ac19e, 0xf7cbe5773314f049, 0x335bc4df9fa0125d},
			2,
		},
		{[]uint64{1, 1, 1, 3, 4, 4}, []uint64{0, 1, 0, 1, 0, 3, 1, 4, 5, 4, 6}, 5},
		{[]uint64{23, 42}, []uint64{23, 23, 42}, 2},
		{[]uint64{42, 23}, []uint64{42, 42, 23}, 2},
		{[]uint64{0xaf63ae4c86019e62}, []uint64{}, 0},
	}

	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			ac := append([]uint64{}, tc.a...)
			bc := append([]uint64{}, tc.b...)
			got := diff.Uint64(tc.a, tc.b)
			if !sameSeq(tc.a, ac) {
				t.Fatalf("Diff modified a")
			}
			if !sameSeq(tc.b, bc) {
				t.Fatalf("Diff modified b")
			}
			if n := commonLength(got); n != tc.wantLen {
				t.Fatalf("Diff(%v, %v) = %v, common sequence of length %d, want %d", tc.a, tc.b, got, n, tc.wantLen)
			}
			var gotA, gotB []uint64

			for _, o := range got {
				if o < diff.OpB && len(ac) == 0 {
					t.Fatalf("Diff(%v, %v) = %v, too many OpA", tc.a, tc.b, got)
				}
				if o > diff.OpA && len(bc) == 0 {
					t.Fatalf("Diff(%v, %v) = %v, too many OpB", tc.a, tc.b, got)
				}
				if o == diff.OpEq && ac[0] != bc[0] {
					t.Fatalf("Diff(%v, %v) = %v, claims different elements are equal", tc.a, tc.b, got)
				}
				if o != diff.OpEq && len(ac) > 0 && len(bc) > 0 && ac[0] == bc[0] {
					t.Fatalf("Diff(%v, %v) = %v, claims equal elements are different", tc.a, tc.b, got)
				}
				if o < diff.OpB {
					gotA, ac = append(gotA, ac[0]), ac[1:]
				}
				if o > diff.OpA {
					gotB, bc = append(gotB, bc[0]), bc[1:]
				}
			}
			if len(ac) > 0 {
				t.Fatalf("Diff(%v, %v) = %v, too few OpA", tc.a, tc.b, got)
			}
			if len(bc) > 0 {
				t.Fatalf("Diff(%v, %v) = %v, too few OpB", tc.a, tc.b, got)
			}
			if !sameSeq(tc.a, gotA) || !sameSeq(tc.b, gotB) {
				t.Errorf("Diff(%v, %v) = %v, restores to wrong sequences:", tc.a, tc.b, got)
				t.Errorf("Restored a: %v", gotA)
				t.Errorf("Restored b: %v", gotB)
			}
		})
	}
}

func commonLength(d []diff.Op) int {
	var total int
	for _, o := range d {
		if o == diff.OpEq {
			total += 1
		}
	}
	return total
}

func sameSeq(l, r []uint64) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}

func ExampleUint64() {
	a := []uint64{1, 1, 1, 3, 4, 4}
	b := []uint64{0, 1, 0, 1, 0, 3, 1, 4, 5, 4, 6}
	for _, o := range diff.Uint64(a, b) {
		switch o {
		case diff.OpA:
			fmt.Printf("-%d\n", a[0])
			a = a[1:]
		case diff.OpEq:
			fmt.Printf(" %d\n", a[0])
			a, b = a[1:], b[1:]
		case diff.OpB:
			fmt.Printf("+%d\n", b[0])
			b = b[1:]
		}
	}
	// Output:
	// +0
	//  1
	// +0
	//  1
	// +0
	// +3
	//  1
	// -3
	//  4
	// +5
	//  4
	// +6
}

func ExampleText() {
	a := []byte("a\nb\nc\nd\nf\ng\nh\nj\nq\nz")
	b := []byte("a\nb\nc\nd\ne\nf\ng\ni\nj\nk\nr\nx\ny\nz")
	for _, δ := range diff.Text(a, b, nil, nil) {
		switch δ.Op {
		case diff.OpA:
			fmt.Printf("- %s\n", δ.Text)
		case diff.OpEq:
			fmt.Printf("  %s\n", δ.Text)
		case diff.OpB:
			fmt.Printf("+ %s\n", δ.Text)
		}
	}
	// Output:
	//   a
	//   b
	//   c
	//   d
	// + e
	//   f
	//   g
	// - h
	// + i
	//   j
	// - q
	// + k
	// + r
	// + x
	// + y
	//   z
}

func TestSplitLines(t *testing.T) {
	tcs := []struct {
		in   string
		want []string
	}{
		{"a", []string{"a"}},
		{"a\n", []string{"a"}},
		{"\na", []string{"", "a"}},
		{"a\nb", []string{"a", "b"}},
		{"a\nb\nc", []string{"a", "b", "c"}},
	}
	for _, tc := range tcs {
		t.Run("", func(t *testing.T) {
			in := tc.in
			var got []string
			for len(in) > 0 {
				tok, skip := diff.SplitLines([]byte(in))
				if tok < 0 || skip < 0 || (tok == 0 && skip == 0) {
					t.Fatalf("SplitLines(%q) = %d, %d, want both non-negative and at least one positive", in, tok, skip)
				}
				if tok+skip > len(in) {
					t.Fatalf("SplitLines(%q) = %d, %d, skips past the end of the input", in, tok, skip)
				}
				got = append(got, in[:tok])
				in = in[tok+skip:]
			}
			if len(got) != len(tc.want) {
				t.Fatalf("SplitLines(%q) → %q, want %q", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("SplitLines(%q) → %q, want %q", tc.in, got, tc.want)
				}
			}
		})
	}
}
