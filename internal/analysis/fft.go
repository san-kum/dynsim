package analysis

import (
	"math"
	"math/cmplx"
)

func FFT(data []float64) []complex128 {
	n := len(data)
	if n <= 1 {
		result := make([]complex128, n)
		for i := range data {
			result[i] = complex(data[i], 0)
		}
		return result
	}

	if n%2 != 0 {
		panic("fft requires power of 2 length")
	}

	even := make([]float64, n/2)
	odd := make([]float64, n/2)

	for i := 0; i < n/2; i++ {
		even[i] = data[2*i]
		odd[i] = data[2*i+1]
	}

	feven := FFT(even)
	fodd := FFT(odd)

	result := make([]complex128, n)
	for k := 0; k < n/2; k++ {
		w := cmplx.Exp(complex(0, -2*math.Pi*float64(k)/float64(n)))
		result[k] = feven[k] + w*fodd[k]
		result[k+n/2] = feven[k] - w*fodd[k]
	}

	return result
}

func PowerSpectrum(data []float64) []float64 {
	fft := FFT(data)
	ps := make([]float64, len(fft)/2)

	for i := range ps {
		ps[i] = cmplx.Abs(fft[i])
	}

	return ps
}
