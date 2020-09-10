package pwgen

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"unicode"
)

var _ = Describe("FromAlphabet", func() {
	It("should generate correct-length strings", func() {
		Expect(FromAlphabet("x", 10)).To(Equal("xxxxxxxxxx"))
		Expect(FromAlphabet("x", 3)).To(Equal("xxx"))
		Expect(FromAlphabet("x", 0)).To(Equal(""))
	})

	It("should only include alphabet characters", func() {
		chars := "0123456789"
		s := FromAlphabet(chars, 1000)
		for idx := 0; idx < len(s); idx++ {
			r := (rune)(s[idx])
			Expect(r).To(BeNumerically(">=", '0'))
			Expect(r).To(BeNumerically("<=", '9'))
		}
	})
})

var _ = Describe("AlphaNumeric", func() {
	It("should generate strings only containing alpha-numeric characters", func() {
		s := AlphaNumeric(1000)
		for idx := 0; idx < len(s); idx++ {
			r := (rune)(s[idx])
			Expect(unicode.IsDigit(r) || unicode.IsLetter(r)).To(BeTrue())
		}
	})
})
