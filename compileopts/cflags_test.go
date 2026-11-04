package compileopts

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestDirectoryPrefixMaps(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "file",
			got:  FilePrefixMap(filepath.Join("tmp", "source"), "/_/module"),
			want: "-ffile-prefix-map=" + filepath.Join("tmp", "source") + string(filepath.Separator) + "=" + filepath.ToSlash(filepath.Join("/_", "module")) + "/",
		},
		{
			name: "debug",
			got:  DebugPrefixMap(string(filepath.Separator), "/_/"),
			want: "-fdebug-prefix-map=" + string(filepath.Separator) + "=" + filepath.ToSlash(filepath.Join("/_")) + "/",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("prefix map = %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestSortPrefixMaps(t *testing.T) {
	maps := []string{
		FilePrefixMap(filepath.Join("root", "module"), "/_/module"),
		FilePrefixMap("root", "/_/root"),
		DebugPrefixMap("root", "/_/root"),
	}
	SortPrefixMaps(maps)
	want := []string{
		FilePrefixMap("root", "/_/root"),
		DebugPrefixMap("root", "/_/root"),
		FilePrefixMap(filepath.Join("root", "module"), "/_/module"),
	}
	if !slices.Equal(maps, want) {
		t.Fatalf("sorted prefix maps = %q, want %q", maps, want)
	}
}
