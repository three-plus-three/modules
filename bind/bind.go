package bind

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/revel/revel"
)

func init() {
	revel.KindBinders[reflect.Slice] = revel.Binder{Bind: bindSlice, Unbind: unbindSlice}
	revel.TypeBinders[reflect.TypeOf(net.IP{})] = revel.Binder{bindIP, unbindIP}

	revel.TypeBinders[reflect.TypeOf((*byte)(nil)).Elem()] = revel.Binder{bindByte, unbindByte}
}

func bindByte(params *revel.Params, name string, typ reflect.Type) reflect.Value {
	value := params.Get(name)
	if "" == value {
		return reflect.Zero(typ)
	}
	if len(value) != 1 {
		panic(errors.New("'" + value + "' is invalid byte."))
	}
	b := value[0]
	return reflect.ValueOf(b)
}

func unbindByte(output map[string]string, name string, val interface{}) {
	output[name] = string(val.(byte))
}

func bindIP(params *revel.Params, name string, typ reflect.Type) reflect.Value {
	value := params.Get(name)
	if "" == value {
		return reflect.Zero(typ)
	}
	ip := net.ParseIP(value)
	if nil == ip {
		panic(errors.New("'" + value + "' is invalid address."))
	}
	return reflect.ValueOf(ip)
}

func unbindIP(output map[string]string, name string, val interface{}) {
	switch v := val.(type) {
	case string:
		output[name] = v
	case *string:
		output[name] = *v
	case net.IP:
		output[name] = v.String()
	case *net.IP:
		output[name] = v.String()
	}
}

// This function creates a slice of the given type, Binds each of the individual
// elements, and then sets them to their appropriate location in the slice.
// If elements are provided without an explicit index, they are added (in
// unspecified order) to the end of the slice.
func bindSlice(params *revel.Params, name string, typ reflect.Type) reflect.Value {
	// Collect an array of slice elements with their indexes (and the max index).
	//maxIndex := -1
	numNoIndex := 0
	sliceValues := []sliceValue{}
	uniqueIndexs := []int{}

	// Factor out the common slice logic (between form values and files).
	processElement := func(key string, vals []string, files []*multipart.FileHeader) {
		if !strings.HasPrefix(key, name+"[") {
			return
		}

		// Extract the index, and the index where a sub-key starts. (e.g. field[0].subkey)
		index := -1
		leftBracket, rightBracket := len(name), strings.Index(key[len(name):], "]")+len(name)
		if rightBracket > leftBracket+1 {
			index, _ = strconv.Atoi(key[leftBracket+1 : rightBracket])
		}
		subKeyIndex := rightBracket + 1

		// Handle the indexed case.
		if index > -1 {
			smallestIndex := -1
			for idx, n := range uniqueIndexs {
				if n == index {
					smallestIndex = idx
				}
			}

			if smallestIndex == -1 {
				uniqueIndexs = append(uniqueIndexs, index)
				smallestIndex = len(uniqueIndexs) - 1
			}

			//if index > maxIndex {
			//  maxIndex = index
			//}
			sliceValues = append(sliceValues, sliceValue{
				index: smallestIndex,
				value: revel.Bind(params, key[:subKeyIndex], typ.Elem()),
			})
			return
		}

		// It's an un-indexed element.  (e.g. element[])
		numNoIndex += len(vals) + len(files)
		for _, val := range vals {
			// Unindexed values can only be direct-bound.
			sliceValues = append(sliceValues, sliceValue{
				index: -1,
				value: revel.BindValue(val, typ.Elem()),
			})
		}

		for _, fileHeader := range files {
			sliceValues = append(sliceValues, sliceValue{
				index: -1,
				value: revel.BindFile(fileHeader, typ.Elem()),
			})
		}
	}

	for key, vals := range params.Values {
		processElement(key, vals, nil)
	}
	for key, fileHeaders := range params.Files {
		processElement(key, nil, fileHeaders)
	}

	resultArray := reflect.MakeSlice(typ, len(uniqueIndexs), len(uniqueIndexs)+numNoIndex)
	//resultArray := reflect.MakeSlice(typ, maxIndex+1, maxIndex+1+numNoIndex)
	for _, sv := range sliceValues {
		if sv.index != -1 {
			resultArray.Index(sv.index).Set(sv.value)
		} else {
			resultArray = reflect.Append(resultArray, sv.value)
		}
	}

	return resultArray
}

func unbindSlice(output map[string]string, name string, val interface{}) {
	v := reflect.ValueOf(val)
	for i := 0; i < v.Len(); i++ {
		revel.Unbind(output, fmt.Sprintf("%s[%d]", name, i), v.Index(i).Interface())
	}
}

// Used to keep track of the index for individual keyvalues.
type sliceValue struct {
	index int           // Index extracted from brackets.  If -1, no index was provided.
	value reflect.Value // the bound value for this slice element.
}
