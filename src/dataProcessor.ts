import _ from 'lodash';
import { MutableDataFrame, ArrayVector, getFieldDisplayName } from '@grafana/data';
import * as math from 'mathjs';

// Transform

function scale(factor: number, times: number[], values: number[]) {
  return {
    times: times,
    values: _.map(values, value => value * factor),
  };
}

function offset(delta: number, times: number[], values: number[]) {
  return {
    times: times,
    values: _.map(values, value => value + delta),
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
  if (values.length < windowSize) {
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
  const dataFrames: MutableDataFrame[] = args[args.length - 1];

  const tsData = _.map(dataFrames, dataFrame => {
    const timesField = dataFrame.fields[0];
    const valField = dataFrame.fields[1];
    const vals = func(...funcArgs, timesField.values.toArray(), valField.values.toArray());

    const newTimesField = {
      ...timesField,
      values: new ArrayVector(vals.times),
    };

    const newValfield = {
      ...valField,
      values: new ArrayVector(vals.values),
    };

    return new MutableDataFrame({
      ...dataFrame,
      fields: [newTimesField, newValfield],
    });
  });

  return tsData;
}

// Filter Series
function exclude(pattern: string, dataFrames: MutableDataFrame[]) {
  const regex = new RegExp(pattern);
  return _.filter(dataFrames, dataFrame => {
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

function datapointsAbsMin(values: number[]) {
  const minPoint = _.minBy(values, value => Math.abs(value));

  if (minPoint === undefined) {
    return minPoint;
  }
  return Math.abs(minPoint);
}

function datapointsAbsMax(values: number[]) {
  const maxPoint = _.maxBy(values, value => Math.abs(value));

  if (maxPoint === undefined) {
    return maxPoint;
  }
  return Math.abs(maxPoint);
}

const datapointsAggFuncs: { [key: string]: (values: number[]) => number | undefined } = {
  avg: datapointsAvg,
  min: datapointsMin,
  max: datapointsMax,
  sum: datapointsSum,
  absoluteMin: datapointsAbsMin,
  absoluteMax: datapointsAbsMax,
};

// [Support Funcs] Wrapper function for top and bottom function

function extraction(order: string, n: number, orderFunc: string, dataFrames: MutableDataFrame[]) {
  const orderByCallback = datapointsAggFuncs[orderFunc];
  const sortByIteratee = (dataFrame: MutableDataFrame) => orderByCallback(dataFrame.fields[1].values.toArray());

  const sortedTsData = _.sortBy(dataFrames, sortByIteratee);
  if (order === 'bottom') {
    return _.slice(sortedTsData, 0, n);
  }

  return _.reverse(_.slice(sortedTsData, -n));
}

// [Support Funcs] Wrapper function for sort by AggFuncs
function sortByAggFuncs(orderFunc: string, order: string, dataFrames: MutableDataFrame[]) {
  const orderByCallback = datapointsAggFuncs[orderFunc];
  const sortByIteratee = (dataFrame: MutableDataFrame) => orderByCallback(dataFrame.fields[1].values.toArray());

  const sortedTsData = _.sortBy(dataFrames, sortByIteratee);

  if (order === 'asc') {
    return sortedTsData;
  }

  return _.reverse(sortedTsData);
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
