import { createDataFrame, FieldType } from '@grafana/data';

import { seriesFunctions } from '../dataProcessor';
import { bench } from './bench';

function frames(points: number) {
  const times = Array.from({ length: points }, (_v, i) => i * 100);
  const values = times.map((t) => Math.sin(t));
  return [
    createDataFrame({
      name: 'PV',
      fields: [
        { name: 'time', type: FieldType.time, values: times },
        { name: 'PV', type: FieldType.number, values },
      ],
    }),
  ];
}

describe('functions', () => {
  it('transforms a series', async () => {
    const input = frames(100_000);
    for (const window of [10, 1000]) {
      await bench(
        `movingAverage, 100k points, window ${window}`,
        () => () => seriesFunctions.movingAverage(window, input)
      );
    }
  });
});
