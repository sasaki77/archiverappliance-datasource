import { AAclient } from '../aaclient';
import { doQueryStream } from '../streamQuery';
import { TargetQuery } from '../types';
import { bench } from './bench';

const PERIOD_MS = 100;
const TICK_MS = 1000;

function samples(from: number, to: number) {
  const data = [];
  for (let t = from - (from % PERIOD_MS) + PERIOD_MS; t <= to; t += PERIOD_MS) {
    data.push({ millis: t, val: t % 1000 });
  }
  return data;
}

// Answers every request with the samples between the range of the target, as the
// archiver would for a PV archived at 10 Hz.
function fakeClient(): AAclient {
  let range = { from: 0, to: 0 };
  return {
    buildUrls: async (target: TargetQuery) => {
      range = { from: target.from.getTime(), to: target.to.getTime() };
      return ['url'];
    },
    createUrlRequests: (urlsArray: string[][]) =>
      urlsArray.map((urls) =>
        urls.map(() =>
          Promise.resolve({ data: [{ meta: { name: 'PV', PREC: '0' }, data: samples(range.from, range.to) }] })
        )
      ),
  } as unknown as AAclient;
}

function target(points: number, strmCap: string): TargetQuery {
  const to = 1_700_000_000_000;
  return {
    refId: 'A',
    target: 'PV',
    operator: 'raw',
    interval: '',
    strmCap,
    maxDataPoints: 1000,
    functions: [],
    options: {},
    from: new Date(to - points * PERIOD_MS),
    to: new Date(to),
  } as unknown as TargetQuery;
}

// Each op is one tick: it fetches the last few seconds again, as the stream does,
// and merges the new samples into the buffer filled by the initial query.
function streamTicks(points: number, strmCap: string) {
  return async () => {
    const client = fakeClient();
    const buffers = {};
    let t = target(points, strmCap);
    await doQueryStream(client, [t], buffers);

    return async () => {
      const to = t.to.getTime() + TICK_MS;
      t = { ...t, from: new Date(t.to.getTime() - 2000), to: new Date(to) };
      return doQueryStream(client, [t], buffers);
    };
  };
}

describe('stream', () => {
  it('merges each tick into the buffer', async () => {
    await bench('stream tick, 1k points, strmCap 1000', streamTicks(1_000, '1000'));
    await bench('stream tick, 100k points, strmCap 1000', streamTicks(100_000, '1000'));
    await bench('stream tick, 100k points, default capacity', streamTicks(100_000, ''));
  });
});
