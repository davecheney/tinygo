package reflectlite_test

import (
	"internal/reflectlite"
	"testing"
	"unsafe"
)

type Embedded int

type fields struct {
	_     [130]byte `first:"blank"`
	Value byte
	last  byte `last:"field"`
	Embedded
}

func TestFieldMetadata(t *testing.T) {
	var value fields
	typ := reflectlite.TypeOf(value).(*reflectlite.RawType)
	for i, want := range []struct {
		name      string
		pkgPath   string
		typ       reflectlite.Type
		tag       reflectlite.StructTag
		offset    uintptr
		anonymous bool
	}{
		{"_", typ.PkgPath(), reflectlite.TypeOf([130]byte{}), `first:"blank"`, 0, false},
		{"Value", "", reflectlite.TypeOf(byte(0)), "", unsafe.Offsetof(value.Value), false},
		{"last", typ.PkgPath(), reflectlite.TypeOf(byte(0)), `last:"field"`, unsafe.Offsetof(value.last), false},
		{"Embedded", "", reflectlite.TypeOf(Embedded(0)), "", unsafe.Offsetof(value.Embedded), true},
	} {
		got := typ.Field(i)
		if got.Name != want.name || got.PkgPath != want.pkgPath || got.Type != want.typ || got.Tag != want.tag || got.Offset != want.offset || got.Anonymous != want.anonymous {
			t.Errorf("Field(%d) = %+v, want %+v", i, got, want)
		}
		if len(got.Index) != 1 || got.Index[0] != i {
			t.Errorf("Field(%d).Index = %v, want [%d]", i, got.Index, i)
		}
	}
}

func TestFieldPanics(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		index int
	}{
		{"not struct", 0, 0},
		{"empty struct", struct{}{}, 0},
		{"negative index", fields{}, -1},
		{"large index", fields{}, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("Field did not panic")
				}
			}()
			reflectlite.TypeOf(test.value).(*reflectlite.RawType).Field(test.index)
		})
	}
}
