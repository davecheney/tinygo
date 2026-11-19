package compileopts

import (
	"encoding/json"
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
			want: "-ffile-prefix-map=" + filepath.Join("tmp", "source") + string(filepath.Separator) + "=/_/module/",
		},
		{
			name: "debug",
			got:  DebugPrefixMap(string(filepath.Separator), "/_/"),
			want: "-fdebug-prefix-map=" + string(filepath.Separator) + "=/_/",
		},
		{
			name: "existing-boundaries",
			got:  FilePrefixMap(filepath.Join("tmp", "source")+string(filepath.Separator), "/_/module/"),
			want: "-ffile-prefix-map=" + filepath.Join("tmp", "source") + string(filepath.Separator) + "=/_/module/",
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

func TestCFlagsJSON(t *testing.T) {
	var flags CFlags
	in := `["-Werror", "-march=rv32imc", "-DFOO=1", "-I{root}/lib/nrfx/mdk"]`
	if err := json.Unmarshal([]byte(in), &flags); err != nil {
		t.Fatal(err)
	}
	if flags["-march"] != "rv32imc" || flags["-DFOO"] != "1" {
		t.Errorf("unexpected flags %v", flags)
	}
	want := []string{"-DFOO=1", "-I{root}/lib/nrfx/mdk", "-Werror", "-march=rv32imc"}
	if got := flags.Args(); !slices.Equal(got, want) {
		t.Errorf("Args() = %v, want %v", got, want)
	}
	out, err := json.Marshal(flags)
	if err != nil {
		t.Fatal(err)
	}
	if want := `["-DFOO=1","-I{root}/lib/nrfx/mdk","-Werror","-march=rv32imc"]`; string(out) != want {
		t.Errorf("Marshal = %s, want %s", out, want)
	}

	for _, in := range []string{`["-march=a", "-march=b"]`, `["-target", "aarch64"]`} {
		if err := json.Unmarshal([]byte(in), &flags); err == nil {
			t.Errorf("Unmarshal(%s) did not fail", in)
		}
	}
}
