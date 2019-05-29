package main

import (
	"errors"
	"testing"
	"time"

	"go4.org/testing/functest"
)

func TestParseWait(t *testing.T) {
	f := functest.New(parseWait)

	f.Test(t,
		f.In("").Want(time.Duration(0), time.Duration(0), nil),
		f.In("500ms").Want(500*time.Millisecond, time.Duration(0), nil),
		f.In("500ms:").Want(500*time.Millisecond, time.Duration(0), nil),
		f.In("500ms:5s").Want(500*time.Millisecond, 5*time.Second, nil),
		f.In(":5s").Want(time.Duration(0), time.Duration(0), errors.New("minimum is required")),
		f.In(":").Want(time.Duration(0), time.Duration(0), errors.New("minimum is required")),
		f.In("1s:500ms").Want(1*time.Second, 500*time.Millisecond, errors.New("max must be greater than or equal to min")),
		f.In("abc").Want(time.Duration(0), time.Duration(0), errors.New("time: invalid duration abc")),
		f.In("100ms:def").Want(100*time.Millisecond, time.Duration(0), errors.New("time: invalid duration def")),
	)
}
