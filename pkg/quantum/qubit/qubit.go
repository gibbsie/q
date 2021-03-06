package qubit

import (
	"fmt"
	"math"
	"math/cmplx"
	"strconv"
	"strings"

	"github.com/itsubaki/q/pkg/math/matrix"
	"github.com/itsubaki/q/pkg/math/number"
	"github.com/itsubaki/q/pkg/math/rand"
	"github.com/itsubaki/q/pkg/math/vector"
)

type Qubit struct {
	vector vector.Vector
	Seed   []int64
	Rand   func(seed ...int64) float64
}

func New(z ...complex128) *Qubit {
	q := &Qubit{
		vector: vector.New(z...),
		Rand:   rand.Crypto,
	}

	q.Normalize()
	return q
}

func Zero(n ...int) *Qubit {
	v := vector.TensorProductN(vector.Vector{1, 0}, n...)
	return New(v.Complex()...)
}

func One(n ...int) *Qubit {
	v := vector.TensorProductN(vector.Vector{0, 1}, n...)
	return New(v.Complex()...)
}

func (q *Qubit) NumberOfBit() int {
	d := float64(q.Dimension())
	n := math.Log2(d)
	return int(n)
}

func (q *Qubit) IsZero(eps ...float64) bool {
	return q.Equals(Zero(), eps...)
}

func (q *Qubit) IsOne(eps ...float64) bool {
	return q.Equals(One(), eps...)
}

func (q *Qubit) InnerProduct(qb *Qubit) complex128 {
	return q.vector.InnerProduct(qb.vector)
}

func (q *Qubit) OuterProduct(qb *Qubit) matrix.Matrix {
	return q.vector.OuterProduct(qb.vector)
}

func (q *Qubit) Dimension() int {
	return q.vector.Dimension()
}

func (q *Qubit) Clone() *Qubit {
	return &Qubit{
		vector: q.vector.Clone(),
		Seed:   q.Seed,
		Rand:   q.Rand,
	}
}

func (q *Qubit) Fidelity(qb *Qubit) float64 {
	p0 := qb.Probability()
	p1 := q.Probability()

	var sum float64
	for i := 0; i < len(p0); i++ {
		sum = sum + math.Sqrt(p0[i]*p1[i])
	}

	return sum
}

func (q *Qubit) TraceDistance(qb *Qubit) float64 {
	p0 := qb.Probability()
	p1 := q.Probability()

	var sum float64
	for i := 0; i < len(p0); i++ {
		sum = sum + math.Abs(p0[i]-p1[i])
	}

	return sum / 2
}

func (q *Qubit) Equals(qb *Qubit, eps ...float64) bool {
	return q.vector.Equals(qb.vector, eps...)
}

func (q *Qubit) TensorProduct(qb *Qubit) *Qubit {
	q.vector = q.vector.TensorProduct(qb.vector)
	return q
}

func (q *Qubit) Apply(m ...matrix.Matrix) *Qubit {
	for _, mm := range m {
		q.vector = q.vector.Apply(mm)
	}

	return q
}

func (q *Qubit) Normalize() *Qubit {
	sum := number.Sum(q.Probability())
	z := 1 / math.Sqrt(sum)
	q.vector = q.vector.Mul(complex(z, 0))
	return q
}

func (q *Qubit) Amplitude() []complex128 {
	return q.vector.Complex()
}

func (q *Qubit) Probability() []float64 {
	p := make([]float64, 0)
	for _, a := range q.Amplitude() {
		p = append(p, math.Pow(cmplx.Abs(a), 2))
	}

	return p
}

func (q *Qubit) Measure(index int) *Qubit {
	zero, p := q.ProbabilityZeroAt(index)
	sum := number.Sum(p)

	r := q.Rand(q.Seed...)
	if r > sum {
		for _, i := range zero {
			q.vector[i] = complex(0, 0)
		}

		q.Normalize()
		return One()
	}

	one, _ := q.ProbabilityOneAt(index)
	for _, i := range one {
		q.vector[i] = complex(0, 0)
	}

	q.Normalize()
	return Zero()
}

func (q *Qubit) ProbabilityZeroAt(index int) ([]int, []float64) {
	dim := q.Dimension()
	den := int(math.Pow(2, float64(index+1)))
	div := dim / den

	p := q.Probability()
	idx, prob := make([]int, 0), make([]float64, 0)
	for i := 0; i < dim; i++ {
		idx, prob = append(idx, i), append(prob, p[i])

		if len(p) == dim/2 {
			break
		}

		if (i+1)%div == 0 {
			i = i + div
		}
	}

	return idx, prob
}

func (q *Qubit) ProbabilityOneAt(index int) ([]int, []float64) {
	z, _ := q.ProbabilityZeroAt(index)

	one := make([]int, 0)
	for i := range q.vector {
		found := false
		for _, zi := range z {
			if i == zi {
				found = true
				break
			}
		}

		if !found {
			one = append(one, i)
		}
	}

	p := q.Probability()
	idx, prob := make([]int, 0), make([]float64, 0)
	for _, i := range one {
		idx, prob = append(idx, i), append(prob, p[i])
	}

	return idx, prob
}

func (q *Qubit) Int() int {
	b := q.BinaryString()
	i, err := strconv.ParseInt(b, 2, 0)
	if err != nil {
		panic(err)
	}

	return int(i)
}

func (q *Qubit) BinaryString() string {
	var sb strings.Builder
	for i := 0; i < q.NumberOfBit(); i++ {
		if q.Clone().Measure(i).IsZero() {
			sb.WriteString("0")
			continue
		}

		sb.WriteString("1")
	}

	return sb.String()
}

func (q *Qubit) String() string {
	return fmt.Sprintf("%v", q.vector)
}

func (q *Qubit) State(index ...[]int) []State {
	if len(index) < 1 {
		idx := make([]int, 0)
		for i := 0; i < q.NumberOfBit(); i++ {
			idx = append(idx, i)
		}

		index = append(index, idx)
	}
	f := fmt.Sprintf("%s%s%s", "%0", strconv.Itoa(q.NumberOfBit()), "s")

	state := make([]State, 0)
	for i, a := range q.Amplitude() {
		if math.Abs(real(a)) < 1e-13 {
			a = complex(0, imag(a))
		}
		if math.Abs(imag(a)) < 1e-13 {
			a = complex(real(a), 0)
		}
		if a == 0 {
			continue
		}
		bin := fmt.Sprintf(f, strconv.FormatInt(int64(i), 2))

		s := State{Amplitude: a, Probability: math.Pow(cmplx.Abs(a), 2)}
		for _, idx := range index {
			bint, bbin := to(bin, idx)
			s.Int = append(s.Int, bint)
			s.BinaryString = append(s.BinaryString, bbin)
		}

		state = append(state, s)
	}

	return state
}

func to(binary string, idx []int) (int, string) {
	var sb strings.Builder
	for _, i := range idx {
		sb.WriteString(binary[i : i+1])
	}
	bin := sb.String()

	bint, err := strconv.ParseInt(bin, 2, 0)
	if err != nil {
		panic(fmt.Sprintf("parse int. binary=%s, bin=%s", binary, bin))
	}

	return int(bint), bin
}

func TensorProduct(qb ...*Qubit) *Qubit {
	q := qb[0]
	for i := 1; i < len(qb); i++ {
		q = q.TensorProduct(qb[i])
	}

	return q
}
