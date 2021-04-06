import _ from 'lodash';

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
