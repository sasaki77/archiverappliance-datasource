/// <reference types="node" />
import fs from 'fs';
import path from 'path';
import { DataFrame, getFieldDisplayName } from '@grafana/data';

import { createFuncDescriptor, getCategories, getFuncDef } from '../aafunc';
import { applyFunctions } from '../query';
import { responseParse } from '../responseParse';
import { AADataQueryResponse, FuncDef, TargetQuery } from '../types';

// The vectors in testdata/functions are shared with pkg/functions/vectors_test.go,
// so that both query paths are held to the same results.
const vectorDir = path.resolve(__dirname, '../../testdata/functions');

type VectorValue = number | null | 'NaN' | number[];

interface VectorSeries {
  name: string;
  times: number[];
  values: VectorValue[];
}

interface VectorCase {
  name: string;
  input: VectorSeries[];
  functions: Array<{ name: string; params: string[] }>;
  output: VectorSeries[];
}

const readJSON = (file: string) => JSON.parse(fs.readFileSync(path.join(vectorDir, file), 'utf8'));

const decode = (v: VectorValue) => (v === 'NaN' ? NaN : v);

function buildTarget(c: VectorCase): TargetQuery {
  const functions = c.functions.map((f) => createFuncDescriptor(getFuncDef(f.name), f.params));
  return {
    refId: 'A',
    target: '',
    functions,
    operator: 'raw',
    interval: '',
    options: { disableExtrapol: 'true' },
  } as unknown as TargetQuery;
}

function buildResponses(c: VectorCase): AADataQueryResponse[] {
  const data = c.input.map((s) => ({
    meta: { name: s.name, waveform: s.values.some(Array.isArray), PREC: '0' },
    data: s.values.map((v, i) => ({ millis: s.times[i], val: decode(v) })),
  }));
  return [{ data } as unknown as AADataQueryResponse];
}

function toVector(frame: DataFrame) {
  const valueField = frame.fields[1];
  return {
    name: getFieldDisplayName(valueField, frame),
    times: [...frame.fields[0].values],
    values: [...valueField.values],
  };
}

function expectSameValues(want: VectorValue[], got: unknown[]) {
  expect(got).toHaveLength(want.length);
  want.forEach((w, i) => {
    const g = got[i];
    const d = decode(w);
    if (d === null) {
      expect(g).toBeNull();
    } else if (Number.isNaN(d)) {
      expect(g).toBeNaN();
    } else {
      // movingAverage carries a running total in the backend, which rounds in the last digits.
      expect(g).toBeCloseTo(d as number, 9);
    }
  });
}

const vectorCases = fs
  .readdirSync(vectorDir)
  .filter((file) => file.endsWith('.json') && file !== 'defs.json')
  .flatMap((file) => (readJSON(file) as VectorCase[]).map((c) => ({ title: `${file}/${c.name}`, c })));

describe('function vectors shared with the backend', () => {
  it.each(vectorCases)('$title', async ({ c }) => {
    const target = buildTarget(c);
    const parsed = await responseParse(buildResponses(c), target);
    const got = (await applyFunctions(parsed, target)).map(toVector);

    expect(got.map((s) => s.name)).toEqual(c.output.map((s) => s.name));
    c.output.forEach((want, i) => {
      expect(got[i].times).toEqual(want.times);
      expectSameValues(want.values, got[i].values);
    });
  });

  it('defs.json matches the function catalog', () => {
    const defs: FuncDef[] = readJSON('defs.json');
    const catalog = Object.values(getCategories())
      .flat()
      .map(({ name, category, params, defaultParams }) => ({ name, category, params, defaultParams }));

    expect(defs).toEqual(catalog);
  });
});
