package services

import "unicode"

func ValidLuhn(number string) bool {
	var total int
	var double bool

	for i := len(number) - 1; i >= 0; i-- {
		if !unicode.IsDigit(rune(number[i])) {
			return false
		}

		digit := int(number[i] - '0')
		if double {
			digit = digit * 2
			if digit >= 10 {
				digit -= 9
			}
		}
		total += digit

		double = !double
	}

	return total > 0 && total%10 == 0
}
