import _ from 'lodash';

let functions = {
  // Transform
  scale: scale,
  offset: offset,
  delta: delta,
};

function scale(factor, datapoints) {
  return _.map(datapoints, point => {
    return [ point[0] * factor, point[1] ];
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
  for (var i = 1; i < datapoints.length; i++) {
    deltaValue = datapoints[i][0] - datapoints[i - 1][0];
    newSeries.push([deltaValue, datapoints[i][1]]);
  }
  return newSeries;
}

export default {
  get aaFunctions() {
    return functions;
  }
};
