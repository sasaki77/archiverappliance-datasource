/// <reference types="node" />
import fs from 'fs';
import path from 'path';

import { parseTargetPV } from '../pvnameParser';

// The cases are shared with pkg/archiverappliance/pvparser_test.go, so that both
// query paths expand a target into the same PVs.
const cases: Array<{ input: string; output: string[] }> = JSON.parse(
  fs.readFileSync(path.resolve(__dirname, '../../testdata/pvnames.json'), 'utf8')
);

describe('PV name expansion shared with the backend', () => {
  it.each(cases)('$input', ({ input, output }) => {
    expect(parseTargetPV(input)).toEqual(output);
  });
});
