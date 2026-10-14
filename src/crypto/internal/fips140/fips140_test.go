package fips140

import "testing"

func TestUnsupported(t *testing.T) {
	if Enabled {
		t.Fatal("FIPS mode is enabled")
	}
	if err := Supported(); err == nil || err.Error() != unsupported {
		t.Fatalf("Supported() = %v, want an unsupported-mode error", err)
	}
}

func TestSelfTests(t *testing.T) {
	for name, test := range map[string]func(string, func() error){
		"CAST": CAST,
		"PCT":  PCT,
	} {
		t.Run(name, func(t *testing.T) {
			test("valid", func() error {
				t.Fatal("self-test callback was called")
				return nil
			})
			for _, invalid := range []string{"comma,", "hash#", "equals=", "colon:"} {
				t.Run(invalid, func(t *testing.T) {
					defer func() {
						if got := recover(); got != "fips: invalid self-test name: "+invalid {
							t.Errorf("unexpected panic for %q: %v", invalid, got)
						}
					}()
					test(invalid, nil)
				})
			}
		})
	}
}

func TestServiceIndicator(t *testing.T) {
	ResetServiceIndicator()
	defer ResetServiceIndicator()
	if ServiceIndicator() {
		t.Fatal("service indicator is set after reset")
	}
	RecordApproved()
	if !ServiceIndicator() {
		t.Fatal("approved service was not recorded")
	}
	RecordNonApproved()
	RecordApproved()
	if ServiceIndicator() {
		t.Fatal("approved service replaced a non-approved service")
	}
}
