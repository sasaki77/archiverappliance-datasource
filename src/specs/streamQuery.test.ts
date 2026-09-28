import { parseInterval } from '../streamQuery';

describe('parseInterval', () => {
  it.each([
    ['500', 500],
    ['100ms', 100],
    ['1s', 1000],
    ['1.5s', 1500],
    ['2m', 120000],
    ['1h', 3600000],
    ['1d', 86400000],
    ['1w', 604800000],
    [' 1s ', 1000],
    ['1S', 1000],
  ])('reads %s', (value, expected) => {
    expect(parseInterval(value)).toBe(expected);
  });

  // A negative interval used to reach setTimeout, which fires at once and
  // queries the archiver as fast as it answers.
  it.each(['', 'abc', '1 hour', '-1s', '0', '0s', '1y', '1 2s'])('leaves %s to the caller', (value) => {
    expect(parseInterval(value)).toBeUndefined();
  });
});
