// patching provides functionality to patch resources
package patching

import (
	"fmt"
	"testing"
	"time"

	"github.com/cccteam/ccc"
)

type Int int

type Stringer int

func (s Stringer) String() string {
	return fmt.Sprintf("%d", s)
}

type Stringer2 Stringer

type Marshaler struct {
	field string
}

func (m Marshaler) MarshalText() ([]byte, error) {
	return []byte(m.field), nil
}

type Marshaler2 Marshaler

func Test_match(t *testing.T) {
	t.Parallel()

	Time := time.Date(2032, 4, 23, 12, 2, 3, 4, time.UTC)
	Time2 := Time.Add(time.Hour)

	type args struct {
		v  any
		v2 any
	}
	tests := []struct {
		name        string
		args        args
		wantMatched bool
		wantErr     bool
	}{
		{name: "primitive matched int", args: args{v: int(1), v2: int(1)}, wantMatched: true},
		{name: "primitive matched int8", args: args{v: int8(1), v2: int8(1)}, wantMatched: true},
		{name: "primitive matched int16", args: args{v: int16(1), v2: int16(1)}, wantMatched: true},
		{name: "primitive matched int32", args: args{v: int32(1), v2: int32(1)}, wantMatched: true},
		{name: "primitive matched int64", args: args{v: int64(1), v2: int64(1)}, wantMatched: true},
		{name: "primitive matched uint", args: args{v: uint(1), v2: uint(1)}, wantMatched: true},
		{name: "primitive matched uint8", args: args{v: uint8(1), v2: uint8(1)}, wantMatched: true},
		{name: "primitive matched uint16", args: args{v: uint16(1), v2: uint16(1)}, wantMatched: true},
		{name: "primitive matched uint32", args: args{v: uint32(1), v2: uint32(1)}, wantMatched: true},
		{name: "primitive matched uint64", args: args{v: uint64(1), v2: uint64(1)}, wantMatched: true},
		{name: "primitive matched float32", args: args{v: float32(1), v2: float32(1)}, wantMatched: true},
		{name: "primitive matched float64", args: args{v: float64(1), v2: float64(1)}, wantMatched: true},
		{name: "primitive matched string", args: args{v: "1", v2: "1"}, wantMatched: true},
		{name: "primitive matched bool", args: args{v: true, v2: true}, wantMatched: true},
		{name: "primitive matched *int", args: args{v: ccc.Ptr(int(1)), v2: ccc.Ptr(int(1))}, wantMatched: true},
		{name: "primitive matched *int8", args: args{v: ccc.Ptr(int8(1)), v2: ccc.Ptr(int8(1))}, wantMatched: true},
		{name: "primitive matched *int16", args: args{v: ccc.Ptr(int16(1)), v2: ccc.Ptr(int16(1))}, wantMatched: true},
		{name: "primitive matched *int32", args: args{v: ccc.Ptr(int32(1)), v2: ccc.Ptr(int32(1))}, wantMatched: true},
		{name: "primitive matched *int64", args: args{v: ccc.Ptr(int64(1)), v2: ccc.Ptr(int64(1))}, wantMatched: true},
		{name: "primitive matched *uint", args: args{v: ccc.Ptr(uint(1)), v2: ccc.Ptr(uint(1))}, wantMatched: true},
		{name: "primitive matched *uint8", args: args{v: ccc.Ptr(uint8(1)), v2: ccc.Ptr(uint8(1))}, wantMatched: true},
		{name: "primitive matched *uint16", args: args{v: ccc.Ptr(uint16(1)), v2: ccc.Ptr(uint16(1))}, wantMatched: true},
		{name: "primitive matched *uint32", args: args{v: ccc.Ptr(uint32(1)), v2: ccc.Ptr(uint32(1))}, wantMatched: true},
		{name: "primitive matched *uint64", args: args{v: ccc.Ptr(uint64(1)), v2: ccc.Ptr(uint64(1))}, wantMatched: true},
		{name: "primitive matched *float32", args: args{v: ccc.Ptr(float32(1)), v2: ccc.Ptr(float32(1))}, wantMatched: true},
		{name: "primitive matched *float64", args: args{v: ccc.Ptr(float64(1)), v2: ccc.Ptr(float64(1))}, wantMatched: true},
		{name: "primitive matched *string", args: args{v: ccc.Ptr("1"), v2: ccc.Ptr("1")}, wantMatched: true},
		{name: "primitive matched *bool", args: args{v: ccc.Ptr(true), v2: ccc.Ptr(true)}, wantMatched: true},
		{name: "primitive not matched", args: args{v: 1, v2: 4}, wantMatched: false},

		{name: "named matched", args: args{v: Int(1), v2: Int(1)}, wantMatched: true},
		{name: "named not matched", args: args{v: Int(1), v2: Int(4)}, wantMatched: false},

		{name: "marshaler matched", args: args{v: Marshaler{field: "1"}, v2: Marshaler{field: "1"}}, wantMatched: true},
		{name: "marshaler not matched", args: args{v: Marshaler{field: "1"}, v2: Marshaler{"4"}}, wantMatched: false},
		{name: "marshaler error", args: args{v: Marshaler{field: "1"}, v2: Marshaler2{"1"}}, wantErr: true},

		{name: "time.Time matched", args: args{v: Time, v2: Time}, wantMatched: true},
		{name: "time.Time not matched", args: args{v: Time, v2: Time2}, wantMatched: false},

		{name: "[]time.Time matched", args: args{v: []time.Time{Time, Time2}, v2: []time.Time{Time, Time2}}, wantMatched: true},
		{name: "[]time.Time not matched", args: args{v: []time.Time{Time, Time2}, v2: []time.Time{Time, Time}}, wantMatched: false},

		{name: "stringer matched", args: args{v: Stringer(1), v2: Stringer(1)}, wantMatched: true},
		{name: "stringer not matched", args: args{v: Stringer(1), v2: Stringer(4)}, wantMatched: false},
		{name: "stringer error", args: args{v: Stringer(1), v2: Stringer2(1)}, wantErr: true},

		{name: "different types error", args: args{v: Int(1), v2: 1}, wantErr: true},

		{name: "[]any matched", args: args{v: []any{1, 5}, v2: []any{1, 5}}, wantMatched: true},
		{name: "[]any slices not matched", args: args{v: []any{1, 5}, v2: []any{4, 5}}, wantMatched: false},

		{name: "[]int matched", args: args{v: []int{1, 5}, v2: []int{1, 5}}, wantMatched: true},
		{name: "[]int not matched", args: args{v: []int{1, 5}, v2: []int{4, 5}}, wantMatched: false},

		{name: "[]*int matched", args: args{v: []*int{ccc.Ptr(1), ccc.Ptr(5)}, v2: []*int{ccc.Ptr(1), ccc.Ptr(5)}}, wantMatched: true},
		{name: "[]*int not matched", args: args{v: []*int{ccc.Ptr(1), ccc.Ptr(5)}, v2: []*int{ccc.Ptr(4), ccc.Ptr(5)}}, wantMatched: false},

		{name: "[]int8 matched", args: args{v: []int8{1, 5}, v2: []int8{1, 5}}, wantMatched: true},
		{name: "[]int8 not matched", args: args{v: []int8{1, 5}, v2: []int8{4, 5}}, wantMatched: false},

		{name: "[]Int matched", args: args{v: []Int{1, 5}, v2: []Int{1, 5}}, wantMatched: true},
		{name: "[]Int not matched", args: args{v: []Int{1, 5}, v2: []Int{4, 5}}, wantMatched: false},

		{name: "*[]Int matched", args: args{v: &[]Int{1, 5}, v2: &[]Int{1, 5}}, wantMatched: true},
		{name: "*[]Int not matched", args: args{v: &[]Int{1, 5}, v2: &[]Int{4, 5}}, wantMatched: false},

		{name: "ccc.UUID matched", args: args{v: ccc.Must(ccc.UUIDFromString("a517b48d-63a9-4c1f-b45b-8474b164e423")), v2: ccc.Must(ccc.UUIDFromString("a517b48d-63a9-4c1f-b45b-8474b164e423"))}, wantMatched: true},
		{name: "ccc.UUID not matched", args: args{v: ccc.Must(ccc.UUIDFromString("a517b48d-63a9-4c1f-b45b-8474b164e423")), v2: ccc.Must(ccc.UUIDFromString("B517b48d-63a9-4c1f-b45b-8474b164e423"))}, wantMatched: false},

		{name: "*ccc.UUID matched", args: args{v: ccc.Ptr(ccc.Must(ccc.UUIDFromString("a517b48d-63a9-4c1f-b45b-8474b164e423"))), v2: ccc.Ptr(ccc.Must(ccc.UUIDFromString("a517b48d-63a9-4c1f-b45b-8474b164e423")))}, wantMatched: true},
		{name: "*ccc.UUID matched", args: args{v: ccc.Ptr(ccc.Must(ccc.UUIDFromString("a517b48d-63a9-4c1f-b45b-8474b164e423"))), v2: ccc.Ptr(ccc.Must(ccc.UUIDFromString("B517b48d-63a9-4c1f-b45b-8474b164e423")))}, wantMatched: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotMatched, err := match(tt.args.v, tt.args.v2)
			if (err != nil) != tt.wantErr {
				t.Errorf("match() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatched != tt.wantMatched {
				t.Errorf("match() = %v, want %v", gotMatched, tt.wantMatched)
			}
		})
	}
}

func TestPatcher_Spanner_Columns(t *testing.T) {
	t.Parallel()
	type SpannerStruct struct {
		Field1 string `spanner:"field1"`
		Field2 string `spanner:"fieldtwo"`
		Field3 int    `spanner:"field3"`
		Field5 string `spanner:"field5"`
		Field4 string `spanner:"field4"`
	}

	tm := NewSpannerPatcher()

	type args struct {
		patchSet     *PatchSet
		databaseType any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "multiple fields in patchSet",
			args: args{
				patchSet: NewPatchSet(map[string]any{
					"Field2": "apple",
					"Field3": 10,
				}),
				databaseType: SpannerStruct{},
			},
			want: "fieldtwo, field3",
		},
		{
			name: "multiple fields not in sorted order",
			args: args{
				patchSet: NewPatchSet(map[string]any{
					"Field4": "apple",
					"Field5": "bannana",
				}),
				databaseType: SpannerStruct{},
			},
			want: "field5, field4",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tm.Columns(tt.args.patchSet, tt.args.databaseType); got != tt.want {
				t.Errorf("Patcher.Columns() = (%v),  want (%v)", got, tt.want)
			}
		})
	}
}

func TestPatcher_Postgres_Columns(t *testing.T) {
	t.Parallel()
	type SpannerStruct struct {
		Field1 string `db:"field1"`
		Field2 string `db:"fieldtwo"`
		Field3 int    `db:"field3"`
		Field5 string `db:"field5"`
		Field4 string `db:"field4"`
	}

	tm := NewPostgresPatcher()

	type args struct {
		patchSet     *PatchSet
		databaseType any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "multiple fields in patchSet",
			args: args{
				patchSet: NewPatchSet(map[string]any{
					"Field2": "apple",
					"Field3": 10,
				}),
				databaseType: SpannerStruct{},
			},
			want: `"fieldtwo", "field3"`,
		},
		{
			name: "multiple fields not in sorted order",
			args: args{
				patchSet: NewPatchSet(map[string]any{
					"Field4": "apple",
					"Field5": "bannana",
				}),
				databaseType: SpannerStruct{},
			},
			want: `"field5", "field4"`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tm.Columns(tt.args.patchSet, tt.args.databaseType); got != tt.want {
				t.Errorf("Patcher.Columns() = (%v),  want (%v)", got, tt.want)
			}
		})
	}
}
