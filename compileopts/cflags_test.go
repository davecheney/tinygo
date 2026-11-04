package compileopts

import (
	"slices"
	"testing"
)

func TestSortPrefixMaps(t *testing.T) {
	tests := []struct {
		name string
		maps []string
		want []string
	}{
		{
			name: "overlapping paths",
			maps: []string{
				"-ffile-prefix-map=/bc=cache",
				"-ffile-prefix-map=/b=module",
			},
			want: []string{
				"-ffile-prefix-map=/b=module",
				"-ffile-prefix-map=/bc=cache",
			},
		},
		{
			name: "mixed flags and equal lengths",
			maps: []string{
				"-fdebug-prefix-map=/long=first",
				"-ffile-prefix-map=/a=file",
				"-fmacro-prefix-map=/b=macro",
				"-ffile-prefix-map=/long=second",
			},
			want: []string{
				"-ffile-prefix-map=/a=file",
				"-fmacro-prefix-map=/b=macro",
				"-fdebug-prefix-map=/long=first",
				"-ffile-prefix-map=/long=second",
			},
		},
		{
			name: "equals in destination",
			maps: []string{
				"-ffile-prefix-map=/ab=short",
				"-ffile-prefix-map=/a=destination=with=equals",
			},
			want: []string{
				"-ffile-prefix-map=/a=destination=with=equals",
				"-ffile-prefix-map=/ab=short",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			SortPrefixMaps(test.maps)
			if !slices.Equal(test.maps, test.want) {
				t.Errorf("SortPrefixMaps() = %q, want %q", test.maps, test.want)
			}
		})
	}
}
