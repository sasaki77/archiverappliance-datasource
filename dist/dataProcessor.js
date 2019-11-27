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
  fluctuation: fluctuation
};

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
}

var _default = {
  get aaFunctions() {
    return functions;
  }

};
exports["default"] = _default;
//# sourceMappingURL=dataProcessor.js.map
