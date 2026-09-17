import { responseParse } from '../responseParse';
import { AADataQueryResponse, TargetQuery } from '../types';
import { bench } from './bench';

function waveforms(samples: number, width: number) {
  const data = Array.from({ length: samples }, (_v, i) => ({
    millis: i * 1000,
    val: Array.from({ length: width }, (_w, j) => j),
  }));
  return [{ data: [{ meta: { name: 'PV', waveform: true, PREC: '0' }, data }] }] as unknown as AADataQueryResponse[];
}

function target(arrayFormat: string) {
  return {
    refId: 'A',
    operator: 'raw',
    interval: '',
    functions: [],
    options: { arrayFormat, disableExtrapol: 'true' },
  } as unknown as TargetQuery;
}

describe('responseParse', () => {
  it('lays out waveforms', async () => {
    // The same number of elements, spread over more or fewer samples.
    for (const [samples, width] of [
      [100, 1000],
      [1000, 100],
      [10000, 10],
    ]) {
      const responses = waveforms(samples, width);
      for (const format of ['timeseries', 'index', 'dt-space']) {
        await bench(`${format}, ${samples} samples x ${width}`, () => () => responseParse(responses, target(format)));
      }
    }
  });
});
