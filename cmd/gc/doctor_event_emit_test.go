package main

import (
	"errors"
	"testing"

	"github.com/gastownhall/gascity/internal/doctor"
)

func TestEventEmitBinaryDoctorCheckWarnsWhenUnavailable(t *testing.T) {
	check := newEventEmitBinaryDoctorCheck(func(name string) (string, error) {
		if name != "gc-event-emit" {
			t.Fatalf("lookPath name = %q, want gc-event-emit", name)
		}
		return "", errors.New("not found")
	})

	result := check.Run(&doctor.CheckContext{})
	if result.Status != doctor.StatusWarning {
		t.Fatalf("status = %v, want warning; result=%+v", result.Status, result)
	}
}

func TestEventEmitBinaryDoctorCheckOKWhenAvailable(t *testing.T) {
	check := newEventEmitBinaryDoctorCheck(func(name string) (string, error) {
		if name != "gc-event-emit" {
			t.Fatalf("lookPath name = %q, want gc-event-emit", name)
		}
		return "/usr/local/bin/gc-event-emit", nil
	})

	result := check.Run(&doctor.CheckContext{})
	if result.Status != doctor.StatusOK {
		t.Fatalf("status = %v, want OK; result=%+v", result.Status, result)
	}
}
