import _ from 'lodash';

export function parseTargetPV(targetPV: string): string[] {
  // ex) A(1(2))(3|4)B
  // parenPhraseData = {phrases: [(1(2)), (3|4)], idxs: [[1,6], [7,11]]}
  // phraseParts = [[1(2)], [3,4]]
  // 1stresult = [A1(2)3B, A1(2)4B]
  // 2ndresult = [A123B, A124B]
  const parenPhraseData = locateOuterParen(targetPV);

  // Identify parenthesis-bound sections
  const parenPhrases = parenPhraseData.phrases;

  // If there are no sub-phrases in this string, return immediately to prevent further recursion
  if (parenPhrases.length === 0) {
    return [targetPV];
  }

  // A list of all the possible phrases
  const phraseParts = _.map(parenPhrases, (phrase, i) => {
    const stripedPhrase = phrase.slice(1, phrase.length - 1);
    return splitLowestLevelOnly(stripedPhrase);
  });

  // list of all the configurations for the in-order phrases to be inserted
  const phraseCase = permuteQuery(phraseParts);

  //// Build results by substituting all phrase combinations in place for 1st-level substitutions
  const result = _.map(phraseCase, (phrase) => selectiveInsert(targetPV, parenPhraseData.idxs, phrase));

  // For any phrase that has sub-phrases in need of parsing, call this function again on the sub-phrase and append the results to the end of the current output.
  result.forEach((chunk, pos) => {
    const parseAttempt = parseTargetPV(chunk);
    if (parseAttempt.length > 1) {
      result.splice(pos, 1); // pop partially parsed entry
      result.push(...parseAttempt); // add new entires at the end of the list.
    }
  });

  return result;
}

export function locateOuterParen(data: string): { phrases: string[]; idxs: number[][] } {
  const phrases: string[] = [];
  const idxs: number[][] = [];

  let nestCounter = 0;
  let stashInitPos = 0;
  const chars = [...data];

  chars.forEach((char, pos) => {
    if (char == '(') {
      if (nestCounter == 0) {
        stashInitPos = pos;
      }
      nestCounter++;
    } else if (char == ')') {
      if (nestCounter == 1) {
        phrases.push(data.slice(stashInitPos, pos + 1));
        idxs.push([stashInitPos, pos + 1]);
      }
      nestCounter--;
    }
  });

  return { phrases, idxs };
}

export function permuteQuery(input: string[][]): string[][] {
  /*
      Generate all ordered permutations of the input strings to make the following operation occur: 

      input:
          {
              {"a", "b"}
              {"c", "d"}
              {"e", "f"}
          }

      output:
          {
              {"a", "c", "e"}
              {"a", "c", "f"}
              {"a", "d", "e"}
              {"a", "d", "f"}
              {"b", "c", "e"}
              {"b", "c", "f"}
              {"b", "d", "e"}
              {"b", "d", "f"}
          }
  */
  const output = _.reduce(
    input,
    (permutedArray: string[][], pushStrings: string[]) => {
      if (permutedArray.length === 0) {
        return _.map(pushStrings, (str) => [str]);
      }

      return _.reduce(
        permutedArray,
        (newPermutedArray: string[][], permutedStrs: string[]) => {
          for (const str of pushStrings) {
            newPermutedArray.push([...permutedStrs, str]);
          }
          return newPermutedArray;
        },
        []
      );
    },
    []
  );

  return output;
}

export function splitLowestLevelOnly(data: string): string[] {
  let nestCounter = 0;
  let stashInitPos = 0;
  const output: string[] = [];
  const chars = [...data];

  chars.forEach((char, pos) => {
    if (char === '(') {
      nestCounter++;
    } else if (char === ')') {
      nestCounter--;
    }
    if (char === '|' && nestCounter === 0) {
      output.push(data.slice(stashInitPos, pos));
      stashInitPos = pos + 1;
    }
  });
  output.push(data.slice(stashInitPos));
  return output;
}

export function selectiveInsert(input: string, idxs: number[][], inserts: string[]): string {
  if (idxs.length !== inserts.length) {
    return '';
  }

  let prevIdx = 0;
  let output = '';

  idxs.forEach((val, idx) => {
    const startIdx = val[0];
    output = output.concat('', input.slice(prevIdx, startIdx));
    output = output.concat('', inserts[idx]);
    prevIdx = val[1];
  });

  // Handle any trailing string
  if (prevIdx < input.length) {
    const lastIdx = input.length;
    output = output.concat('', input.slice(prevIdx, lastIdx));
  }

  return output;
}
