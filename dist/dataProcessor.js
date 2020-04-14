"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

var _lodash = _interopRequireDefault(require("lodash"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _toConsumableArray(arr) { return _arrayWithoutHoles(arr) || _iterableToArray(arr) || _nonIterableSpread(); }

function _nonIterableSpread() { throw new TypeError("Invalid attempt to spread non-iterable instance"); }

function _iterableToArray(iter) { if (Symbol.iterator in Object(iter) || Object.prototype.toString.call(iter) === "[object Arguments]") return Array.from(iter); }

function _arrayWithoutHoles(arr) { if (Array.isArray(arr)) { for (var i = 0, arr2 = new Array(arr.length); i < arr.length; i++) { arr2[i] = arr[i]; } return arr2; } }

// Transform
function scale(factor, datapoints) {
  return _lodash.default.map(datapoints, function (point) {
    return [point[0] * factor, point[1]];
  });
}

function offset(delta, datapoints) {
  return _lodash.default.map(datapoints, function (point) {
    return [point[0] + delta, point[1]];
  });
}

function delta(datapoints) {
  var newSeries = [];

  for (var i = 1; i < datapoints.length; i += 1) {
    var deltaValue = datapoints[i][0] - datapoints[i - 1][0];
    newSeries.push([deltaValue, datapoints[i][1]]);
  }

  return newSeries;
}

function fluctuation(datapoints) {
  var newSeries = [];

  for (var i = 0; i < datapoints.length; i += 1) {
    var flucValue = datapoints[i][0] - datapoints[0][0];
    newSeries.push([flucValue, datapoints[i][1]]);
  }

  return newSeries;
} // [Support Funcs] Transform wrapper


function transformWrapper(func) {
  for (var _len = arguments.length, args = new Array(_len > 1 ? _len - 1 : 0), _key = 1; _key < _len; _key++) {
    args[_key - 1] = arguments[_key];
  }

  var funcArgs = args.slice(0, -1);
  var timeseriesData = args[args.length - 1];

  var tsData = _lodash.default.map(timeseriesData, function (timeseries) {
    timeseries.datapoints = func.apply(void 0, _toConsumableArray(funcArgs).concat([timeseries.datapoints]));
    return timeseries;
  });

  return tsData;
} // Filter Series
// [Support Funcs] Datapoints aggregation functions


function datapointsAvg(datapoints) {
  return _lodash.default.meanBy(datapoints, function (point) {
    return point[0];
  });
}

function datapointsMin(datapoints) {
  var minPoint = _lodash.default.minBy(datapoints, function (point) {
    return point[0];
  });

  return minPoint[0];
}

function datapointsMax(datapoints) {
  var maxPoint = _lodash.default.maxBy(datapoints, function (point) {
    return point[0];
  });

  return maxPoint[0];
}

function datapointsSum(datapoints) {
  return _lodash.default.sumBy(datapoints, function (point) {
    return point[0];
  });
}

function datapointsAbsMin(datapoints) {
  var minPoint = _lodash.default.minBy(datapoints, function (point) {
    return Math.abs(point[0]);
  });

  return Math.abs(minPoint[0]);
}

function datapointsAbsMax(datapoints) {
  var maxPoint = _lodash.default.maxBy(datapoints, function (point) {
    return Math.abs(point[0]);
  });

  return Math.abs(maxPoint[0]);
}

var datapointsAggFuncs = {
  avg: datapointsAvg,
  min: datapointsMin,
  max: datapointsMax,
  sum: datapointsSum,
  absoluteMin: datapointsAbsMin,
  absoluteMax: datapointsAbsMax
}; // [Support Funcs] Wrapper function for top and bottom function

function extraction(order, n, orderFunc, timeseriesData) {
  var orderByCallback = datapointsAggFuncs[orderFunc];

  var sortByIteratee = function sortByIteratee(ts) {
    return orderByCallback(ts.datapoints);
  };

  var sortedTsData = _lodash.default.sortBy(timeseriesData, sortByIteratee);

  if (order === 'bottom') {
    return _lodash.default.slice(sortedTsData, 0, n);
  }

  return _lodash.default.reverse(_lodash.default.slice(sortedTsData, -n));
} // Function list


var functions = {
  // Transform
  scale: _lodash.default.partial(transformWrapper, scale),
  offset: _lodash.default.partial(transformWrapper, offset),
  delta: _lodash.default.partial(transformWrapper, delta),
  fluctuation: _lodash.default.partial(transformWrapper, fluctuation),
  // Filter Series
  top: _lodash.default.partial(extraction, 'top'),
  bottom: _lodash.default.partial(extraction, 'bottom')
};
var _default = {
  get aaFunctions() {
    return functions;
  }

};
exports.default = _default;
//# sourceMappingURL=dataProcessor.js.map
