package goldentest

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

type Golden[T any] struct {
	path          string
	packed        bool
	encoder       Encoder
	writeMode     os.FileMode
	ignoreFields  []string
	beforeCompare func(want, got *T) error
	beforeUpdate  func(want, got *T) error
}

type CompareResult[T any] struct {
	Want *T
	Got  *T
	Diff string
}

func (r CompareResult[T]) OK() bool {
	return r.Diff == ""
}

func New[T any](path string) Golden[T] {
	return Golden[T]{
		path:    path,
		encoder: JSONEncoder{},
	}
}

func (g Golden[T]) Path() string {
	return g.path
}

func (g Golden[T]) WithIgnoreFields(fields ...string) Golden[T] {
	g.ignoreFields = fields
	return g
}
func (g Golden[T]) WithPackedOutput(packed bool) Golden[T] {
	g.packed = packed
	return g
}
func (g Golden[T]) WriteOutputFileMode(mode os.FileMode) Golden[T] {
	g.writeMode = mode
	return g
}
func (g Golden[T]) WithBeforeUpdate(f func(want, got *T) error) Golden[T] {
	g.beforeUpdate = f
	return g
}
func (g Golden[T]) WithBeforeCompare(f func(want, got *T) error) Golden[T] {
	g.beforeCompare = f
	return g
}
func (g Golden[T]) WithEncoder(encoder Encoder) Golden[T] {
	g.encoder = encoder
	return g
}

func (g Golden[T]) Update(got T) error {
	return g.update(got)
}

func (g Golden[T]) UpdateValues(got []T) error {
	return g.update(got)
}

func (g Golden[T]) update(golden any) error {
	data, err := g.encoder.Marshal(golden, g.packed)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return g.Write(data)
}

func (g Golden[T]) Write(data []byte) error {
	if g.writeMode == 0 {
		g.writeMode = os.ModePerm
	}
	return os.WriteFile(g.path, data, g.writeMode)
}

func (g Golden[T]) cmp(want, got *T) (CompareResult[T], error) {
	if g.beforeCompare != nil {
		err := g.beforeCompare(want, got)
		if err != nil {
			return CompareResult[T]{Want: want, Got: got}, err
		}
	}

	var opts = []cmp.Option{protocmp.Transform()}
	if len(g.ignoreFields) > 0 {
		opts = append(opts, cmpopts.IgnoreFields(*want, g.ignoreFields...))
	}

	return CompareResult[T]{
		Want: want,
		Got:  got,
		Diff: cmp.Diff(want, got, opts...),
	}, nil
}

// TODO: maybe JSON can call this one instead of compare values
func (g Golden[T]) Compare(got *T) (CompareResult[T], error) {
	goldenData, err := os.ReadFile(g.path)
	if err != nil {
		return CompareResult[T]{}, fmt.Errorf("read file: %w", err)
	}

	var want T
	err = g.encoder.Unmarshal(goldenData, &want)
	if err != nil {
		return CompareResult[T]{}, fmt.Errorf("unmarshal golden: %w", err)
	}

	return g.cmp(&want, got)
}

func (g Golden[T]) CompareValues(gotValues []T) (map[int]CompareResult[T], error) {
	result := make(map[int]CompareResult[T])

	goldenData, err := os.ReadFile(g.path)
	if err != nil {
		return result, fmt.Errorf("read file: %w", err)
	}

	var wantValues []T
	err = g.encoder.Unmarshal(goldenData, &wantValues)
	if err != nil {
		return result, fmt.Errorf("unmarshal golden values: %w", err)
	}

	if len(gotValues) != len(wantValues) {
		return result, fmt.Errorf("want %d items but got %d", len(wantValues), len(gotValues))
	}

	if len(wantValues) == 0 {
		return result, nil
	}

	for i := range gotValues {
		want := wantValues[i]
		got := gotValues[i]

		res, err := g.cmp(&want, &got)
		if err != nil {
			return result, err
		}
		if res.OK() {
			continue
		}
		result[i] = res
	}

	return result, nil
}

func (g Golden[T]) CompareElements(gotValues []T) (map[int]CompareResult[T], error) {
	result := make(map[int]CompareResult[T])

	goldenData, err := os.ReadFile(g.path)
	if err != nil {
		return result, fmt.Errorf("read file: %w", err)
	}

	var wantValues []T
	err = g.encoder.Unmarshal(goldenData, &wantValues)
	if err != nil {
		return result, fmt.Errorf("unmarshal golden values: %w", err)
	}

	if len(gotValues) != len(wantValues) {
		return result, fmt.Errorf("want %d items but got %d", len(wantValues), len(gotValues))
	}

	if len(wantValues) == 0 {
		return result, nil
	}

	for i := range gotValues {
		want := wantValues[i]
		got := gotValues[i]

		res, err := g.cmp(&want, &got)
		if err != nil {
			return result, err
		}
		if res.OK() {
			continue
		}
		result[i] = res
	}

	return result, nil
}

type Encoder interface {
	Marshal(v any, packed bool) ([]byte, error)
	Unmarshal(data []byte, v any) error
}

type JSONEncoder struct{}

func (j JSONEncoder) Marshal(v any, packed bool) ([]byte, error) {
	if packed {
		return json.MarshalIndent(v, "", "\t")
	}
	return json.Marshal(v)
}

func (j JSONEncoder) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type ProtoJSONEncoder struct{}

func (j ProtoJSONEncoder) marshal(m proto.Message, packed bool) ([]byte, error) {
	indent := ""
	if !packed {
		indent = "\t"
	}
	return protojson.MarshalOptions{Indent: indent}.Marshal(m)
}

func (j ProtoJSONEncoder) toObj(m proto.Message, packed bool) (map[string]any, error) {
	data, err := j.marshal(m, packed)
	if err != nil {
		return nil, err
	}
	obj := make(map[string]any)
	if err = json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (j ProtoJSONEncoder) Marshal(v any, packed bool) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if ok {
		return j.marshal(msg, packed)
	}

	kind := reflect.TypeOf(v).Kind()
	if kind != reflect.Slice {
		return nil, fmt.Errorf("unsupport type: %s", kind)
	}

	values := reflect.ValueOf(v)
	results := make([]map[string]any, 0, values.Len())
	for i := 0; i < values.Len(); i++ {
		vv := values.Index(i).Interface()
		msg, ok := vv.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("invalid proto message")
		}
		obj, err := j.toObj(msg, packed)
		if err != nil {
			return nil, err
		}
		results = append(results, obj)
	}
	return JSONEncoder{}.Marshal(results, packed)
}

func (j ProtoJSONEncoder) unmarshal(data []byte, v any) error {
	msg, ok := v.(proto.Message)
	if ok {
		return protojson.Unmarshal(data, msg)
	}
	msgm, ok := v.(*proto.Message)
	if ok {
		return protojson.Unmarshal(data, *msgm)
	}
	return fmt.Errorf("non proto message")
}

func (j ProtoJSONEncoder) Unmarshal(data []byte, v any) error {
	if v == nil {
		return fmt.Errorf("missing value")
	}
	return j.unmarshal(data, v)
}
