"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports["default"] = void 0;

var _lodash = _interopRequireDefault(require("lodash"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

var functions = {
  // Transform
  scale: scale,
  offset: offset,
  delta: delta,
  fluctuation: fluctuation,
  // Filter Series
  top: _lodash["default"].partial(extraction, 'top'),
  bottom: _lodash["default"].partial(extraction, 'bottom')
}; // Transform

function scale(factor, datapoints) {
  return _lodash["default"].map(datapoints, function (point) {
    return [point[0] * factor, point[1]];
  });
}

function offset(delta, datapoints) {
  for (var i = 0; i < datapoints.length; i++) {
    datapoints[i] = [datapoints[i][0] + delta, datapoints[i][1]];
  }

  return datapoints;
}

function delta(datapoints) {
  var newSeries = [];
  var deltaValue;

  for (var i = 1; i < datapoints.length; i++) {
    deltaValue = datapoints[i][0] - datapoints[i - 1][0];
    newSeries.push([deltaValue, datapoints[i][1]]);
  }

  return newSeries;
}

function fluctuation(datapoints) {
  var newSeries = [];
  var flucValue;

  for (var i = 0; i < datapoints.length; i++) {
    flucValue = datapoints[i][0] - datapoints[0][0];
    newSeries.push([flucValue, datapoints[i][1]]);
  }

  return newSeries;
} // Filter Series


function extraction(order, n, orderFunc, timeseriesData) {
  var orderByCallback = datapointsAggFuncs[orderFunc];

  var sortByIteratee = function sortByIteratee(ts) {
    return orderByCallback(ts.datapoints);
  };

  var sortedTsData = _lodash["default"].sortBy(timeseriesData, sortByIteratee);

  if (order === 'bottom') {
    return _lodash["default"].slice(sortedTsData, 0, n);
  } else {
    return _lodash["default"].reverse(_lodash["default"].slice(sortedTsData, -n));
  }
} // [Support Funcs] Datapoints aggregation functions


var datapointsAggFuncs = {
  avg: datapointsAvg,
  min: datapointsMin,
  max: datapointsMax,
  sum: datapointsSum,
  absoluteMin: datapointsAbsMin,
  absoluteMax: datapointsAbsMax
};

function datapointsAvg(datapoints) {
  return _lodash["default"].meanBy(datapoints, function (point) {
    return point[0];
  });
}

function datapointsMin(datapoints) {
  var minPoint = _lodash["default"].minBy(datapoints, function (point) {
    return point[0];
  });

  return minPoint[0];
}

function datapointsMax(datapoints) {
  var maxPoint = _lodash["default"].maxBy(datapoints, function (point) {
    return point[0];
  });

  return maxPoint[0];
}

function datapointsSum(datapoints) {
  return _lodash["default"].sumBy(datapoints, function (point) {
    return point[0];
  });
}

function datapointsAbsMin(datapoints) {
  var minPoint = _lodash["default"].minBy(datapoints, function (point) {
    return Math.abs(point[0]);
  });

  return Math.abs(minPoint[0]);
}

function datapointsAbsMax(datapoints) {
  var maxPoint = _lodash["default"].maxBy(datapoints, function (point) {
    return Math.abs(point[0]);
  });

  return Math.abs(maxPoint[0]);
}

var _default = {
  get aaFunctions() {
    return functions;
  }

};
exports["default"] = _default;
//# sourceMappingURL=dataProcessor.js.map
