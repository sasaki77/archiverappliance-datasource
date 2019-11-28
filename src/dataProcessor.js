import _ from 'lodash';

const functions = {
  // Transform
  scale: scale,
  offset: offset,
  delta: delta,
  fluctuation: fluctuation
};

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

export default {
  get aaFunctions() {
    return functions;
  }
};
