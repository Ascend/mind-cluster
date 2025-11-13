// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package common a series of util test function
package common

import (
	"reflect"
	"testing"
)

// TestDeepCopy_BasicTypes tests deep copy of basic types
func TestDeepCopy_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		src      interface{}
		expected interface{}
	}{
		{
			name:     "int type",
			src:      42,
			expected: 42,
		},
		{
			name:     "string type",
			src:      "hello world",
			expected: "hello world",
		},
		{
			name:     "bool type",
			src:      true,
			expected: true,
		},
		{
			name:     "float64 type",
			src:      3.14159,
			expected: 3.14159,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := reflect.New(reflect.TypeOf(tt.src)).Interface()
			if err := DeepCopy(dst, tt.src); err != nil {
				t.Errorf("DeepCopy() error = %v", err)
				return
			}
			actual := reflect.ValueOf(dst).Elem().Interface()
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("DeepCopy() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

// TestDeepCopy_Struct tests func DeepCopy of struct types
func TestDeepCopy_Struct(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}
	type Person struct {
		Name    string
		Address Address
		Hobbies []string
	}
	src := Person{
		Name: "Alice",
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
		},
		Hobbies: []string{"reading", "swimming"},
	}
	dst := Person{}
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("DeepCopy() = %v, want %v", dst, src)
	}
	// Verification is a deep copy (modifying the original object should not affect the copied object)
	src.Name = "Bob"
	src.Hobbies[0] = "cooking"
	if dst.Name == src.Name {
		t.Error("DeepCopy() did not create a deep copy - Name field was affected")
	}
	if dst.Hobbies[0] == src.Hobbies[0] {
		t.Error("DeepCopy() did not create a deep copy - Hobbies slice was affected")
	}

	// test empty struct
	src = Person{}
	dst = Person{}
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("DeepCopy() = %v, want %v", dst, src)
	}
}

// TestDeepCopy_NilSource tests behavior when source is nil
func TestDeepCopy_NilSource(t *testing.T) {
	var dst interface{}
	var src interface{} = nil
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() with nil source should not return error, got = %v", err)
	}
}

// TestDeepCopy_TypeMismatch tests behavior when destination type doesn't match source type
func TestDeepCopy_TypeMismatch(t *testing.T) {
	// src type: string, dst type: int
	src := "this is a string"
	var dst int
	if err := DeepCopy(&dst, src); err == nil {
		t.Error("DeepCopy() should return error for type mismatch")
	}
}

// TestDeepCopy_PointerToPointer tests deep copy of pointer to pointer
func TestDeepCopy_PointerToPointer(t *testing.T) {
	value := 100
	src := &value
	var dst *int
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}
	if *dst != *src {
		t.Errorf("DeepCopy() = %v, want %v", *dst, *src)
	}
	// Verification refers to pointing to different memory addresses
	if dst == src {
		t.Error("DeepCopy() did not create new pointer")
	}
}

// TestDeepCopy_SliceAndMap tests deep copy of slices and maps
func TestDeepCopy_SliceAndMap(t *testing.T) {
	src := struct {
		Slice []int
		Map   map[string]int
	}{
		Slice: []int{1, 2, 3, 4, 5},
		Map:   map[string]int{"one": 1, "two": 2, "three": 3},
	}
	var dst struct {
		Slice []int
		Map   map[string]int
	}
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}

	// Verify if the content is copied correctly
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("DeepCopy() = %v, want %v", dst, src)
	}
	// Verification is a deep copy
	src.Slice[0] = 999
	src.Map["one"] = 999
	if dst.Slice[0] == 999 || dst.Map["one"] == 999 {
		t.Error("DeepCopy() did not create a deep copy - Slice or Map was affected")
	}
}
