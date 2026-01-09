package audio

import (
	"fmt"
	"math"
	"math/cmplx"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/mjibson/go-dsp/fft"
)

const (
	SampleRate = 44100
	BufferSize = 1024
)

type Processor struct {
	Stream *portaudio.Stream

	// Input Analysis
	ComplexBuffer   []complex128
	Bass, Mid, High float64
	MaxLevel        float64

	// Smoothing
	bassSmooth, midSmooth, highSmooth float64

	// --- Synthesis (Soothing Pad Engine) ---
	Time        float64
	FilterState [2]float64   // Stereo LPF state
	DelayLine   [2][]float64 // Stereo Delay Buffer (Reverb-ish)
	DelayHead   int

	// Physics Inputs
	mu           sync.Mutex
	TotalEnergy  float64
	EnergySmooth float64 // For slow morphing

	Active bool
}

func NewProcessor() *Processor {
	// 0.6 second delay for larger space
	delayLen := int(float64(SampleRate) * 0.6)

	return &Processor{
		ComplexBuffer: make([]complex128, BufferSize),
		MaxLevel:      0.1,
		DelayLine:     [2][]float64{make([]float64, delayLen), make([]float64, delayLen)},
	}
}

func (a *Processor) Start() error {
	portaudio.Initialize()

	// DEBUG: explicit device info
	// devices, _ := portaudio.Devices()
	// for _, d := range devices { fmt.Println(d.Name) }

	// Output Only (0 In, 2 Out) to verify Sound works
	// Duplex (1, 2) often fails on Linux if devices differ
	stream, err := portaudio.OpenDefaultStream(0, 2, SampleRate, BufferSize, a.ProcessAudio)
	if err != nil {
		fmt.Printf("AUDIO ERROR: %v\n", err)
		return err
	}
	if err := stream.Start(); err != nil {
		fmt.Printf("STREAM START ERROR: %v\n", err)
		return err
	}

	fmt.Println("AUDIO STARTED: Output Only Mode")

	a.Stream = stream
	a.Active = true
	return nil
}

func (a *Processor) Stop() {
	if a.Stream != nil {
		a.Stream.Stop()
		a.Stream.Close()
	}
	portaudio.Terminate()
	a.Active = false
}

func (a *Processor) UpdatePhysics(energy, velocity, entropy float64) {
	a.mu.Lock()
	a.TotalEnergy = energy
	a.mu.Unlock()
}

// Triangle Wave: Smooth, flute-like, no harsh buzz
func triangle(phase float64) float64 {
	p := phase - math.Floor(phase)
	return 4.0*math.Abs(p-0.5) - 1.0
}

// Low Pass Filter (One Pole)
func lpf(sample, cutoff, dt, state float64) (float64, float64) {
	rc := 1.0 / (2.0 * math.Pi * cutoff)
	alpha := dt / (rc + dt)
	out := state + alpha*(sample-state)
	return out, out
}

func (a *Processor) ProcessAudio(in []float32, out [][]float32) {
	// 1. Input Analysis (FFT)
	for i, v := range in {
		window := 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(BufferSize-1)))
		a.ComplexBuffer[i] = complex(float64(v)*window, 0)
	}
	spectrum := fft.FFT(a.ComplexBuffer)

	// Bucketing
	bassSum, midSum, highSum := 0.0, 0.0, 0.0
	for i := 0; i < BufferSize/2; i++ {
		mag := cmplx.Abs(spectrum[i])
		if i < 5 {
			bassSum += mag
		} else if i < 46 {
			midSum += mag
		} else if i < 460 {
			highSum += mag
		}
	}

	// AGC
	peak := math.Max(bassSum/100.0, math.Max(midSum/500.0, highSum/1000.0))
	if peak > a.MaxLevel {
		a.MaxLevel = peak
	} else {
		a.MaxLevel *= 0.999
	}
	gain := 1.0
	if a.MaxLevel > 0.001 {
		gain = 1.0 / a.MaxLevel
	}
	if gain > 50.0 {
		gain = 50.0
	}

	a.Bass = a.Bass*0.9 + (math.Min(bassSum/100.0*gain, 1.0))*0.1
	a.Mid = a.Mid*0.9 + (math.Min(midSum/500.0*gain, 1.0))*0.1
	a.High = a.High*0.9 + (math.Min(highSum/1000.0*gain, 1.0))*0.1

	// 2. Synthesis (Smooth/Ambient)
	// Harmony: Gm7 Add9 (Classic Ambient Stack)
	// G2, Bb2, D3, F3, A3
	freqs := []float64{98.00, 116.54, 146.83, 174.61, 220.00}

	a.mu.Lock()
	targetEnergy := a.TotalEnergy
	a.mu.Unlock()

	// Slow Morphing of Energy to prevent jumps
	a.EnergySmooth = a.EnergySmooth*0.995 + targetEnergy*0.005

	// Mellow Filter Control
	// Dynamic Cutoff: Energy opens the filter
	// Base: 300Hz (Visible/Audible) -> Max: 1200Hz
	cutoff := 300.0 + math.Min(a.EnergySmooth/5.0, 900.0)
	dt := 1.0 / float64(SampleRate)

	// Master Volume (Boosted)
	vol := 0.252

	for i := 0; i < len(out[0]); i++ {
		sampleL := 0.0
		sampleR := 0.0

		for j, f := range freqs {
			// Triangle Waves (No Saw Buzz)
			// Slight Detune
			oscL := triangle(a.Time * (f * 0.999))
			oscR := triangle(a.Time * (f * 1.001))

			g := 1.0 / float64(len(freqs))

			// Very Slow LFO (Breathing)
			lfo := math.Sin(a.Time*0.2 + float64(j))

			sampleL += oscL * g * (0.7 + 0.3*lfo)
			sampleR += oscR * g * (0.7 + 0.3*lfo)
		}

		// Filter (Smoothes triangles even further into pure sine-ish tones)
		var outL, outR float64
		outL, a.FilterState[0] = lpf(sampleL, cutoff, dt, a.FilterState[0])
		outR, a.FilterState[1] = lpf(sampleR, cutoff, dt, a.FilterState[1])

		// Delay/Reverb (Longer, more diffuse)
		delayL := a.DelayLine[0][a.DelayHead]
		delayR := a.DelayLine[1][a.DelayHead]

		// Feedback Cross-Talk (Ping Pong)
		// L hears a bit of R's delay, smears the stereo image
		mixL := outL + delayL*0.3 + delayR*0.1
		mixR := outR + delayR*0.3 + delayL*0.1

		a.DelayLine[0][a.DelayHead] = mixL * 0.7 // High Feedback (Long Tail)
		a.DelayLine[1][a.DelayHead] = mixR * 0.7

		a.DelayHead = (a.DelayHead + 1) % len(a.DelayLine[0])

		out[0][i] = float32(mixL * vol)
		out[1][i] = float32(mixR * vol)

		a.Time += dt
	}
}

func (a *Processor) Update() {}
