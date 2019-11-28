"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.ArchiverapplianceDatasourceQueryCtrl = void 0;

var _sdk = require("app/plugins/sdk");

require("./css/query-editor.css!");

var aafunc = _interopRequireWildcard(require("./aafunc"));

function _getRequireWildcardCache() { if (typeof WeakMap !== "function") return null; var cache = new WeakMap(); _getRequireWildcardCache = function _getRequireWildcardCache() { return cache; }; return cache; }

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } if (obj === null || _typeof(obj) !== "object" && typeof obj !== "function") { return { "default": obj }; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj["default"] = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function _typeof(obj) { if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

function _possibleConstructorReturn(self, call) { if (call && (_typeof(call) === "object" || typeof call === "function")) { return call; } return _assertThisInitialized(self); }

function _getPrototypeOf(o) { _getPrototypeOf = Object.setPrototypeOf ? Object.getPrototypeOf : function _getPrototypeOf(o) { return o.__proto__ || Object.getPrototypeOf(o); }; return _getPrototypeOf(o); }

function _assertThisInitialized(self) { if (self === void 0) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function"); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, writable: true, configurable: true } }); if (superClass) _setPrototypeOf(subClass, superClass); }

function _setPrototypeOf(o, p) { _setPrototypeOf = Object.setPrototypeOf || function _setPrototypeOf(o, p) { o.__proto__ = p; return o; }; return _setPrototypeOf(o, p); }

var ArchiverapplianceDatasourceQueryCtrl =
/*#__PURE__*/
function (_QueryCtrl) {
  _inherits(ArchiverapplianceDatasourceQueryCtrl, _QueryCtrl);

  function ArchiverapplianceDatasourceQueryCtrl($scope, $injector) {
    var _this;

    _classCallCheck(this, ArchiverapplianceDatasourceQueryCtrl);

    _this = _possibleConstructorReturn(this, _getPrototypeOf(ArchiverapplianceDatasourceQueryCtrl).call(this, $scope, $injector));
    _this.scope = $scope;
    _this.target.type = _this.target.type || 'timeserie';
    _this.target.functions = _this.target.functions || [];
    _this.getPVNames = _.bind(_this.getPVNames_, _assertThisInitialized(_this));
    _this.getOperators = _.bind(_this.getOperators_, _assertThisInitialized(_this));
    return _this;
  }

  _createClass(ArchiverapplianceDatasourceQueryCtrl, [{
    key: "addFunction",
    value: function addFunction(funcDef) {
      var newFunc = aafunc.createFuncInstance(funcDef);
      newFunc.added = true;
      this.target.functions.push(newFunc);
      this.moveAliasFuncLast();

      if (newFunc.params.length && newFunc.added || newFunc.def.params.length === 0) {
        this.targetChanged();
      }
    }
  }, {
    key: "removeFunction",
    value: function removeFunction(func) {
      this.target.functions = _.without(this.target.functions, func);
      this.targetChanged();
    }
  }, {
    key: "moveFunction",
    value: function moveFunction(func, offset) {
      var index = this.target.functions.indexOf(func);

      _.move(this.target.functions, index, index + offset);

      this.targetChanged();
    }
  }, {
    key: "moveAliasFuncLast",
    value: function moveAliasFuncLast() {
      var aliasFunc = _.find(this.target.functions, function (func) {
        return func.def.category === 'Alias';
      });

      if (aliasFunc) {
        this.target.functions = _.without(this.target.functions, aliasFunc);
        this.target.functions.push(aliasFunc);
      }
    }
  }, {
    key: "getPVNames_",
    value: function getPVNames_(query, callback) {
      if (this.target.regex) {
        return [];
      }

      var str = ['.*', query, '.*'].join('');
      this.datasource.pvNamesFindQuery(str).then(function (res) {
        callback(res);
      });
    }
  }, {
    key: "getOperators_",
    value: function getOperators_(query) {
      return this.datasource.operatorList;
    }
  }, {
    key: "toggleEditorMode",
    value: function toggleEditorMode() {
      this.target.rawQuery = !this.target.rawQuery;
    }
  }, {
    key: "targetChanged",
    value: function targetChanged() {
      this.panelCtrl.refresh(); // Asks the panel to refresh data.
    }
  }, {
    key: "onKeyup",
    value: function onKeyup(e) {
      if (e.keyCode === 13) {
        e.target.blur();
      }
    }
  }]);

  return ArchiverapplianceDatasourceQueryCtrl;
}(_sdk.QueryCtrl);

exports.ArchiverapplianceDatasourceQueryCtrl = ArchiverapplianceDatasourceQueryCtrl;
ArchiverapplianceDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
//# sourceMappingURL=query_ctrl.js.map
