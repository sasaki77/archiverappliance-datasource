//import {locateOuterParen, permuteQuery, selectiveInsert} from '../utils'
import {locateOuterParen, permuteQuery, splitLowestLevelOnly, selectiveInsert} from '../utils'

describe('Utils tests', () => {
  it.each([
    ['one (two (three)) (four five) six', ['(two (three))','(four five)'], [[4,17],[18,29]]],
    ['one (match) (here)', ['(match)', '(here)'], [[4, 11], [12, 18]]],
    ['one (match) here', ['(match)'], [[4, 11]]],
    ['no matches here', [], []]
  ])('locateOuterParen test', (text, phrases, idxs) => {
    const data = locateOuterParen(text);
    expect(data.phrases).toStrictEqual(phrases);
    expect(data.idxs).toStrictEqual(idxs);
  });

  it.each([
    [[['a', 'd'], ['b', 'c']], [['a', 'b'], ['a', 'c'], ['d', 'b'], ['d', 'c']]],
    [[['a'], ['b', 'c']], [['a', 'b'], ['a', 'c']]],
    [[['a'], ['b']], [['a', 'b']]],
    [[['a']], [['a']]],
    [[['a', 'b']], [['a'], ['b']]],
    [[[]], []],
  ])('permuteQuery test', (input, output) => {
    const data = permuteQuery(input);
    expect(data).toStrictEqual(output);
  });

  it.each([
    ["one (two (three)) (four five) six", ["one (two (three)) (four five) six"]],
    ["one two |three", ["one two ", "three"]],
    ["one two |three | four", ["one two ", "three ", " four"]],
    ["one (two |three) | four", ["one (two |three) ", " four"]]
  ])('splitLowestLevelOnly test', (input, output) => {
    const data = splitLowestLevelOnly(input);
    expect(data).toStrictEqual(output);
  });

  it.each([
    ['hello there', [[0, 5]], ['be'], 'be there'],
    ['hello there', [[6, 11]], ['friend'], 'hello friend'],
    ['hello there', [[0, 4], [6, 11]], ['y', "what's up"], "yo what's up"],
    ["This won't work", [[0, 4], [6, 11]], ['y'], ''],
  ])('selectiveInsert test', (input, idxs, inserts, output) => {
    const data = selectiveInsert(input, idxs, inserts);
    expect(data).toStrictEqual(output);
  });
});
