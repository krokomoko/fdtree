package fdtree

import (
	"math"
)

type Position struct {
	parameter int
	word      int
}

type FDTree struct {
	cInd       int
	parameters []Parameter
	p          [][]float64
	order      [][]Position
}

func NewFDTree(data [][]float64, wordsCount int) FDTree {
	var (
		columns    = len(data[0])
		rows       = len(data)
		parameters = make([]Parameter, columns)
		values     = make([]float64, rows)
	)

	for i := 0; i < columns; i++ {
		for j := 0; j < rows; j++ {
			values[j] = data[j][i]
		}
		parameters[i] = NewParameter(values, wordsCount)
	}

	return FDTree{
		cInd:       columns - 1,
		parameters: parameters,
		p:          [][]float64{},
		order:      [][]Position{},
	}
}

func (fdt *FDTree) entropy(data [][]float64, order []Position) (float64, float64) {
	// общее количество элементов в узле
	var nTotal float64

	if len(order) == 0 {
		nTotal = float64(len(data))
	} else {
		var mult float64
		for _, row := range data {
			mult = 1
			for _, position := range order {
				mult *= fdt.parameters[position.parameter].Words[position.word].mu(row[position.parameter])
			}
			nTotal += mult
		}
	}

	if nTotal == 0.0 {
		return 0.0, 0.0
	}

	// подсчёт энтропии узла и возврат из функции
	var r, p, mult float64

	for _, word := range fdt.parameters[fdt.cInd].Words {
		p = 0
		for _, row := range data {
			mult = 1
			for _, position := range order {
				mult *= fdt.parameters[position.parameter].Words[position.word].mu(row[position.parameter])
			}
			p += word.mu(row[fdt.cInd]) * mult
		}
		p /= nTotal
		if p != 0.0 {
			p = math.Log2(p)
		}
		r += p
	}

	return -1.0 * r, nTotal
}

func (fdt *FDTree) Feet(data [][]float64) {
	var (
		order   = []Position{}
		indexes = make([]int, fdt.cInd)
	)

	for i := 0; i < fdt.cInd; i++ {
		indexes[i] = i
	}

	fdt.p = [][]float64{}
	fdt.order = [][]Position{}

	fdt.feet(data, indexes, order)
}

func (fdt *FDTree) feet(data [][]float64, indexes []int, order []Position) {
	var (
		maxInd     int
		indexesLen int = len(indexes)
	)

	if indexesLen == 0 {
		var n, mult float64

		for _, row := range data {
			mult = 1
			for _, position := range order {
				mult *= fdt.parameters[position.parameter].Words[position.word].mu(row[position.parameter])
			}
			n += mult
		}

		if n != 0.0 {
			var ps = make([]float64, len(fdt.parameters[fdt.cInd].Words))

			for i, word := range fdt.parameters[fdt.cInd].Words {
				for _, row := range data {
					mult = 1
					for _, position := range order {
						mult *= fdt.parameters[position.parameter].Words[position.word].mu(row[position.parameter])
					}
					ps[i] += word.mu(row[fdt.cInd]) * mult
				}
				ps[i] /= n
			}

			fdt.p = append(fdt.p, ps)

			var _order = make([]Position, len(order))
			copy(_order, order)
			fdt.order = append(fdt.order, _order)
		}
		return

	} else if indexesLen == 1 {
		maxInd = 0

	} else {
		// энтропия родительского узла
		var eParent, nParent float64 = fdt.entropy(data, order)

		// рассмотрение в отдельности каждого оставшегося параметра
		var gains = make([]float64, len(indexes))

		if nParent != 0 {
			var (
				gainSum, e, n float64
				lnOrder       = len(order)
				_order        = make([]Position, lnOrder+1)
			)
			copy(_order, order)

			for i, parameterIndex := range indexes {
				for wordInd := range fdt.parameters[parameterIndex].Words {
					_order[lnOrder] = Position{parameter: parameterIndex, word: wordInd}
					e, n = fdt.entropy(data, _order)
					gainSum += e * n / nParent
				}
				gains[i] = eParent - gainSum
			}
		} else {
			for i := 0; i < len(indexes); i++ {
				gains[i] = eParent
			}
		}

		var max float64
		for i, gain := range gains {
			if max < gain {
				max = gain
				maxInd = i
			}
		}
	}

	newIndexes := []int{}
	for _, ind := range indexes {
		if ind != indexes[maxInd] {
			newIndexes = append(newIndexes, ind)
		}
	}

	var (
		lnOrder = len(order)
		_order  = make([]Position, lnOrder+1)
	)
	copy(_order, order)

	for wordIndex := 0; wordIndex < len(fdt.parameters[indexes[maxInd]].Words); wordIndex++ {
		_order[lnOrder] = Position{parameter: indexes[maxInd], word: wordIndex}
		fdt.feet(data, newIndexes, _order)
	}
}

func (fdt *FDTree) Predict(data []float64) float64 {
	var mult float64
	wordsMu := make([]float64, len(fdt.parameters[fdt.cInd].Words))

	for wordInd := 0; wordInd < len(fdt.parameters[fdt.cInd].Words); wordInd++ {
		for leafInd := 0; leafInd < len(fdt.p); leafInd++ {
			mult = 1
			for _, position := range fdt.order[leafInd] {
				mult *= fdt.parameters[position.parameter].Words[position.word].mu(data[position.parameter])
			}
			wordsMu[wordInd] += fdt.p[leafInd][wordInd] * mult
		}
	}

	result, _ := fdt.parameters[fdt.cInd].value(wordsMu)
	return result
}
