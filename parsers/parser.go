package parsers

type Parser struct {
}

func (p *Parser) Parse(s string) []string {
	result := []string{}
	symbol := ""
	stringParsing := false
	for _, rune := range s {
		char := string(rune)
		if char == "'" {
			if stringParsing {
				result = appendIfNotEmpty(result, symbol)
				symbol = ""
				stringParsing = false
			} else {
				stringParsing = true
			}
			continue
		}

		if stringParsing {
			symbol += char
			continue
		}

		if char == " " {
			result = appendIfNotEmpty(result, symbol)
			symbol = ""
			continue
		}

		symbol += char
	}

	result = appendIfNotEmpty(result, symbol)

	return result
}

func appendIfNotEmpty(slice []string, value string) []string {
	if value != "" {
		return append(slice, value)
	}
	return slice
}
