package reflectlite_test

import (
	"internal/reflectlite"
	"testing"
)

var (
	fieldSink reflectlite.StructField
	nameSink  string
)

func BenchmarkField(b *testing.B) {
	typ := reflectlite.TypeOf(fields{}).(*reflectlite.RawType)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fieldSink = typ.Field(i % 4)
	}
}

func BenchmarkFieldName(b *testing.B) {
	typ := reflectlite.TypeOf(fields{}).(*reflectlite.RawType)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		nameSink = typ.Field(i % 4).Name
	}
}
