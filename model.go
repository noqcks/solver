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

package solver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/noqcks/solver/internal"
	"github.com/noqcks/solver/internal/pb"
)

// Model is a constraint programming problem. It's not safe for concurrent use.
type Model struct {
	pb *pb.CpModelProto

	// We hold onto these only for String()
	vars, constants []IntVar
	literals        []Literal
	intervals       []Interval
	constraints     []Constraint
	objective       LinearExpr
	minimize        bool
}

// TODO(irfansharif): Add assumption literals and examples for unsat debugging.
// And add hints. Add some documentation from
// https://github.com/google/or-tools/blob/stable/ortools/sat/doc/boolean_logic.md
// (reification, channeling constraints). Export async handler to stop search
// process. Probably part of enumerator?
// TODO(irfansharif): Export solution hint/partial assignment, when previous one
// is available.
// TODO(irfansharif): Export model/result statistics. Export domain.complement
// for reification.
// TODO(irfansharif): Export verbose view of types (specifically -- include
// internal indexes, so you could debug the validation error).

// NewModel instantiates a new model.
func NewModel(name string) *Model {
	return &Model{
		pb: &pb.CpModelProto{
			Name: name,
		},
	}
}

// NewLiteral adds a new literal to the model.
func (m *Model) NewLiteral(name string) Literal {
	literal := m.newIntVarFromDomainInternal(NewDomain(0, 1), true, false, name).(Literal)
	m.literals = append(m.literals, literal)
	return literal
}

// NewConstant adds a new constant to the model.
func (m *Model) NewConstant(c int64, name string) IntVar {
	constant := m.newIntVarFromDomainInternal(NewDomain(c, c), false, true, name)
	m.constants = append(m.constants, constant)
	return constant
}

// NewIntVar adds a new integer variable to the model, one that's constrained to
// the given inclusive upper/lower bound.
func (m *Model) NewIntVar(lb int64, ub int64, name string) IntVar {
	return m.NewIntVarFromDomain(NewDomain(lb, ub), name)
}

// NewIntVarFromDomain adds a new integer variable to the model, one that's
// constrained to the given domain.
func (m *Model) NewIntVarFromDomain(d Domain, name string) IntVar {
	iv := m.newIntVarFromDomainInternal(d, false, false, name)
	m.vars = append(m.vars, iv)
	return iv
}

// NewInterval adds a new interval to the model, one that's defined using the
// given start, end and size.
func (m *Model) NewInterval(start, end, size IntVar, name string) Interval {
	idx := len(m.pb.GetConstraints())
	itv := newInterval(start, end, size, int32(idx), name)
	m.addConstraintsInternal(itv)
	m.intervals = append(m.intervals, itv)
	return itv
}

// AddConstraints adds constraints to the model. When deciding on a solution,
// these constraints will need to be satisfied.
func (m *Model) AddConstraints(cs ...Constraint) {
	m.addConstraintsInternal(cs...)
	m.constraints = append(m.constraints, cs...)
}

// Minimize sets a minimization objective for the model.
func (m *Model) Minimize(e LinearExpr) {
	m.pb.Objective = m.toObjectiveProto(e)
	m.objective, m.minimize = e, true
}

// Maximize sets a maximization objective for the model.
func (m *Model) Maximize(e LinearExpr) {
	// For maximization objectives, we want to negate all the coefficients and
	// set the scaling factor to -1.
	proto := m.toObjectiveProto(e)
	for i, coeff := range proto.Coeffs {
		proto.Coeffs[i] = -coeff
	}
	proto.Offset = -proto.Offset
	proto.ScalingFactor = -1
	m.pb.Objective = proto

	m.objective, m.minimize = e, false
}

// Validate checks whether the model is valid. If not, a descriptive error
// message is returned.
//
// TODO(irfansharif): This validation message refers to things using indexes,
// which is not really usable.
func (m *Model) Validate() (ok bool, _ error) {
	validation := internal.CpSatHelperValidateModel(*m.pb)
	if validation == "" {
		return true, nil
	}

	return false, errors.New(validation)
}

// String provides a string representation of the model.
func (m *Model) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("model=%s\n", m.name()))

	for i, v := range m.vars {
		if i == 0 {
			b.WriteString(fmt.Sprintf("  variables (num = %d)\n", len(m.vars)))
		}
		b.WriteString(fmt.Sprintf("    %s\n", v.String()))
	}

	for i, c := range m.constants {
		if i == 0 {
			b.WriteString(fmt.Sprintf("  constants (num = %d)\n", len(m.constants)))
		}
		b.WriteString(fmt.Sprintf("    %s\n", c.String()))
	}

	for i, l := range m.literals {
		if i == 0 {
			b.WriteString(fmt.Sprintf("  literals (num = %d)\n", len(m.literals)))
		}
		b.WriteString(fmt.Sprintf("    %s\n", l.String()))
	}

	for i, iv := range m.intervals {
		if i == 0 {
			b.WriteString(fmt.Sprintf("  intervals (num = %d)\n", len(m.intervals)))
		}
		b.WriteString(fmt.Sprintf("    %s\n", iv.String()))
	}

	for i, c := range m.constraints {
		if i == 0 {
			b.WriteString(fmt.Sprintf("  constraints (num = %d)\n", len(m.constraints)))
		}
		b.WriteString(fmt.Sprintf("    %s\n", c.String()))
	}

	if o := m.objective; o != nil {
		direction := "minimize"
		if !m.minimize {
			direction = "maximize"
		}
		b.WriteString(fmt.Sprintf("   objective: %s: %s\n", direction, o.String()))
	}

	return b.String()
}

// Solve attempts to satisfy the model's constraints, if any, by deciding values
// for all the variables/literals that were instantiated into it. It returns the
// optimal result if an objective function is declared. If not, it returns
// the first found result that satisfies the model.
//
// The solve process itself can be configured with various options.
func (m *Model) Solve(os ...Option) Result {
	solver := internal.NewSolveWrapper()
	defer func() { internal.DeleteSolveWrapper(solver) }()

	var opts options
	for _, o := range os {
		o(&opts, solver)
	}
	if ok, err := opts.validate(); !ok {
		panic(err)
	}
	if opts.solution != nil {
		defer func() { internal.DeleteDirectorSolutionCallback(opts.solution.hook) }()
	}

	solver.SetParameters(opts.params)
	resp := solver.Solve(*m.pb)

	if opts.logger != nil {
		for _, line := range strings.Split(resp.SolveLog, "\n") {
			opts.logger.Print(line)
		}
	}
	return Result{pb: &resp}
}

func (m *Model) name() string {
	name := m.pb.GetName()
	if name == "" {
		name = "<unnamed>"
	}
	return name
}

func (m *Model) newIntVarFromDomainInternal(d Domain, isLiteral, isConst bool, name string) IntVar {
	idx := len(m.pb.GetVariables())
	iv := newIntVar(d, int32(idx), isLiteral, isConst, name)
	m.pb.Variables = append(m.pb.Variables, iv.pb)
	return iv
}

func (m *Model) addConstraintsInternal(cs ...Constraint) {
	for _, c := range cs {
		m.pb.Constraints = append(m.pb.Constraints, c.protos()...)
	}
}

func (m *Model) toObjectiveProto(e LinearExpr) *pb.CpObjectiveProto {
	return &pb.CpObjectiveProto{
		Vars:   e.vars(),
		Coeffs: e.coeffs(),
		Offset: float64(e.offset()),
	}
}
