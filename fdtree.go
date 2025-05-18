package fdtree

import (
	"math"

	"github.com/krokomoko/fuzzy"
)

type __Position struct {
	parameter int
	word      int
}

type FDTree struct {
	cInd       int
	parameters []fuzzy.Parameter
	p          [][]float64
	order      [][]__Position
}

func NewFDTree(data [][]float64, wordsCount int) FDTree {
	var (
		columns    = len(data[0])
		rows       = len(data)
		parameters = make([]fuzzy.Parameter, columns)
		values     = make([]float64, rows)
	)

	for i := 0; i < columns; i++ {
		for j := 0; j < rows; j++ {
			values[j] = data[j][i]
		}
		parameters[i] = fuzzy.NewParameter(values, wordsCount)
	}

	return FDTree{
		cInd:       columns - 1,
		parameters: parameters,
		p:          [][]float64{},
		order:      [][]__Position{},
	}
}

func (fdt *FDTree) entropy(data [][]float64, order []__Position, cc *__ConcrCalc) (float64, float64) {
	// общее количество элементов в узле
	var nTotal float64

	if len(order) == 0 {
		nTotal = float64(len(data))
	} else {
		nTotal = cc.calc(order)
	}

	if nTotal == 0.0 {
		return 0.0, 0.0
	}

	// подсчёт энтропии узла и возврат из функции
	var r, p float64

	for _, word := range fdt.parameters[fdt.cInd].Words {
		p = cc.calc(order, &word) / nTotal
		if p != 0.0 {
			p = math.Log2(p)
		}
		r += p
	}

	return -1.0 * r, nTotal
}

func (fdt *FDTree) Feet(data [][]float64, depth int, calcDiv ...int) {
	var (
		order        = []__Position{}
		indexes      = make([]int, fdt.cInd)
		_calcDiv int = 1
	)

	for i := 0; i < fdt.cInd; i++ {
		indexes[i] = i
	}

	fdt.p = [][]float64{}
	fdt.order = [][]__Position{}

	if len(calcDiv) > 0 {
		_calcDiv = calcDiv[0]
	}
	concurrentCalculator := __NewConcrCalc(fdt, data, _calcDiv)
	defer concurrentCalculator.close()

	fdt.feet(data, indexes, order, &concurrentCalculator, depth, 0)
}

func (fdt *FDTree) feet(data [][]float64, indexes []int, order []__Position, cc *__ConcrCalc, maxDepth, depth int) {
	var (
		maxInd      int
		indexesLen  int = len(indexes)
		childsCount     = cc.calc(order)
	)

	if childsCount == 0 {
		return
	}

	if indexesLen == 0 || depth >= maxDepth {
		var ps = make([]float64, len(fdt.parameters[fdt.cInd].Words))

		for i, word := range fdt.parameters[fdt.cInd].Words {
			ps[i] = cc.calc(order, &word) / childsCount
		}

		fdt.p = append(fdt.p, ps)

		var _order = make([]__Position, len(order))
		copy(_order, order)
		fdt.order = append(fdt.order, _order)

		return

	} else if indexesLen == 1 {
		maxInd = 0

	} else {
		// энтропия родительского узла
		var eParent, nParent float64 = fdt.entropy(data, order, cc)

		// рассмотрение в отдельности каждого оставшегося параметра
		var gains = make([]float64, len(indexes))

		if nParent != 0 {
			var (
				gainSum, e, n float64
				lnOrder       = len(order)
				_order        = make([]__Position, lnOrder+1)
			)
			copy(_order, order)

			for i, parameterIndex := range indexes {
				for wordInd := range fdt.parameters[parameterIndex].Words {
					_order[lnOrder] = __Position{parameter: parameterIndex, word: wordInd}
					e, n = fdt.entropy(data, _order, cc)
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
		_order  = make([]__Position, lnOrder+1)
		_depth  = depth + 1
	)
	copy(_order, order)

	for wordIndex := 0; wordIndex < len(fdt.parameters[indexes[maxInd]].Words); wordIndex++ {
		_order[lnOrder] = __Position{parameter: indexes[maxInd], word: wordIndex}
		fdt.feet(data, newIndexes, _order, cc, maxDepth, _depth)
	}
}

func (fdt *FDTree) Predict(data []float64) float64 {
	var mult, _mult float64
	wordsMu := make([]float64, len(fdt.parameters[fdt.cInd].Words))

	for wordInd := 0; wordInd < len(fdt.parameters[fdt.cInd].Words); wordInd++ {
		for leafInd := 0; leafInd < len(fdt.p); leafInd++ {
			mult = 1
			for _, position := range fdt.order[leafInd] {
				_mult, _ = fdt.parameters[position.parameter].Words[position.word].Mu(data[position.parameter])
				mult *= _mult
			}
			wordsMu[wordInd] += fdt.p[leafInd][wordInd] * mult
		}
	}

	result, _ := fdt.parameters[fdt.cInd].Value(wordsMu)
	return result
}
