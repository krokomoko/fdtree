package fdtree

import "github.com/krokomoko/fuzzy"

type __CData struct {
	order []__Position
	word  *fuzzy.Word
}

type __ConcrCalc struct {
	fdt      *FDTree
	data     [][]float64
	channels []chan __CData
	output   []chan float64
}

func __NewConcrCalc(fdt *FDTree, data [][]float64, div int) __ConcrCalc {
	var (
		_end int
		_d   = len(data) / div

		channels = make([]chan __CData, div)
		output   = make([]chan float64, div)
	)

	for i := 0; i < div; i++ {
		channels[i] = make(chan __CData, 1)
		output[i] = make(chan float64)

		if i == div-1 {
			_end = len(data)
		} else {
			_end = (i + 1) * _d
		}

		go __calc(fdt, data[i*_d:_end], channels[i], output[i])
	}

	return __ConcrCalc{
		fdt:      fdt,
		data:     data,
		channels: channels,
		output:   output,
	}
}

func (cc *__ConcrCalc) close() {
	for i := 0; i < len(cc.channels); i++ {
		close(cc.channels[i])
	}
}

func (cc *__ConcrCalc) calc(order []__Position, word ...*fuzzy.Word) float64 {
	var (
		sum   float64
		cdata = __CData{
			order: order,
		}
	)

	if len(word) > 0 && word[0] != nil {
		cdata.word = word[0]
	}

	for i := 0; i < len(cc.channels); i++ {
		cc.channels[i] <- cdata
	}

	for i := 0; i < len(cc.channels); i++ {
		sum += <-cc.output[i]
	}

	return sum
}

func __calc(fdt *FDTree, data [][]float64, input chan __CData, output chan float64) {
	var (
		cdata            __CData
		mult, _mult, sum float64
	)
	for cdata = range input {
		sum = 0
		for _, row := range data {
			mult = 1
			for _, position := range cdata.order {
				// TODO: гонка данных???
				_mult, _ = fdt.parameters[position.parameter].Words[position.word].Mu(row[position.parameter])
				mult *= _mult
			}
			if cdata.word != nil {
				_mult, _ = cdata.word.Mu(row[fdt.cInd])
				mult *= _mult
			}
			sum += mult
		}
		output <- sum
	}
}
