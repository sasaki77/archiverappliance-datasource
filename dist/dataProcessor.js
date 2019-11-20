"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports["default"] = void 0;

var _lodash = _interopRequireDefault(require("lodash"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

var functions = {
  multiply: multiply
};

function multiply(n, data) {
  return _lodash["default"].map(data, function (d) {
    var datapoints = _lodash["default"].map(d.datapoints, function (point) {
      return [point[0] * n, point[1]];
    });

    return {
      "target": d.target,
      "datapoints": datapoints
    };
  });
}

var _default = {
  get aaFunctions() {
    return functions;
  }

};
exports["default"] = _default;
//# sourceMappingURL=dataProcessor.js.map
