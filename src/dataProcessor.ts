import _ from 'lodash';
import { createDataFrame, DataFrame, getFieldDisplayName } from '@grafana/data';
import * as math from 'mathjs';

// Transform

function scale(factor: number, times: number[], values: number[]) {
  return {
    times: times,
    values: _.map(values, (value) => value * factor),
  };
}

function offset(delta: number, times: number[], values: number[]) {
  return {
    times: times,
    values: _.map(values, (value) => value + delta),
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

  const newSeries = _.map(values, (_value, i) => {
    const window = _.slice(values, _.max([0, i - windowSize + 1]), i + 1);
    return _.mean(window);
  });

  return {
    times: times,
    values: newSeries,
  };
}

// [Support Funcs] Transform wrapper

function transformWrapper(func: (...args: any) => { times: number[]; values: number[] }, ...args: any) {
  const funcArgs = args.slice(0, -1);
  const dataFrames: DataFrame[] = args[args.length - 1];

  const tsData = _.map(dataFrames, (dataFrame) => {
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

  return _.filter(dataFrames, (dataFrame) => {
    const valfield = dataFrame.fields[1];
    const displayName = getFieldDisplayName(valfield, dataFrame);
    return !regex.test(displayName);
  });
}

// [Support Funcs] Datapoints aggregation functions

function datapointsAvg(values: number[]) {
  return _.mean(values);
}

function datapointsMin(values: number[]) {
  return _.min(values);
}

function datapointsMax(values: number[]) {
  return _.max(values);
}

function datapointsSum(values: number[]) {
  return _.sum(values);
}

// Mirrors Scalars.Rank and cmp.Compare in the backend: an empty series ranks as
// 0, except by avg, where its NaN ranks below every number.
const rankFuncs = new Map<string, (values: number[]) => number>([
  ['avg', (values) => _.mean(values)],
  ['min', (values) => _.min(values) ?? 0],
  ['max', (values) => _.max(values) ?? 0],
  ['sum', (values) => _.sum(values)],
  ['absoluteMin', (values) => _.min(values.map(Math.abs)) ?? 0],
  ['absoluteMax', (values) => _.max(values.map(Math.abs)) ?? 0],
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

const functions = {
  // Transform
  scale: _.partial(transformWrapper, scale),
  offset: _.partial(transformWrapper, offset),
  delta: _.partial(transformWrapper, delta),
  fluctuation: _.partial(transformWrapper, fluctuation),
  movingAverage: _.partial(transformWrapper, movingAverage),
  // Filter Series
  top: _.partial(extraction, 'top'),
  bottom: _.partial(extraction, 'bottom'),
  exclude,
  // Sort
  sortByAvg: _.partial(sortByAggFuncs, 'avg'),
  sortByMax: _.partial(sortByAggFuncs, 'max'),
  sortByMin: _.partial(sortByAggFuncs, 'min'),
  sortBySum: _.partial(sortByAggFuncs, 'sum'),
  sortByAbsMax: _.partial(sortByAggFuncs, 'absoluteMax'),
  sortByAbsMin: _.partial(sortByAggFuncs, 'absoluteMin'),
};

const arrayFunctions: { [key: string]: { func: any; label: string } } = {
  toScalarByAvg: { func: datapointsAvg, label: 'avg' },
  toScalarByMax: { func: datapointsMax, label: 'max' },
  toScalarByMin: { func: datapointsMin, label: 'min' },
  toScalarBySum: { func: datapointsSum, label: 'sum' },
  toScalarByMed: { func: math.median, label: 'median' },
  toScalarByStd: { func: math.std, label: 'std' },
};

export { functions as seriesFunctions, arrayFunctions };
