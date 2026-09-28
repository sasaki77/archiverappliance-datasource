import { createDataFrame, DataFrame, getFieldDisplayName } from '@grafana/data';

// Transform

function scale(factor: number, times: number[], values: number[]) {
  return {
    times: times,
    values: values.map((value) => value * factor),
  };
}

function offset(delta: number, times: number[], values: number[]) {
  return {
    times: times,
    values: values.map((value) => value + delta),
  };
}

function delta(times: number[], values: number[]) {
  const newTimes = [];
  const newValues = [];

  for (let i = 1; i < values.length; i += 1) {
    const deltaValue = values[i] - values[i - 1];
    newTimes.push(times[i]);
    newValues.push(deltaValue);
  }

  return {
    times: newTimes,
    values: newValues,
  };
}

function fluctuation(times: number[], values: number[]) {
  const newSeries = [];

  for (let i = 0; i < values.length; i += 1) {
    const flucValue = values[i] - values[0];
    newSeries.push(flucValue);
  }

  return {
    times: times,
    values: newSeries,
  };
}

function movingAverage(windowSize: number, times: number[], values: number[]) {
  if (windowSize < 1) {
    return {
      times: times,
      values: values,
    };
  }

  const newSeries = new Array<number>(values.length);
  let total = 0;

  for (let i = 0; i < values.length; i++) {
    total += values[i];
    if (i >= windowSize) {
      total -= values[i - windowSize];
    }
    newSeries[i] = total / Math.min(i + 1, windowSize);
  }

  return {
    times: times,
    values: newSeries,
  };
}

// [Support Funcs] Transform wrapper

function transformWrapper(func: (...args: any) => { times: number[]; values: number[] }, ...args: any) {
  const funcArgs = args.slice(0, -1);
  const dataFrames: DataFrame[] = args[args.length - 1];

  const tsData = dataFrames.map((dataFrame) => {
    const timesField = dataFrame.fields[0];
    const valField = dataFrame.fields[1];
    const vals = func(...funcArgs, timesField.values, valField.values);

    const newTimesField = {
      ...timesField,
      values: vals.times,
    };

    const newValfield = {
      ...valField,
      values: vals.values,
    };

    return createDataFrame({
      ...dataFrame,
      fields: [newTimesField, newValfield],
    });
  });

  return tsData;
}

// Filter Series
function exclude(pattern: string, dataFrames: DataFrame[]) {
  let regex: RegExp;
  try {
    regex = new RegExp(pattern);
  } catch {
    return dataFrames;
  }

  return dataFrames.filter((dataFrame) => {
    const valfield = dataFrame.fields[1];
    const displayName = getFieldDisplayName(valfield, dataFrame);
    return !regex.test(displayName);
  });
}

// [Support Funcs] Datapoints aggregation functions

function sum(values: number[]) {
  let total = 0;
  for (const value of values) {
    total += value;
  }
  return total;
}

function mean(values: number[]) {
  return sum(values) / values.length;
}

// Both walk the values as the backend does (minimum and maximum in
// pkg/functions/arrayfuncs.go): a NaN compares neither smaller nor larger, so it
// is carried only when it comes first. An empty series has neither.
function minimum(values: number[]): number | undefined {
  let min = values[0];
  for (const value of values) {
    if (value < min) {
      min = value;
    }
  }
  return min;
}

function maximum(values: number[]): number | undefined {
  let max = values[0];
  for (const value of values) {
    if (value > max) {
      max = value;
    }
  }
  return max;
}

function datapointsAvg(values: number[]) {
  return mean(values);
}

function datapointsMin(values: number[]) {
  return minimum(values);
}

function datapointsMax(values: number[]) {
  return maximum(values);
}

function datapointsSum(values: number[]) {
  return sum(values);
}

function datapointsMed(values: number[]) {
  const sorted = [...values].sort((a, b) => a - b);
  const mid = sorted.length >> 1;

  if (sorted.length % 2 !== 0) {
    return sorted[mid];
  }
  return (sorted[mid - 1] + sorted[mid]) / 2;
}

function datapointsStd(values: number[]) {
  const average = mean(values);
  const variance = values.reduce((total, value) => total + (value - average) * (value - average), 0);

  return Math.sqrt(variance / values.length);
}

// Mirrors Scalars.Rank and cmp.Compare in the backend: an empty series ranks as
// 0, except by avg, where its NaN ranks below every number.
const rankFuncs = new Map<string, (values: number[]) => number>([
  ['avg', (values) => mean(values)],
  ['min', (values) => minimum(values) ?? 0],
  ['max', (values) => maximum(values) ?? 0],
  ['sum', (values) => sum(values)],
  ['absoluteMin', (values) => minimum(values.map(Math.abs)) ?? 0],
  ['absoluteMax', (values) => maximum(values.map(Math.abs)) ?? 0],
]);

function compareRank(a: number, b: number) {
  if (Number.isNaN(a)) {
    return Number.isNaN(b) ? 0 : -1;
  }
  if (Number.isNaN(b)) {
    return 1;
  }
  return a < b ? -1 : a > b ? 1 : 0;
}

function sortByRank(dataFrames: DataFrame[], rankFunc: string, descending: boolean) {
  const rank = rankFuncs.get(rankFunc);
  if (!rank) {
    return dataFrames;
  }

  return dataFrames
    .map((frame) => ({ frame, rank: rank(frame.fields[1].values) }))
    .sort((a, b) => (descending ? compareRank(b.rank, a.rank) : compareRank(a.rank, b.rank)))
    .map(({ frame }) => frame);
}

// [Support Funcs] Wrapper function for top and bottom function

function extraction(order: string, n: number, orderFunc: string, dataFrames: DataFrame[]) {
  if (n < 0 || !rankFuncs.has(orderFunc)) {
    return dataFrames;
  }

  return sortByRank(dataFrames, orderFunc, order === 'top').slice(0, n);
}

// [Support Funcs] Wrapper function for sort by AggFuncs
function sortByAggFuncs(orderFunc: string, order: string, dataFrames: DataFrame[]) {
  if (order !== 'asc' && order !== 'desc') {
    return dataFrames;
  }

  return sortByRank(dataFrames, orderFunc, order !== 'asc');
}

// Function list

// Each entry takes the parameters of the function and the frames last, as
// bindFunction in aafunc.ts calls it.
const functions: { [name: string]: (...args: any[]) => DataFrame[] } = {
  // Transform
  scale: (factor: number, frames: DataFrame[]) => transformWrapper(scale, factor, frames),
  offset: (delta: number, frames: DataFrame[]) => transformWrapper(offset, delta, frames),
  delta: (frames: DataFrame[]) => transformWrapper(delta, frames),
  fluctuation: (frames: DataFrame[]) => transformWrapper(fluctuation, frames),
  movingAverage: (windowSize: number, frames: DataFrame[]) => transformWrapper(movingAverage, windowSize, frames),
  // Filter Series
  top: (n: number, orderFunc: string, frames: DataFrame[]) => extraction('top', n, orderFunc, frames),
  bottom: (n: number, orderFunc: string, frames: DataFrame[]) => extraction('bottom', n, orderFunc, frames),
  exclude,
  // Sort
  sortByAvg: (order: string, frames: DataFrame[]) => sortByAggFuncs('avg', order, frames),
  sortByMax: (order: string, frames: DataFrame[]) => sortByAggFuncs('max', order, frames),
  sortByMin: (order: string, frames: DataFrame[]) => sortByAggFuncs('min', order, frames),
  sortBySum: (order: string, frames: DataFrame[]) => sortByAggFuncs('sum', order, frames),
  sortByAbsMax: (order: string, frames: DataFrame[]) => sortByAggFuncs('absoluteMax', order, frames),
  sortByAbsMin: (order: string, frames: DataFrame[]) => sortByAggFuncs('absoluteMin', order, frames),
};

// An empty waveform reduces to NaN, as in the backend, which draws a gap.
function nonEmpty(reduce: (values: number[]) => number | undefined) {
  return (values: number[]) => (values.length === 0 ? NaN : reduce(values));
}

const arrayFunctions: { [key: string]: { func: any; label: string } } = {
  toScalarByAvg: { func: nonEmpty(datapointsAvg), label: 'avg' },
  toScalarByMax: { func: nonEmpty(datapointsMax), label: 'max' },
  toScalarByMin: { func: nonEmpty(datapointsMin), label: 'min' },
  toScalarBySum: { func: nonEmpty(datapointsSum), label: 'sum' },
  toScalarByMed: { func: nonEmpty(datapointsMed), label: 'median' },
  toScalarByStd: { func: nonEmpty(datapointsStd), label: 'std' },
};

export { functions as seriesFunctions, arrayFunctions };
