import _ from 'lodash';

const functions = {
  // Transform
  scale: scale,
  offset: offset,
  delta: delta,
  fluctuation: fluctuation,
  // Filter Series
  top: _.partial(extraction, 'top'),
  bottom: _.partial(extraction, 'bottom')
};

// Transform

function scale(factor, datapoints) {
  return _.map(datapoints, (point) => {
    return [point[0] * factor, point[1]];
  });
}

function offset(delta, datapoints) {
  for (let i = 0; i < datapoints.length; i++) {
    datapoints[i] = [
      datapoints[i][0] + delta,
      datapoints[i][1]
    ];
  }

  return datapoints;
}

function delta(datapoints) {
  let newSeries = [];
  let deltaValue;
  for (let i = 1; i < datapoints.length; i++) {
    deltaValue = datapoints[i][0] - datapoints[i - 1][0];
    newSeries.push([deltaValue, datapoints[i][1]]);
  }
  return newSeries;
}

function fluctuation(datapoints) {
  let newSeries = [];
  let flucValue;
  for (let i = 0; i < datapoints.length; i++) {
    flucValue = datapoints[i][0] - datapoints[0][0];
    newSeries.push([flucValue, datapoints[i][1]]);
  }
  return newSeries;
}

// Filter Series

function extraction(order, n, orderFunc, timeseriesData) {
  const orderByCallback = datapointsAggFuncs[orderFunc];
  const sortByIteratee = (ts) => {
      return orderByCallback(ts.datapoints);
  };

  const sortedTsData = _.sortBy(timeseriesData, sortByIteratee);
  if (order === 'bottom') {
    return _.slice(sortedTsData, 0, n);
  } else {
    return _.reverse(_.slice(sortedTsData, -n));
  }
}

// [Support Funcs] Datapoints aggregation functions

const datapointsAggFuncs = {
  avg: datapointsAvg,
  min: datapointsMin,
  max: datapointsMax,
  sum: datapointsSum
};

function datapointsAvg(datapoints) {
  return _.meanBy(datapoints, (point) => {
    return point[0];
  });
}

function datapointsMin(datapoints) {
  const minPoint = _.minBy(datapoints, (point) => {
    return point[0];
  });
  return minPoint[0];
}

function datapointsMax(datapoints) {
  const maxPoint = _.maxBy(datapoints, (point) => {
    return point[0];
  });
  return maxPoint[0]
}

function datapointsSum(datapoints) {
  return _.sumBy(datapoints, (point) => {
    return point[0];
  });
}

export default {
  get aaFunctions() {
    return functions;
  }
};
