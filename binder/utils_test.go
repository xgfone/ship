package binder

import (
	"testing"
	"time"
)

func TestSetValue(t *testing.T) {
	var b bool
	var bs []byte
	var s string
	var f32 float32
	var f64 float64
	var i int
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64
	var u uint
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var u64 uint64
	var tt1 time.Time
	var tt2 time.Time

	if err := SetValue(&b, "on"); err != nil || !b {
		t.Fail()
	}
	if err := SetValue(&bs, "bytes"); err != nil || string(bs) != "bytes" {
		t.Fail()
	}
	if err := SetValue(&s, "string"); err != nil || s != "string" {
		t.Fail()
	}
	if err := SetValue(&f32, "1.0"); err != nil || f32 != 1.0 {
		t.Fail()
	}
	if err := SetValue(&f64, "1.0"); err != nil || f64 != 1.0 {
		t.Fail()
	}
	if err := SetValue(&i, "123"); err != nil || i != 123 {
		t.Fail()
	}
	if err := SetValue(&i8, "123"); err != nil || i8 != 123 {
		t.Fail()
	}
	if err := SetValue(&i16, "123"); err != nil || i16 != 123 {
		t.Fail()
	}
	if err := SetValue(&i32, "123"); err != nil || i32 != 123 {
		t.Fail()
	}
	if err := SetValue(&i64, "123"); err != nil || i64 != 123 {
		t.Fail()
	}
	if err := SetValue(&u, "123"); err != nil || u != 123 {
		t.Fail()
	}
	if err := SetValue(&u8, "123"); err != nil || u8 != 123 {
		t.Fail()
	}
	if err := SetValue(&u16, "123"); err != nil || u16 != 123 {
		t.Fail()
	}
	if err := SetValue(&u32, "123"); err != nil || u32 != 123 {
		t.Fail()
	}
	if err := SetValue(&u64, "123"); err != nil || u64 != 123 {
		t.Fail()
	}
	if err := SetValue(&tt1, "2019-01-16T15:39:40Z"); err != nil || tt1.String() != "2019-01-16 15:39:40 +0000 UTC" {
		t.Error(tt1)
	}
	if err := SetValue(&tt2, "2019-01-16T15:39:40+08:00"); err != nil ||
		(tt2.String() != "2019-01-16 15:39:40 +0800 CST" && tt2.String() != "2019-01-16 15:39:40 +0800 +0800") {
		t.Error(tt2)
	}

	tt2 = tt2.UTC()

	if tt1.Year() != tt2.Year() {
		t.Error(tt1.Year(), tt2.Year())
	}
	if tt1.Month() != tt2.Month() {
		t.Error(tt1.Month(), tt2.Month())
	}
	if tt1.Day() != tt2.Day() {
		t.Error(tt1.Day(), tt2.Day())
	}
	if tt1.Hour()-tt2.Hour() != 8 {
		t.Error(tt1.Hour(), tt2.Hour())
	}
	if tt1.Minute() != tt2.Minute() {
		t.Error(tt1.Minute(), tt2.Minute())
	}
	if tt1.Second() != tt2.Second() {
		t.Error(tt1.Second(), tt2.Second())
	}
}

func TestSetStructValue(t *testing.T) {
	type S struct {
		Name string
		Age  int
	}
	s := S{}
	if err := SetStructValue(&s, "Name", "abc"); err != nil {
		t.Error(err)
	}
	if err := SetStructValue(&s, "Age", "123"); err != nil {
		t.Error(err)
	}
	if s.Name != "abc" || s.Age != 123 {
		t.Error(s.Name, s.Age)
	}
}
