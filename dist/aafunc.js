"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.createFuncInstance = createFuncInstance;
exports.getFuncDef = getFuncDef;
exports.getCategories = getCategories;

var _lodash = _interopRequireDefault(require("lodash"));

var _jquery = _interopRequireDefault(require("jquery"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

function ownKeys(object, enumerableOnly) { var keys = Object.keys(object); if (Object.getOwnPropertySymbols) { var symbols = Object.getOwnPropertySymbols(object); if (enumerableOnly) symbols = symbols.filter(function (sym) { return Object.getOwnPropertyDescriptor(object, sym).enumerable; }); keys.push.apply(keys, symbols); } return keys; }

function _objectSpread(target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i] != null ? arguments[i] : {}; if (i % 2) { ownKeys(source, true).forEach(function (key) { _defineProperty(target, key, source[key]); }); } else if (Object.getOwnPropertyDescriptors) { Object.defineProperties(target, Object.getOwnPropertyDescriptors(source)); } else { ownKeys(source).forEach(function (key) { Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key)); }); } } return target; }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

var funcIndex = [];
var categories = {
  Transform: [],
  'Filter Series': [],
  Options: []
};

function addFuncDef(newFuncDef) {
  var funcDef = _objectSpread({}, newFuncDef);

  funcDef.params = funcDef.params || [];
  funcDef.defaultParams = funcDef.defaultParams || [];

  if (funcDef.category) {
    categories[funcDef.category].push(funcDef);
  }

  funcIndex[funcDef.name] = funcDef;
  funcIndex[funcDef.shortName || funcDef.name] = funcDef;
} // Transform


addFuncDef({
  name: 'scale',
  category: 'Transform',
  params: [{
    name: 'factor',
    type: 'float',
    options: [100, 0.01, 10, -1]
  }],
  defaultParams: [100]
});
addFuncDef({
  name: 'offset',
  category: 'Transform',
  params: [{
    name: 'delta',
    type: 'float',
    options: [-100, 100]
  }],
  defaultParams: [100]
});
addFuncDef({
  name: 'delta',
  category: 'Transform',
  params: [],
  defaultParams: []
});
addFuncDef({
  name: 'fluctuation',
  category: 'Transform',
  params: [],
  defaultParams: []
}); // Filter Series

addFuncDef({
  name: 'top',
  category: 'Filter Series',
  params: [{
    name: 'number',
    type: 'int'
  }, {
    name: 'value',
    type: 'string',
    options: ['avg', 'min', 'max', 'absoluteMin', 'absoluteMax', 'sum']
  }],
  defaultParams: [5, 'avg']
});
addFuncDef({
  name: 'bottom',
  category: 'Filter Series',
  params: [{
    name: 'number',
    type: 'int'
  }, {
    name: 'value',
    type: 'string',
    options: ['avg', 'min', 'max', 'absoluteMin', 'absoluteMax', 'sum']
  }],
  defaultParams: [5, 'avg']
});
addFuncDef({
  name: 'exclude',
  category: 'Filter Series',
  params: [{
    name: 'pattern',
    type: 'string'
  }],
  defaultParams: []
}); // Options

addFuncDef({
  name: 'maxNumPVs',
  category: 'Options',
  params: [{
    name: 'number',
    type: 'int'
  }],
  defaultParams: [100]
});
addFuncDef({
  name: 'binInterval',
  category: 'Options',
  params: [{
    name: 'interval',
    type: 'int'
  }],
  defaultParams: [900]
});

var FuncInstance =
/*#__PURE__*/
function () {
  function FuncInstance(funcDef, params) {
    _classCallCheck(this, FuncInstance);

    this.def = funcDef;

    if (params) {
      this.params = params;
    } else {
      // Create with default params
      this.params = [];
      this.params = funcDef.defaultParams.slice(0);
    }

    this.updateText();
  }

  _createClass(FuncInstance, [{
    key: "bindFunction",
    value: function bindFunction(metricFunctions) {
      var func = metricFunctions[this.def.name];

      if (!func) {
        throw new Error({
          message: "Method not found ".concat(this.def.name)
        });
      } // Bind function arguments


      var bindedFunc = func;
      var param;

      for (var i = 0; i < this.params.length; i += 1) {
        param = this.params[i]; // Convert numeric params

        if (this.def.params[i].type === 'int' || this.def.params[i].type === 'float') {
          param = Number(param);
        }

        bindedFunc = _lodash.default.partial(bindedFunc, param);
      }

      return bindedFunc;
    }
  }, {
    key: "render",
    value: function render(metricExp) {
      var _this = this;

      var str = "".concat(this.def.name, "(");

      var parameters = _lodash.default.map(this.params, function (value, index) {
        var paramType = _this.def.params[index].type;

        if (paramType === 'int' || paramType === 'float' || paramType === 'value_or_series' || paramType === 'boolean') {
          return value;
        }

        if (paramType === 'int_or_interval' && _jquery.default.isNumeric(value)) {
          return value;
        }

        return "'".concat(value, "'");
      }, this);

      if (metricExp) {
        parameters.unshift(metricExp);
      }

      return "".concat(str).concat(parameters.join(', '), ")");
    }
  }, {
    key: "_hasMultipleParamsInString",
    value: function _hasMultipleParamsInString(strValue, index) {
      if (strValue.indexOf(',') === -1) {
        return false;
      }

      return this.def.params[index + 1] && this.def.params[index + 1].optional;
    }
  }, {
    key: "updateParam",
    value: function updateParam(strValue, index) {
      var _this2 = this;

      // handle optional parameters
      // if string contains ',' and next param is optional, split and update both
      if (this._hasMultipleParamsInString(strValue, index)) {
        _lodash.default.each(strValue.split(','), function (partVal, idx) {
          _this2.updateParam(partVal.trim(), idx);
        }, this);

        return;
      }

      if (strValue === '' && this.def.params[index].optional) {
        this.params.splice(index, 1);
      } else {
        this.params[index] = strValue;
      }

      this.updateText();
    }
  }, {
    key: "updateText",
    value: function updateText() {
      if (this.params.length === 0) {
        this.text = "".concat(this.def.name, "()");
        return;
      }

      var text = "".concat(this.def.name, "(").concat(this.params.join(', '), ")");
      this.text = text;
    }
  }]);

  return FuncInstance;
}();

function createFuncInstance(funcDef, params) {
  if (_lodash.default.isString(funcDef)) {
    if (!funcIndex[funcDef]) {
      throw new Error({
        message: "Method not found ".concat(funcDef.name)
      });
    }

    return new FuncInstance(funcIndex[funcDef], params);
  }

  return new FuncInstance(funcDef, params);
}

function getFuncDef(name) {
  return funcIndex[name];
}

function getCategories() {
  return categories;
}
//# sourceMappingURL=aafunc.js.map
