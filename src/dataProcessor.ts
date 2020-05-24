import _ from 'lodash';

// Transform

function scale(factor: number, datapoints: any[]) {
  return _.map(datapoints, point => [point[0] * factor, point[1]]);
}

function offset(delta: number, datapoints: any[]) {
  return _.map(datapoints, point => [point[0] + delta, point[1]]);
}

function delta(datapoints: any[]) {
  const newSeries = [];
  for (let i = 1; i < datapoints.length; i += 1) {
    const deltaValue = datapoints[i][0] - datapoints[i - 1][0];
    newSeries.push([deltaValue, datapoints[i][1]]);
  }
  return newSeries;
}

function fluctuation(datapoints: any[]) {
  const newSeries = [];
  for (let i = 0; i < datapoints.length; i += 1) {
    const flucValue = datapoints[i][0] - datapoints[0][0];
    newSeries.push([flucValue, datapoints[i][1]]);
  }
  return newSeries;
}

// [Support Funcs] Transform wrapper

function transformWrapper(func: any, ...args: any) {
  const funcArgs = args.slice(0, -1);
  const timeseriesData = args[args.length - 1];

  const tsData = _.map(timeseriesData, timeseries => {
    timeseries.datapoints = func(...funcArgs, timeseries.datapoints);
    return timeseries;
  });

  return tsData;
}

// Filter Series
function exclude(pattern: string, timeseriesData: any[]) {
  const regex = new RegExp(pattern);
  return _.filter(timeseriesData, timeseries => !regex.test(timeseries.target));
}

// [Support Funcs] Datapoints aggregation functions

function datapointsAvg(datapoints: any[]) {
  return _.meanBy(datapoints, point => point[0]);
}

function datapointsMin(datapoints: any[]) {
  const minPoint = _.minBy(datapoints, point => point[0]);
  return minPoint[0];
}

function datapointsMax(datapoints: any[]) {
  const maxPoint = _.maxBy(datapoints, point => point[0]);
  return maxPoint[0];
}

function datapointsSum(datapoints: any[]) {
  return _.sumBy(datapoints, point => point[0]);
}

function datapointsAbsMin(datapoints: any[]) {
  const minPoint = _.minBy(datapoints, point => Math.abs(point[0]));
  return Math.abs(minPoint[0]);
}

function datapointsAbsMax(datapoints: any[]) {
  const maxPoint = _.maxBy(datapoints, point => Math.abs(point[0]));
  return Math.abs(maxPoint[0]);
}

const datapointsAggFuncs: { [key: string]: any } = {
  avg: datapointsAvg,
  min: datapointsMin,
  max: datapointsMax,
  sum: datapointsSum,
  absoluteMin: datapointsAbsMin,
  absoluteMax: datapointsAbsMax,
};

// [Support Funcs] Wrapper function for top and bottom function

function extraction(order: string, n: number, orderFunc: string, timeseriesData: any[]) {
  const orderByCallback = datapointsAggFuncs[orderFunc];
  const sortByIteratee = (ts: any) => orderByCallback(ts.datapoints);

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
  exclude,
};

export default {
  get aaFunctions() {
    return functions;
  },
};
