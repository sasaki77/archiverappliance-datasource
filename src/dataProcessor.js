import _ from 'lodash';

// Transform

function scale(factor, datapoints) {
  return _.map(datapoints, (point) => (
    [point[0] * factor, point[1]]
  ));
}

function offset(delta, datapoints) {
  return _.map(datapoints, (point) => (
    [point[0] + delta, point[1]]
  ));
}

function delta(datapoints) {
  const newSeries = [];
  for (let i = 1; i < datapoints.length; i += 1) {
    const deltaValue = datapoints[i][0] - datapoints[i - 1][0];
    newSeries.push([deltaValue, datapoints[i][1]]);
  }
  return newSeries;
}

function fluctuation(datapoints) {
  const newSeries = [];
  for (let i = 0; i < datapoints.length; i += 1) {
    const flucValue = datapoints[i][0] - datapoints[0][0];
    newSeries.push([flucValue, datapoints[i][1]]);
  }
  return newSeries;
}

// [Support Funcs] Transform wrapper

function transformWrapper(func, ...args) {
  const funcArgs = args.slice(0, -1);
  const timeseriesData = args[args.length - 1];

  const tsData = _.map(timeseriesData, (timeseries) => {
    timeseries.datapoints = func(...funcArgs, timeseries.datapoints);
    return timeseries;
  });

  return Promise.resolve(tsData);
}

// Filter Series

// [Support Funcs] Datapoints aggregation functions

function datapointsAvg(datapoints) {
  return _.meanBy(datapoints, (point) => point[0]);
}

function datapointsMin(datapoints) {
  const minPoint = _.minBy(datapoints, (point) => point[0]);
  return minPoint[0];
}

function datapointsMax(datapoints) {
  const maxPoint = _.maxBy(datapoints, (point) => point[0]);
  return maxPoint[0];
}

function datapointsSum(datapoints) {
  return _.sumBy(datapoints, (point) => point[0]);
}

function datapointsAbsMin(datapoints) {
  const minPoint = _.minBy(datapoints, (point) => Math.abs(point[0]));
  return Math.abs(minPoint[0]);
}

function datapointsAbsMax(datapoints) {
  const maxPoint = _.maxBy(datapoints, (point) => Math.abs(point[0]));
  return Math.abs(maxPoint[0]);
}

const datapointsAggFuncs = {
  avg: datapointsAvg,
  min: datapointsMin,
  max: datapointsMax,
  sum: datapointsSum,
  absoluteMin: datapointsAbsMin,
  absoluteMax: datapointsAbsMax,
};

// [Support Funcs] Wrapper function for top and bottom function

function extraction(order, n, orderFunc, timeseriesData) {
  const orderByCallback = datapointsAggFuncs[orderFunc];
  const sortByIteratee = (ts) => (
    orderByCallback(ts.datapoints)
  );

  const sortedTsData = _.sortBy(timeseriesData, sortByIteratee);
  if (order === 'bottom') {
    return _.slice(sortedTsData, 0, n);
  }

  return _.reverse(_.slice(sortedTsData, -n));
}

// Function list

const functions = {
  // Transform
  scale: _.partial(transformWrapper, scale),
  offset: _.partial(transformWrapper, offset),
  delta: _.partial(transformWrapper, delta),
  fluctuation: _.partial(transformWrapper, fluctuation),
  // Filter Series
  top: _.partial(extraction, 'top'),
  bottom: _.partial(extraction, 'bottom'),
};

export default {
  get aaFunctions() {
    return functions;
  },
};
