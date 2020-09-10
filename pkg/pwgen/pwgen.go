package pwgen

import (
	"bytes"
	"crypto/rand"
	"math/bits"
)

const (
	alnumAlpha = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func nextPowerOfTwo(x int) int {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x++
	return x
}

func calculateBitsPerRune(numChars int) int {
	nextPowerOfTwo := nextPowerOfTwo(numChars)
	bestBits := bits.TrailingZeros((uint)(nextPowerOfTwo))
	bestScore := numChars

	for numBits := bestBits + 1; numBits <= 32; numBits++ {
		m := (1 << numBits) % numChars
		if m == 0 {
			return numBits
		}

		if m > bestScore {
			bestBits = numBits
			bestScore = m
		}
	}

	return bestBits
}

// FromAlphabet returns a random string with a given alphabet of characters
// and length using the system's secure random source.
func FromAlphabet(alphabet string, length int) string {
	n := len(alphabet)
	bits := calculateBitsPerRune(n)
	totalBits := length * bits
	totalBytes := (totalBits + 7) / 8

	if totalBytes == 0 {
		return ""
	}

	random := make([]byte, totalBytes)
	_, err := rand.Read(random)
	if err != nil {
		panic(err)
	}

	randomOffset := 0
	buf := bytes.NewBuffer(make([]byte, 0, length))
	for i := 0; i < length; i++ {
		haveBits := 0
		index := 0

		for haveBits < bits {
			b := random[randomOffset/8]
			o := randomOffset % 8

			needed := bits - haveBits
			left := 8 - o
			take := left

			if needed < take {
				take = needed
			}

			index = (index << take) | (int)((b>>o)&((1<<take)-1))
			haveBits += take
			randomOffset += take
		}

		buf.WriteByte(alphabet[index%n])
	}

	return buf.String()
}

// AlphaNumeric returns a random string with `length` number of alpha-numeric
// characters.
func AlphaNumeric(length int) string {
	return FromAlphabet(alnumAlpha, length)
}
