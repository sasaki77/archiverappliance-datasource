package archiverappliance

import "strings"

func isolateBasicQuery(unparsed string) []string {
	// Non-regex queries can request multiple PVs using this syntax: (PV:NAME:1|PV:NAME:2|...)
	// This function takes queries in this format and breaks them up into a list of individual PVs
	unparsed_clean := strings.TrimSpace(unparsed)

	phrases := locateOuterParen(unparsed_clean)
	// Identify parenthesis-bound sections
	multiPhrases := phrases.Phrases
	// Locate parenthesis-bound sections
	phraseIdxs := phrases.Idxs

	// If there are no sub-phrases in this string, return immediately to prevent further recursion
	if len(multiPhrases) == 0 {
		return []string{unparsed_clean}
	}

	// A list of all the possible phrases
	phraseParts := make([][]string, 0, len(multiPhrases))

	for idx := range multiPhrases {
		// Strip leading and ending parenthesis
		multiPhrases[idx] = multiPhrases[idx][1 : len(multiPhrases[idx])-1]
		// Break parsed phrases on "|"
		phraseParts = append(phraseParts, splitLowestLevelOnly(multiPhrases[idx]))
	}

	// list of all the configurations for the in-order phrases to be inserted
	phraseCase := permuteQuery(phraseParts)

	result := make([]string, 0, len(phraseCase))

	// Build results by substituting all phrase combinations in place for 1st-level substitutions
	for _, phrase := range phraseCase {
		createdString := selectiveInsert(unparsed_clean, phraseIdxs, phrase)
		result = append(result, createdString)
	}

	// For any phrase that has sub-phrases in need of parsing, call this function again on the sub-phrase and append the results to the end of the current output.
	for pos, chunk := range result {
		parseAttempt := isolateBasicQuery(chunk)
		if len(parseAttempt) > 1 {
			result = append(result[:pos], result[pos+1:]...) // pop partially parsed entry
			result = append(result, parseAttempt...)         // add new entires at the end of the list.
		}
	}

	return result
}

func splitLowestLevelOnly(inputData string) []string {
	output := make([]string, 0, 5)
	nestCounter := 0
	stashInitPos := 0
	for pos, char := range inputData {
		switch {
		case char == '(':
			nestCounter++
		case char == ')':
			nestCounter--
		}
		if char == '|' && nestCounter == 0 {
			output = append(output, inputData[stashInitPos:pos])
			stashInitPos = pos + 1
		}
	}
	output = append(output, inputData[stashInitPos:])
	return output
}

type ParenLoc struct {
	// Use to identify unenenclosed parenthesis, and their content
	Phrases []string
	Idxs    [][]int
}

func locateOuterParen(inputData string) ParenLoc {
	// read through a string to identify all paired sets of parentheses that are not contained within another set of parenthesis.
	// When found, report the indexes of all instantces as well as the content contained within the parenthesis.
	// This function ignores internal parenthesis.
	var output ParenLoc

	nestCounter := 0
	var stashInitPos int
	for pos, char := range inputData {
		if char == '(' {
			if nestCounter == 0 {
				stashInitPos = pos
			}
			nestCounter++
		}
		if char == ')' {
			if nestCounter == 1 {
				// A completed 1st-level parenthesis set has been completed
				output.Phrases = append(output.Phrases, inputData[stashInitPos:pos+1])
				output.Idxs = append(output.Idxs, []int{stashInitPos, pos + 1})
			}
			nestCounter--
		}
	}
	return output
}

func permuteQuery(inputData [][]string) [][]string {
	/*
	   Generate all ordered permutations of the input strings to make the following operation occur:

	   input:
	       {
	           {"a", "b"}
	           {"c", "d"}
	       }

	   output:
	       {
	           {"a", "c"}
	           {"a", "d"}
	           {"b", "c"}
	           {"b", "d"}
	       }
	*/
	return permuteQueryRecurse(inputData, []string{})
}

func permuteQueryRecurse(inputData [][]string, trace []string) [][]string {
	/*
	   recursive method for visitng all the permutations.
	*/

	// If you've assigned a value for each phrase, just return the full sequence of phrases
	if len(trace) == len(inputData) {
		output := make([][]string, 0, 1)
		output = append(output, trace)
		return output
	}

	/*
	   When values aren't assigned to all phrases, begin by assigning the first unassigned value
	   to each of the possible values and recursing on the result of that
	*/
	targetIdx := len(trace)
	output := make([][]string, 0, len(inputData[targetIdx]))
	for _, value := range inputData[targetIdx] {
		response := permuteQueryRecurse(inputData, append(trace, value))
		output = append(output, response...)
	}
	return output
}

func selectiveInsert(input string, idxs [][]int, inserts []string) string {
	/*
	   Selectively replace portions of the input string

	   idxs indicates the indices which will be removed

	   inserts provides the new string that will be put into the input string at the place designated by idxs

	   idxs and inserts should have the same length and are matched with each other element-wise to support any number of substitutions
	*/
	var builder strings.Builder

	prevIdx := 0

	// return a blank string if the arguments are bad
	if len(idxs) != len(inserts) {
		return ""
	}

	// At each place marked by idx, insert the corresponding new string
	for idx, val := range idxs {
		firstVal := val[0]
		builder.WriteString(input[prevIdx:firstVal])
		builder.WriteString(inserts[idx])
		prevIdx = val[1]
	}

	// Handle any trailing string
	if prevIdx < len(input) {
		lastVal := len(input)
		builder.WriteString(input[prevIdx:lastVal])
	}

	return builder.String()
}
