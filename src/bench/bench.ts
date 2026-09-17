type Op = () => unknown;

const RUNS = 5;
const MIN_RUN_MS = 100;

// Present when node runs with --expose-gc, as `yarn bench` does. Without it the
// retained heap is not measured.
const gc = (globalThis as { gc?: () => void }).gc;

function heapUsed() {
  gc?.();
  return process.memoryUsage().heapUsed;
}

const median = (xs: number[]) => [...xs].sort((a, b) => a - b)[Math.floor(xs.length / 2)];

// Repeats the op returned by setup for at least MIN_RUN_MS, RUNS times over, and
// prints the median time per op and the median heap still held by the state of a
// run, such as a stream buffer.
export async function bench(name: string, setup: () => Op | Promise<Op>) {
  await (
    await setup()
  )();

  const times: number[] = [];
  const retained: number[] = [];
  for (let r = 0; r < RUNS; r++) {
    const before = heapUsed();
    const keep = [await setup()];
    const start = performance.now();
    let ops = 0;
    do {
      await keep[0]();
      ops++;
    } while (performance.now() - start < MIN_RUN_MS);
    times.push((performance.now() - start) / ops);
    retained.push(Math.max(0, heapUsed() - before));
    keep.pop();
  }

  const heap = gc ? `${(median(retained) / 1024).toFixed(0).padStart(8)} KB retained` : '';
  process.stdout.write(`${name.padEnd(48)}${median(times).toFixed(3).padStart(12)} ms/op${heap}\n`);
}
