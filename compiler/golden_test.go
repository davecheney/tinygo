package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoldenFilePath(t *testing.T) {
	tests := []struct {
		name       string
		version    int
		legacy     bool
		wantLegacy bool
	}{
		{"shared LLVM 22", 22, false, false},
		{"shared LLVM 23", 23, false, false},
		{"shared future LLVM", 24, false, false},
		{"legacy LLVM 20", 20, true, true},
		{"legacy LLVM 21", 21, true, true},
		{"legacy LLVM 22", 22, true, true},
		{"current LLVM 23", 23, true, false},
		{"future LLVM", 24, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "golden.ll")
			if err := os.WriteFile(path, []byte("current fixture"), 0o600); err != nil {
				t.Fatal(err)
			}
			legacyPath := filepath.Join(filepath.Dir(path), "llvm22", filepath.Base(path))
			if tc.legacy {
				if err := os.Mkdir(filepath.Dir(legacyPath), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(legacyPath, []byte("legacy fixture"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := goldenFilePath(path, tc.version)
			if err != nil {
				t.Fatal(err)
			}
			want := path
			if tc.wantLegacy {
				want = legacyPath
			}
			if got != want {
				t.Errorf("golden path = %q, want %q", got, want)
			}
		})
	}
}

func TestGoldenFilePathError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid\x00.ll")
	if _, err := goldenFilePath(path, 22); err == nil {
		t.Fatal("expected an error for an invalid override path")
	}
}

func TestGoldenIRDiff(t *testing.T) {
	floatIR := func(value string) string {
		return "define float @f() {\n  ret float " + value + "\n}\n"
	}
	gepIR := func(element, flags string) string {
		return `target datalayout = "e-p:32:32"
define ptr @f(ptr %base, i32 %index) {
  %result = getelementptr ` + flags + " " + element + `, ptr %base, i32 %index
  ret ptr %result
}
`
	}
	attributes := "declare void @f() #7\nattributes #7 = { nounwind }\n"
	duplicates := "declare void @f() #7\ndeclare void @g() #8\nattributes #7 = { nounwind }\nattributes #8 = { nounwind }\n"
	calls := `declare i32 @one()
declare i32 @two()
define i32 @f() {
  %result = call i32 @one()
  ret i32 %result
}
`
	branches := `define i32 @f(i1 %condition) {
  br i1 %condition, label %then, label %else
then:
  ret i32 1
else:
  ret i32 2
}
`
	tests := []struct {
		name     string
		expected string
		actual   string
		equal    bool
	}{
		{
			name:     "identical IR",
			expected: floatIR("f0x4F800000"),
			actual:   floatIR("f0x4F800000"),
			equal:    true,
		},
		{
			name:     "float spelling",
			expected: floatIR("0x41F0000000000000"),
			actual:   floatIR("4.294967296000000e+09"),
		},
		{
			name:     "float32 bit spelling",
			expected: floatIR("0x41F0000000000000"),
			actual:   floatIR("f0x4F800000"),
		},
		{
			name:     "GEP element spelling",
			expected: gepIR("i32", "inbounds"),
			actual:   gepIR("[4 x i8]", "inbounds"),
		},
		{
			name:     "attribute numbering",
			expected: attributes,
			actual:   strings.ReplaceAll(attributes, "#7", "#42"),
		},
		{
			name:     "duplicate attribute groups",
			expected: duplicates,
			actual:   "declare void @f() #7\ndeclare void @g() #7\nattributes #7 = { nounwind }\n",
		},
		{
			name:     "float value",
			expected: floatIR("0x41F0000000000000"),
			actual:   floatIR("0x41EFFFFFC0000000"),
		},
		{
			name:     "GEP stride",
			expected: gepIR("i32", "inbounds"),
			actual:   gepIR("i64", "inbounds"),
		},
		{
			name:     "GEP flags",
			expected: gepIR("i32", "inbounds"),
			actual:   gepIR("i32", ""),
		},
		{
			name:     "call target",
			expected: calls,
			actual:   strings.Replace(calls, "call i32 @one()", "call i32 @two()", 1),
		},
		{
			name:     "branch targets",
			expected: branches,
			actual:   strings.Replace(branches, "label %then, label %else", "label %else, label %then", 1),
		},
		{
			name:     "memory effects",
			expected: "declare void @f(ptr) readonly\n",
			actual:   "declare void @f(ptr) writeonly\n",
		},
		{
			name:     "parameter attributes",
			expected: "declare void @f(ptr readonly)\n",
			actual:   "declare void @f(ptr writeonly)\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diff := diffIR(tc.expected, tc.actual)
			if (diff == "") != tc.equal {
				t.Fatalf("expected equal=%v, got diff:\n%s", tc.equal, diff)
			}
		})
	}
}
