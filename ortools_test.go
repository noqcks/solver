/* ----------------------------------------------------------------------------
 * This file was automatically generated by SWIG (http://www.swig.org).
 * Version 3.0.12
 *
 * This file is not intended to be easily readable and contains a number of
 * coding conventions designed to improve portability and efficiency. Do not make
 * changes to this file unless you know what you are doing--modify the SWIG
 * interface file instead.
 * ----------------------------------------------------------------------------- */

// source: linear_solver.i
// Package ortools does stuff
package ortools

import (
	//"C"
	//_ "runtime/cgo"

	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewSolver(t *testing.T) {
	tests := []struct {
		name      string
		got, want interface{}
	}{
		{"test 1", "a", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.got); diff != "" {
				t.Errorf("unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}