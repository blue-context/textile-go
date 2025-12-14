// Copyright 2025 Blue Context Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package clone provides deep cloning utilities using reflection.
package clone

import (
	"fmt"
	"reflect"
	"time"
)

// Deep performs a deep copy of the input using reflection.
// It handles all common Go types including primitives, pointers, slices, maps, and structs.
//
// Special handling is provided for:
// - time.Time (copied by value)
// - Unexported fields (skipped)
// - Circular references (not detected - caller must ensure no cycles)
func Deep[T any](src T) (T, error) {
	var zero T

	srcVal := reflect.ValueOf(src)
	if !srcVal.IsValid() {
		return zero, nil
	}

	dstVal := reflect.New(srcVal.Type()).Elem()
	if err := deepCopyValue(srcVal, dstVal); err != nil {
		return zero, fmt.Errorf("failed to deep copy: %w", err)
	}

	result, ok := dstVal.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("type assertion failed after deep copy")
	}

	return result, nil
}

// deepCopyValue recursively copies src to dst using reflection.
func deepCopyValue(src, dst reflect.Value) error {
	if !src.IsValid() {
		return nil
	}

	// Special case: time.Time should be copied by value
	if src.Type() == reflect.TypeOf(time.Time{}) {
		if dst.CanSet() {
			dst.Set(src)
		}
		return nil
	}

	switch src.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		// Primitive types: direct copy
		if dst.CanSet() {
			dst.Set(src)
		}

	case reflect.Ptr:
		// Pointer: allocate new, copy pointed value
		if src.IsNil() {
			// Keep dst as nil
			return nil
		}

		if !dst.CanSet() {
			return nil
		}

		// Allocate new pointer
		newPtr := reflect.New(src.Type().Elem())
		if err := deepCopyValue(src.Elem(), newPtr.Elem()); err != nil {
			return err
		}
		dst.Set(newPtr)

	case reflect.Slice:
		// Slice: allocate new backing array, copy elements
		if src.IsNil() {
			return nil
		}

		if !dst.CanSet() {
			return nil
		}

		newSlice := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		for i := 0; i < src.Len(); i++ {
			if err := deepCopyValue(src.Index(i), newSlice.Index(i)); err != nil {
				return fmt.Errorf("failed to copy slice element %d: %w", i, err)
			}
		}
		dst.Set(newSlice)

	case reflect.Array:
		// Array: copy each element
		if !dst.CanSet() {
			return nil
		}

		for i := 0; i < src.Len(); i++ {
			if err := deepCopyValue(src.Index(i), dst.Index(i)); err != nil {
				return fmt.Errorf("failed to copy array element %d: %w", i, err)
			}
		}

	case reflect.Map:
		// Map: allocate new map, copy all keys/values
		if src.IsNil() {
			return nil
		}

		if !dst.CanSet() {
			return nil
		}

		newMap := reflect.MakeMap(src.Type())
		iter := src.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()

			// Deep copy both key and value
			newKey := reflect.New(key.Type()).Elem()
			if err := deepCopyValue(key, newKey); err != nil {
				return fmt.Errorf("failed to copy map key: %w", err)
			}

			newVal := reflect.New(val.Type()).Elem()
			if err := deepCopyValue(val, newVal); err != nil {
				return fmt.Errorf("failed to copy map value: %w", err)
			}

			newMap.SetMapIndex(newKey, newVal)
		}
		dst.Set(newMap)

	case reflect.Struct:
		// Struct: copy all exported fields recursively
		for i := 0; i < src.NumField(); i++ {
			srcField := src.Field(i)
			dstField := dst.Field(i)

			// Skip unexported fields
			if !srcField.CanInterface() {
				continue
			}

			if err := deepCopyValue(srcField, dstField); err != nil {
				structType := src.Type()
				fieldName := structType.Field(i).Name
				return fmt.Errorf("failed to copy struct field %s: %w", fieldName, err)
			}
		}

	case reflect.Interface:
		// Interface: copy underlying concrete type
		if src.IsNil() {
			return nil
		}

		if !dst.CanSet() {
			return nil
		}

		// Get the concrete value
		concreteVal := src.Elem()
		newVal := reflect.New(concreteVal.Type()).Elem()

		if err := deepCopyValue(concreteVal, newVal); err != nil {
			return fmt.Errorf("failed to copy interface value: %w", err)
		}

		dst.Set(newVal)

	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// These types cannot be meaningfully deep copied
		// Channels and functions are references by nature
		// Just copy the reference
		if dst.CanSet() {
			dst.Set(src)
		}

	default:
		return fmt.Errorf("unsupported type for deep copy: %v", src.Kind())
	}

	return nil
}
