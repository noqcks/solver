// Copyright 2021 Irfan Sharif.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package parser_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cockroachdb/datadriven"
	"github.com/noqcks/solver/internal/testutils/bazel"
	"github.com/noqcks/solver/internal/testutils/parser"
	"github.com/noqcks/solver/internal/testutils/parser/ast"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/ebnf"
)

func TestDatadriven(t *testing.T) {
	datadriven.Walk(t, "testdata", func(t *testing.T, path string) {
		path, implant := bazel.WritableSandboxPathFor(t, "internal/testutils/parser", path)
		defer implant()

		datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
			p := parser.New(t, d.Input)
			var out string
			switch d.Cmd {
			case "receiver":
				out = p.Receiver()
			case "identifier":
				out = p.Identifier()
			case "method":
				var method ast.Method
				method = p.Method()
				out = method.String()
			case "variable":
				out = p.Variable()
			case "variables":
				variables := p.Variables()
				out = strings.Join(variables, ", ")
			case "enforcement":
				e := p.Enforcement()
				out = e.String()
			case "interval":
				i := p.Interval()
				out = i.String()
			case "boolean":
				boolean := p.Boolean()
				out = fmt.Sprintf("%t", boolean)
			case "booleans":
				booleans := p.Booleans()
				var strs []string
				for _, boolean := range booleans {
					strs = append(strs, fmt.Sprintf("%t", boolean))
				}
				out = strings.Join(strs, ", ")
			case "number":
				n := p.Number()
				out = fmt.Sprintf("%d", n)
			case "numbers":
				numbers := p.Numbers()
				var strs []string
				for _, number := range numbers {
					strs = append(strs, fmt.Sprintf("%d", number))
				}
				out = strings.Join(strs, ", ")
			case "intervals":
				intervals := p.Intervals()
				var strs []string
				for _, interval := range intervals {
					strs = append(strs, interval.String())
				}
				out = strings.Join(strs, ", ")
			case "interval-demand":
				demand := p.IntervalDemand()
				out = demand.String()
			case "domain":
				domain := p.Domain()
				out = domain.String()
			case "linear-term":
				term := p.LinearTerm()
				out = term.String()
			case "linear-expr":
				expr := p.LinearExpr()
				out = expr.String()
			case "linear-exprs":
				exprs := p.LinearExprs()
				var strs []string
				for _, expr := range exprs {
					strs = append(strs, expr.String())
				}
				out = strings.Join(strs, ", ")
			case "domains":
				domains := p.Domains()
				var strs []string
				for _, domain := range domains {
					strs = append(strs, domain.String())
				}
				out = strings.Join(strs, " ∪ ")
			case "statement":
				stmt := p.Statement()
				out = stmt.String()
			case "numbers-list":
				list := p.NumbersList()
				var strs []string
				for _, l := range list {
					var inner []string
					for _, n := range l {
						inner = append(inner, fmt.Sprintf("%d", n))
					}
					strs = append(strs, fmt.Sprintf("[%s]", strings.Join(inner, ", ")))
				}
				out = strings.Join(strs, " ∪ ")
			case "booleans-list":
				list := p.BooleansList()
				var strs []string
				for _, l := range list {
					var inner []string
					for _, b := range l {
						inner = append(inner, fmt.Sprintf("%t", b))
					}
					strs = append(strs, fmt.Sprintf("[%s]", strings.Join(inner, ", ")))
				}
				out = strings.Join(strs, " ∪ ")
			case "assignments-argument":
				arg := p.AssignmentsArgument()
				out = arg.String()
			case "binary-op-argument":
				arg := p.BinaryOpArgument()
				out = arg.String()
			case "constants-argument":
				arg := p.ConstantsArgument()
				out = arg.String()
			case "cumulative-argument":
				arg := p.CumulativeArgument()
				out = arg.String()
			case "k-argument":
				arg := p.KArgument()
				out = arg.String()
			case "domain-argument":
				arg := p.DomainArgument()
				out = arg.String()
			case "element-argument":
				arg := p.ElementArgument()
				out = arg.String()
			default:
				t.Errorf("unrecognized command: %s", d.Cmd)
			}

			if !p.EOF() {
				return fmt.Sprintf("err: expected EOF; parsed %q", out)
			}
			return out
		})
	})
}

func TestGrammar(t *testing.T) {
	filename := `grammar.ebnf`
	contents, err := ioutil.ReadFile(filename)
	require.Nil(t, err)

	grammar, err := ebnf.Parse(filename, bytes.NewReader(contents))
	if err != nil {
		t.Fatal(err)
	}
	if err := ebnf.Verify(grammar, `Statement`); err != nil { // verify the top-level statement
		t.Fatal(err)
	}
}
