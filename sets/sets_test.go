package sets

import (
	"cmp"
	"testing"

	cmp2 "github.com/google/go-cmp/cmp"
	"pgregory.net/rapid"
)

func TestUnion(t *testing.T) {
	testCases := map[string]struct {
		Input1   *Set[int]
		Input2   *Set[int]
		Expected *Set[int]
	}{
		"both empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New[int](cmp.Compare),
			Expected: New[int](cmp.Compare),
		},
		"empty and not empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New(cmp.Compare, 1),
			Expected: New(cmp.Compare, 1),
		},
		"not empty and empty": {
			Input1:   New(cmp.Compare, 1),
			Input2:   New[int](cmp.Compare),
			Expected: New(cmp.Compare, 1),
		},
		"with fsmcommon element": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 2, 3),
			Expected: New(cmp.Compare, 1, 2, 3),
		},
		"with no fsmcommon element": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 3, 4),
			Expected: New(cmp.Compare, 1, 2, 3, 4),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Input1.Clone()
			actual.Union(cmp.Compare, testCase.Input2)

			if !IsEqual(cmp.Compare, actual, testCase.Expected) {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("left identity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			xs := New[int](cmp.Compare)
			ys := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "ys")...)
			actual := xs.Clone()
			actual.Union(cmp.Compare, ys)

			if !IsEqual(cmp.Compare, actual, ys) {
				t.Fatal(cmp2.Diff(ys, actual))
			}
		})
	})

	t.Run("associativity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			xs := rapid.SliceOf(rapid.Int()).Draw(t, "xs")
			ys := rapid.SliceOf(rapid.Int()).Draw(t, "ys")
			zs := rapid.SliceOf(rapid.Int()).Draw(t, "zs")
			s1 := New(cmp.Compare, xs...)
			s2 := New(cmp.Compare, ys...)
			s3 := New(cmp.Compare, zs...)

			actual1 := s1.Clone()
			actual1.Union(cmp.Compare, s2)
			actual1.Union(cmp.Compare, s3)

			actual2 := s1.Clone()
			actual2_ := s2.Clone()
			actual2_.Union(cmp.Compare, s3)
			actual2.Union(cmp.Compare, actual2_)

			if !IsEqual(cmp.Compare, actual1, actual2) {
				t.Fatal(cmp2.Diff(actual1, actual2))
			}
		})
	})

	t.Run("commutativity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "xs")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "ys")...)

			actual1 := s1.Clone()
			actual1.Union(cmp.Compare, s2)

			actual2 := s2.Clone()
			actual2.Union(cmp.Compare, s1)

			if !IsEqual(cmp.Compare, actual1, actual2) {
				t.Fatal(cmp2.Diff(actual1, actual2))
			}
		})
	})

	t.Run("idempotence", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			actual := s.Clone()
			actual.Union(cmp.Compare, s)
			if !IsEqual(cmp.Compare, actual, s) {
				t.Fatal(cmp2.Diff(s, actual))
			}
		})
	})

	t.Run("union is a superset of both operands", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.Clone()
			actual.Union(cmp.Compare, s2)

			if !s1.IsSubsetOf(cmp.Compare, actual) {
				t.Fatal(cmp2.Diff(actual, s1))
			}
			if !s2.IsSubsetOf(cmp.Compare, actual) {
				t.Fatal(cmp2.Diff(actual, s2))
			}
		})
	})
}

func TestIntersection(t *testing.T) {
	testCases := map[string]struct {
		Input1   *Set[int]
		Input2   *Set[int]
		Expected *Set[int]
	}{
		"both empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New[int](cmp.Compare),
			Expected: New[int](cmp.Compare),
		},
		"empty and singleton": {
			Input1:   New[int](cmp.Compare),
			Input2:   New(cmp.Compare, 1),
			Expected: New[int](cmp.Compare),
		},
		"singleton and empty": {
			Input1:   New(cmp.Compare, 1),
			Input2:   New[int](cmp.Compare),
			Expected: New[int](cmp.Compare),
		},
		"with fsmcommon element": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 2, 3),
			Expected: New(cmp.Compare, 2),
		},
		"with no fsmcommon element": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 3, 4),
			Expected: New[int](cmp.Compare),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Input1.Clone()
			actual.Intersection(cmp.Compare, testCase.Input2)

			if !IsEqual(cmp.Compare, actual, testCase.Expected) {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("left absorption", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New[int](cmp.Compare)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)

			actual := s1.Clone()
			actual.Intersection(cmp.Compare, s2)

			if actual.Len() != 0 {
				t.Fatal(cmp2.Diff(New[int](cmp.Compare), actual))
			}
		})
	})

	t.Run("right absorption", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New[int](cmp.Compare)

			actual := s1.Clone()
			actual.Intersection(cmp.Compare, s2)

			if actual.Len() != 0 {
				t.Fatal(cmp2.Diff(New[int](cmp.Compare), actual))
			}
		})
	})

	t.Run("commutativity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)

			actual1 := s1.Clone()
			actual1.Intersection(cmp.Compare, s2)

			actual2 := s2.Clone()
			actual2.Intersection(cmp.Compare, s1)

			if !IsEqual(cmp.Compare, actual1, actual2) {
				t.Fatal(cmp2.Diff(actual1, actual2))
			}
		})
	})

	t.Run("idempotence", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			actual := s.Clone()
			actual.Intersection(cmp.Compare, s)

			if !IsEqual(cmp.Compare, actual, s) {
				t.Fatal(cmp2.Diff(s, actual))
			}
		})
	})

	t.Run("associativity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			s3 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s3")...)

			actual1 := s1.Clone()
			actual1.Intersection(cmp.Compare, s2)
			actual1.Intersection(cmp.Compare, s3)

			actual2 := s1.Clone()
			actual2_ := s2.Clone()
			actual2_.Intersection(cmp.Compare, s3)
			actual2.Intersection(cmp.Compare, actual2_)

			if !IsEqual(cmp.Compare, actual1, actual2) {
				t.Fatal(cmp2.Diff(actual1, actual2))
			}
		})
	})

	t.Run("intersection is a subset of both operands", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.Clone()
			actual.Intersection(cmp.Compare, s2)

			if !actual.IsSubsetOf(cmp.Compare, s1) {
				t.Fatal(cmp2.Diff(actual, s1))
			}
			if !actual.IsSubsetOf(cmp.Compare, s2) {
				t.Fatal(cmp2.Diff(actual, s2))
			}
		})
	})
}

func TestDifference(t *testing.T) {
	testCases := map[string]struct {
		Input1   *Set[int]
		Input2   *Set[int]
		Expected *Set[int]
	}{
		"both empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New[int](cmp.Compare),
			Expected: New[int](cmp.Compare),
		},
		"empty and not empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New(cmp.Compare, 1),
			Expected: New[int](cmp.Compare),
		},
		"not empty and empty": {
			Input1:   New(cmp.Compare, 1),
			Input2:   New[int](cmp.Compare),
			Expected: New(cmp.Compare, 1),
		},
		"with fsmcommon element": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 2, 3),
			Expected: New(cmp.Compare, 1),
		},
		"with no fsmcommon element": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 3, 4),
			Expected: New(cmp.Compare, 1, 2),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Input1.Clone()
			actual.Difference(cmp.Compare, testCase.Input2)

			if !IsEqual(cmp.Compare, actual, testCase.Expected) {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("left absorption", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New[int](cmp.Compare)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.Clone()
			actual.Difference(cmp.Compare, s2)

			if actual.Len() != 0 {
				t.Fatal(cmp2.Diff(s1, actual))
			}
		})
	})

	t.Run("right identity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New[int](cmp.Compare)
			actual := s1.Clone()
			actual.Difference(cmp.Compare, s2)

			if !IsEqual(cmp.Compare, actual, s1) {
				t.Fatal(cmp2.Diff(actual, s1))
			}
		})
	})

	t.Run("commutativity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			s3 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s3")...)
			actual1 := s1.Clone()
			actual1.Difference(cmp.Compare, s2)
			actual1.Difference(cmp.Compare, s3)

			actual2 := s1.Clone()
			actual2.Difference(cmp.Compare, s3)
			actual2.Difference(cmp.Compare, s2)

			if !IsEqual(cmp.Compare, actual1, actual2) {
				t.Fatal(cmp2.Diff(actual1, actual2))
			}
		})
	})

	t.Run("idempotence", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)

			actual1 := s1.Clone()
			actual1.Difference(cmp.Compare, s2)

			actual2 := s1.Clone()
			actual2.Difference(cmp.Compare, s2)
			actual2.Difference(cmp.Compare, s2)

			if !IsEqual(cmp.Compare, actual1, actual2) {
				t.Fatal(cmp2.Diff(actual1, actual2))
			}
		})
	})

	t.Run("difference is a subset of the first operand", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.Clone()
			actual.Difference(cmp.Compare, s2)

			if !actual.IsSubsetOf(cmp.Compare, s1) {
				t.Fatal(cmp2.Diff(actual, s1))
			}
		})
	})

	t.Run("difference is a disjoint with the second operand", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.Clone()
			actual.Difference(cmp.Compare, s2)

			if !actual.IsDisjointWith(cmp.Compare, s2) {
				t.Fatal(cmp2.Diff(actual, s2))
			}
		})
	})
}

func TestContains(t *testing.T) {
	testCases := map[string]struct {
		Input    *Set[int]
		Contains int
		Expected bool
	}{
		"empty": {
			Input:    New[int](cmp.Compare),
			Contains: 1,
			Expected: false,
		},
		"not empty": {
			Input:    New(cmp.Compare, 1),
			Contains: 1,
			Expected: true,
		},
		"not found": {
			Input:    New(cmp.Compare, 1),
			Contains: 2,
			Expected: false,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Input.Contains(cmp.Compare, testCase.Contains)
			if actual != testCase.Expected {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("empty", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New[int](cmp.Compare)
			actual := s.Contains(cmp.Compare, rapid.Int().Draw(t, "contains"))
			if actual {
				t.Fatalf("actual is true")
			}
		})
	})

	t.Run("after remove", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			x := rapid.Int().Draw(t, "x")
			s.Remove(cmp.Compare, x)

			actual := s.Contains(cmp.Compare, x)
			if actual {
				t.Fatalf("actual is true")
			}
		})
	})

	t.Run("after add", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			x := rapid.Int().Draw(t, "x")
			s.Add(cmp.Compare, x)

			actual := s.Contains(cmp.Compare, x)
			if !actual {
				t.Fatalf("actual is false")
			}
		})
	})
}

func TestIsSubsetOf(t *testing.T) {
	testCases := map[string]struct {
		Input1   *Set[int]
		Input2   *Set[int]
		Expected bool
	}{
		"both empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New[int](cmp.Compare),
			Expected: true,
		},
		"empty and not empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New(cmp.Compare, 1),
			Expected: true,
		},
		"not empty and empty": {
			Input1:   New(cmp.Compare, 1),
			Input2:   New[int](cmp.Compare),
			Expected: false,
		},
		"proper subset": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 1, 2, 3),
			Expected: true,
		},
		"equal": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 1, 2),
			Expected: true,
		},
		"disjoint": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 3, 4),
			Expected: false,
		},
		"disjoint (reflexive)": {
			Input1:   New(cmp.Compare, 3, 4),
			Input2:   New(cmp.Compare, 1, 2),
			Expected: false,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Input1.IsSubsetOf(cmp.Compare, testCase.Input2)
			if actual != testCase.Expected {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("empty", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New[int](cmp.Compare)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.IsSubsetOf(cmp.Compare, s2)

			if !actual && s2.Len() != 0 {
				t.Fatalf("actual is false")
			}
		})
	})

	t.Run("after add", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			s2 := s1.Clone()
			x := rapid.Int().Draw(t, "x")
			s2.Add(cmp.Compare, x)
			actual := s1.IsSubsetOf(cmp.Compare, s2)
			if !actual {
				t.Fatalf("actual is false")
			}
		})
	})

	t.Run("after remove", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			s2 := s1.Clone()
			x := rapid.Int().Draw(t, "x")
			s2.Remove(cmp.Compare, x)
			actual := s2.IsSubsetOf(cmp.Compare, s1)
			if !actual {
				t.Fatalf("actual is false")
			}
		})
	})

	t.Run("equal", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			actual := s.IsSubsetOf(cmp.Compare, s)
			if !actual {
				t.Fatalf("actual is false")
			}
		})
	})

	t.Run("antisymmetric", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)

			if s1.IsSubsetOf(cmp.Compare, s2) && s2.IsSubsetOf(cmp.Compare, s1) {
				if !IsEqual(cmp.Compare, s1, s2) {
					t.Fatalf("s1 and s2 are not equal")
				}
			}
		})
	})

	t.Run("antisymmetric contr", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)

			if !IsEqual(cmp.Compare, s1, s2) {
				if s1.IsSubsetOf(cmp.Compare, s2) && s2.IsSubsetOf(cmp.Compare, s1) {
					t.Fatalf("s1 ~= s2 -> ~(s1 <= s2 && s2 <= s1)")
				}
			}
		})
	})
}

func TestDisjointWith(t *testing.T) {
	testCases := map[string]struct {
		Input1   *Set[int]
		Input2   *Set[int]
		Expected bool
	}{
		"both empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New[int](cmp.Compare),
			Expected: true,
		},
		"empty and not empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New(cmp.Compare, 1),
			Expected: true,
		},
		"not empty and empty": {
			Input1:   New(cmp.Compare, 1),
			Input2:   New[int](cmp.Compare),
			Expected: true,
		},
		"disjoint": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 3, 4),
			Expected: true,
		},
		"not disjoint": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 2, 3),
			Expected: false,
		},
		"not disjoint reflexive": {
			Input1:   New(cmp.Compare, 2, 3),
			Input2:   New(cmp.Compare, 1, 2),
			Expected: false,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := testCase.Input1.IsDisjointWith(cmp.Compare, testCase.Input2)
			if actual != testCase.Expected {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("left empty", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New[int](cmp.Compare)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			actual := s1.IsDisjointWith(cmp.Compare, s2)
			if !actual {
				t.Fatalf("actual is false")
			}
		})
	})

	t.Run("commutative", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)

			actual1 := s1.IsDisjointWith(cmp.Compare, s2)
			actual2 := s2.IsDisjointWith(cmp.Compare, s1)

			if actual1 != actual2 {
				t.Fatalf("disjnt s1 s2 ~= disjnt s2 s1")
			}
		})
	})

	t.Run("add fsmcommon", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			x := rapid.Int().Draw(t, "x")
			s1.Add(cmp.Compare, x)
			s2.Add(cmp.Compare, x)
			actual := s1.IsDisjointWith(cmp.Compare, s2)
			if actual {
				t.Fatalf("disjnt s1 s2")
			}
		})
	})
}

func TestAdd(t *testing.T) {
	testCases := map[string]struct {
		Input    *Set[int]
		Add      int
		Expected *Set[int]
	}{
		"empty": {
			Input:    New[int](cmp.Compare),
			Add:      1,
			Expected: New(cmp.Compare, 1),
		},
		"greater than existing element": {
			Input:    New(cmp.Compare, 1),
			Add:      2,
			Expected: New(cmp.Compare, 1, 2),
		},
		"less than existing element": {
			Input:    New(cmp.Compare, 1),
			Add:      0,
			Expected: New(cmp.Compare, 0, 1),
		},
		"already exists": {
			Input:    New(cmp.Compare, 1),
			Add:      1,
			Expected: New(cmp.Compare, 1),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.Input.Add(cmp.Compare, testCase.Add)

			if !IsEqual(cmp.Compare, testCase.Input, testCase.Expected) {
				t.Error(cmp2.Diff(testCase.Expected, testCase.Input))
			}
		})
	}

	t.Run("not exist", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New[int](cmp.Compare)
			s2 := s1.Clone()

			x := rapid.Int().Draw(t, "x")
			s1.Add(cmp.Compare, x)

			if IsEqual(cmp.Compare, s1, s2) {
				t.Fatalf("s1 = s2")
			}
		})
	})

	t.Run("already exists", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			x := rapid.Int().Draw(t, "x")
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s1.Add(cmp.Compare, x)

			s2 := s1.Clone()
			s2.Add(cmp.Compare, x)

			if !IsEqual(cmp.Compare, s1, s2) {
				t.Fatalf("s1 ~= s2")
			}
		})
	})
}

func TestRemove(t *testing.T) {
	testCases := map[string]struct {
		Input    *Set[int]
		Remove   int
		Expected *Set[int]
	}{
		"empty": {
			Input:    New[int](cmp.Compare),
			Remove:   1,
			Expected: New[int](cmp.Compare),
		},
		"not exists": {
			Input:    New(cmp.Compare, 1),
			Remove:   2,
			Expected: New(cmp.Compare, 1),
		},
		"exists": {
			Input:    New(cmp.Compare, 1, 2, 3),
			Remove:   2,
			Expected: New(cmp.Compare, 1, 3),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.Input.Remove(cmp.Compare, testCase.Remove)

			if !IsEqual(cmp.Compare, testCase.Input, testCase.Expected) {
				t.Error(cmp2.Diff(testCase.Expected, testCase.Input))
			}
		})
	}

	t.Run("not exist", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.IntRange(1, 100)).Draw(t, "s1")...)
			s2 := s1.Clone()
			x := rapid.IntRange(-100, -1).Draw(t, "x")
			s2.Remove(cmp.Compare, x)

			if !IsEqual(cmp.Compare, s1, s2) {
				t.Fatalf("s ~= {}")
			}
		})
	})

	t.Run("add and remove identity", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.IntRange(1, 100)).Draw(t, "s1")...)
			x := rapid.IntRange(-100, -1).Draw(t, "x")
			s2 := s1.Clone()
			s2.Add(cmp.Compare, x)
			s2.Remove(cmp.Compare, x)

			if !IsEqual(cmp.Compare, s1, s2) {
				t.Fatalf("s1 ~= s2")
			}
		})
	})
}

func TestIsEqual(t *testing.T) {
	testCases := map[string]struct {
		Input1   *Set[int]
		Input2   *Set[int]
		Expected bool
	}{
		"both empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New[int](cmp.Compare),
			Expected: true,
		},
		"empty and not empty": {
			Input1:   New[int](cmp.Compare),
			Input2:   New(cmp.Compare, 1),
			Expected: false,
		},
		"not empty and empty": {
			Input1:   New(cmp.Compare, 1),
			Input2:   New[int](cmp.Compare),
			Expected: false,
		},
		"equal": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 1, 2),
			Expected: true,
		},
		"not equal": {
			Input1:   New(cmp.Compare, 1, 2),
			Input2:   New(cmp.Compare, 1, 3),
			Expected: false,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := IsEqual(cmp.Compare, testCase.Input1, testCase.Input2)
			if actual != testCase.Expected {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("reflexive", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s")...)
			actual := IsEqual(cmp.Compare, s, s)
			if !actual {
				t.Fatalf("s ~= s")
			}
		})
	})

	t.Run("symmetric", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)

			if IsEqual(cmp.Compare, s1, s2) != IsEqual(cmp.Compare, s2, s1) {
				t.Fatalf("(s1 ~= s2) ~= (s2 ~= s1)")
			}
		})
	})

	t.Run("transitive", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s1 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s1")...)
			s2 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s2")...)
			s3 := New(cmp.Compare, rapid.SliceOf(rapid.Int()).Draw(t, "s3")...)

			if IsEqual(cmp.Compare, s1, s2) && IsEqual(cmp.Compare, s2, s3) && !IsEqual(cmp.Compare, s1, s3) {
				t.Fatalf("s1 ~= s2 ~= s3 -> s1 ~= s3")
			}
		})
	})
}

func TestPowerSet(t *testing.T) {
	testCases := map[string]struct {
		Input    *Set[int]
		Expected *Set[*Set[int]]
	}{
		"empty": {
			Input:    New[int](cmp.Compare),
			Expected: New(Compare[int](cmp.Compare), New[int](cmp.Compare)),
		},
		"not empty": {
			Input: New(cmp.Compare, 1, 2, 3),
			Expected: New(
				Compare[int](cmp.Compare),
				New[int](cmp.Compare),
				New(cmp.Compare, 1),
				New(cmp.Compare, 2),
				New(cmp.Compare, 3),
				New(cmp.Compare, 1, 2),
				New(cmp.Compare, 1, 3),
				New(cmp.Compare, 2, 3),
				New(cmp.Compare, 1, 2, 3),
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			actual := PowerSet(testCase.Input, cmp.Compare)
			if !IsEqual(Compare[int](cmp.Compare), actual, testCase.Expected) {
				t.Error(cmp2.Diff(testCase.Expected, actual))
			}
		})
	}

	t.Run("allcheckers subset", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOfN(rapid.Int(), 0, 6).Draw(t, "s")...)

			actual := PowerSet(s, cmp.Compare)
			for _, subset := range actual.Iter() {
				if !subset.IsSubsetOf(cmp.Compare, s) {
					t.Fatalf("subset <= s")
				}
			}
		})
	})

	t.Run("cardinality", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			s := New(cmp.Compare, rapid.SliceOfN(rapid.Int(), 0, 6).Draw(t, "s")...)
			actual := PowerSet(s, cmp.Compare)
			if actual.Len() != 1<<uint(s.Len()) {
				t.Fatalf("|P(s)| = 2^|s|")
			}
		})
	})
}
