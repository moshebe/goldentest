package goldentest

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Bar struct {
	Bla string
	Env map[string]string
}
type Foo struct {
	Name   string
	Values []int
	Barbi  Bar
}

func TestCompare(t *testing.T) {
	got := Foo{
		Name:   "bla",
		Values: []int{1, 2, 3},
		Barbi: Bar{
			Bla: "blabla",
			Env: map[string]string{
				"ACCOUNT": "1234",
				"ID":      "4321",
				"OTHER":   "STUFF",
			},
		},
	}

	g := New[Foo](path.Join("testdata", "single.golden.json")).WithIgnoreFields("Barbi.Bla")
	res, err := g.Compare(&got)
	require.NoError(t, err)
	require.Empty(t, res.Diff)
}

func TestUpdate(t *testing.T) {
	got := Foo{
		Name:   "newvalue",
		Values: []int{7},
		Barbi: Bar{
			Bla: "bldsabldsa",
		},
	}

	g := New[Foo](path.Join("testdata", "updated.golden.json")).WithIgnoreFields("Barbi.Bla")
	err := g.Update(got)
	require.NoError(t, err)
}

func TestCompareProtoJSON(t *testing.T) {
	got := &pb.Span{
		Name:      "John",
		SpanId:    "1337",
		StartTime: timestamppb.New(time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC)),
	}

	g := New[pb.Span](path.Join("testdata", "updated.golden.proto.json")).WithEncoder(ProtoJSONEncoder{})
	res, err := g.Compare(got)
	require.NoError(t, err)
	require.Empty(t, res.Diff)
}

func TestUpdateProtoJSON(t *testing.T) {
	got := &pb.Span{
		Name:      "John",
		SpanId:    "1337",
		StartTime: timestamppb.New(time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC)),
	}

	g := New[*pb.Span](path.Join("testdata", "updated.golden.proto.json")).WithEncoder(ProtoJSONEncoder{})
	err := g.Update(got)
	require.NoError(t, err)
}

func TestCompareValues(t *testing.T) {
	got := []Foo{
		{
			Name:   "bla1",
			Values: []int{1, 2, 3},
			Barbi: Bar{
				Bla: "blabla",
				Env: map[string]string{
					"ACCOUNT": "1234",
					"ID":      "4321",
					"OTHER":   "STUFF",
				},
			},
		},
		{
			Name:   "bla2",
			Values: []int{1},
			Barbi: Bar{
				Bla: "XXX",
				Env: map[string]string{
					"ACCOUNT": "1234",
					"OTHER":   "ABCD",
				},
			},
		},
	}

	g := New[Foo](path.Join("testdata", "update_multiple.golden.json")).
		WithIgnoreFields("Barbi.Bla")
	res, err := g.CompareValues(got)
	require.NoError(t, err)
	require.Empty(t, res)

}

func TestUpdateValues(t *testing.T) {
	got := []Foo{
		{
			Name:   "bla1",
			Values: []int{1, 2, 3},
			Barbi: Bar{
				Bla: "blabla",
				Env: map[string]string{
					"ACCOUNT": "1234",
					"ID":      "4321",
					"OTHER":   "STUFF",
				},
			},
		},
		{
			Name:   "bla2",
			Values: []int{1},
			Barbi: Bar{
				Bla: "XXX",
				Env: map[string]string{
					"ACCOUNT": "1234",
					"OTHER":   "ABCD",
				},
			},
		},
	}

	g := New[Foo](path.Join("testdata", "update_multiple.golden.json")).WithIgnoreFields("Barbi.Bla")
	err := g.UpdateValues(got)
	require.NoError(t, err)
}

func TestCompareValuesProto(t *testing.T) {
	got := []*pb.Span{
		{
			Name:   "First",
			SpanId: "1",
		},
		{
			Name:   "Second",
			SpanId: "2",
		},
	}

	g := New[*pb.Span](path.Join("testdata", "update_multiple.golden.proto.json")).WithEncoder(ProtoJSONEncoder{})
	res, err := g.CompareValues(got)
	require.NoError(t, err)
	require.Empty(t, res)
}

func TestUpdateValuesProto(t *testing.T) {
	got := []*pb.Span{
		{
			Name:   "First",
			SpanId: "1",
		},
		{
			Name:   "Second",
			SpanId: "2",
		},
	}

	g := New[*pb.Span](path.Join("testdata", "update_multiple.golden.proto.json")).WithEncoder(ProtoJSONEncoder{})
	err := g.UpdateValues(got)
	require.NoError(t, err)
}
