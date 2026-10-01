package compileopts

import (
	"slices"
	"testing"
)

func TestExtraFilesBoehm(t *testing.T) {
	target := &TargetSpec{
		GC: "precise",
		ExtraFiles: []string{
			"src/runtime/asm_tinygowasm.S",
		},
	}
	config := &Config{
		Options: &Options{},
		Target:  target,
	}

	got := config.ExtraFiles()
	want := []string{"src/runtime/asm_tinygowasm.S"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected precise GC files: got %v, want %v", got, want)
	}

	config.Options.GC = "boehm"
	got = config.ExtraFiles()
	want = []string{
		"src/runtime/asm_tinygowasm.S",
		"src/runtime/gc_boehm.c",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected Boehm GC files: got %v, want %v", got, want)
	}
}

func TestRuntimeStress(t *testing.T) {
	for _, gc := range []string{"none", "leaking", "custom", "conservative", "precise", "boehm"} {
		for _, tag := range []string{"runtime_gcstress", "runtime_clobberfree"} {
			for _, override := range []bool{false, true} {
				config := &Config{Options: &Options{}, Target: &TargetSpec{GC: gc, BuildTags: []string{tag}}}
				if override {
					config.Target.GC = "none"
					config.Options.GC = gc
					config.Target.BuildTags = nil
					config.Options.Tags = []string{tag}
				}
				wantOK := gc == "conservative" || gc == "precise" || gc == "boehm" && tag == "runtime_gcstress"
				if err := config.VerifyRuntimeStress(); (err == nil) != wantOK {
					t.Errorf("gc=%s tag=%s override=%t: %v", gc, tag, override, err)
				}
			}
		}
	}
	config := &Config{Options: &Options{}, Target: &TargetSpec{GC: "conservative", Scheduler: "cores", BuildTags: []string{"runtime_gcstress"}}}
	if err := config.VerifyRuntimeStress(); err == nil {
		t.Error("runtime_gcstress accepted -scheduler=cores")
	}
}
