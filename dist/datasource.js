"use strict";

function _typeof(obj) { if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.ArchiverapplianceDatasource = void 0;

var _lodash = _interopRequireDefault(require("lodash"));

var _dataProcessor = _interopRequireDefault(require("./dataProcessor"));

var aafunc = _interopRequireWildcard(require("./aafunc"));

function _getRequireWildcardCache() { if (typeof WeakMap !== "function") return null; var cache = new WeakMap(); _getRequireWildcardCache = function _getRequireWildcardCache() { return cache; }; return cache; }

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } if (obj === null || _typeof(obj) !== "object" && typeof obj !== "function") { return { "default": obj }; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj["default"] = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { "default": obj }; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

var ArchiverapplianceDatasource =
/*#__PURE__*/
function () {
  function ArchiverapplianceDatasource(instanceSettings, $q, backendSrv, templateSrv) {
    _classCallCheck(this, ArchiverapplianceDatasource);

    this.type = instanceSettings.type;
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.q = $q;
    this.backendSrv = backendSrv;
    this.templateSrv = templateSrv;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = {
      'Content-Type': 'application/json'
    };

    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers['Authorization'] = instanceSettings.basicAuth;
    }

    var jsonData = instanceSettings.jsonData || {};
    this.operatorList = ["firstSample", "lastSample", "firstFill", "lastFill", "mean", "min", "max", "count", "ncount", "nth", "median", "std", "jitter", "ignoreflyers", "flyers", "variance", "popvariance", "kurtosis", "skewness", "raw"];
  }

  _createClass(ArchiverapplianceDatasource, [{
    key: "query",
    value: function query(options) {
      var _this = this;

      var query = this.buildQueryParameters(options);
      query.targets = _lodash["default"].filter(query.targets, function (t) {
        return !t.hide;
      });

      if (query.targets.length <= 0) {
        return this.q.when({
          data: []
        });
      }

      var targets = _lodash["default"].map(query.targets, function (target) {
        return _this.targetProcess(target, options);
      });

      return this.q.all(targets).then(function (data) {
        return _this.postProcess(data);
      });
    }
  }, {
    key: "targetProcess",
    value: function targetProcess(target, options) {
      var _this2 = this;

      return this.buildUrls(target, options).then(function (urls) {
        return _this2.doMultiUrlRequests(urls);
      }).then(function (responses) {
        return _this2.responseParse(responses);
      }).then(function (data) {
        return _this2.setAlias(data, target);
      }).then(function (data) {
        return _this2.applyFunctions(data, target);
      });
    }
  }, {
    key: "postProcess",
    value: function postProcess(data) {
      var d = _lodash["default"].flatten(data);

      return {
        data: d
      };
    }
  }, {
    key: "buildUrls",
    value: function buildUrls(target) {
      var _this3 = this;

      var pvnames_promise;

      if (target.regex) {
        pvnames_promise = this.PVNamesFindQuery(target.target);
      } else {
        var pvnames = this.q.defer();
        pvnames.resolve([target.target]);
        pvnames_promise = pvnames.promise;
      }

      return pvnames_promise.then(function (pvnames) {
        var deferred = _this3.q.defer();

        var urls;

        try {
          urls = _lodash["default"].map(pvnames, function (pvname) {
            return _this3.buildUrl(pvname, target.operator, target.interval, target.from, target.to);
          });
        } catch (e) {
          deferred.reject(e);
        }

        deferred.resolve(urls);
        return deferred.promise;
      });
    }
  }, {
    key: "buildUrl",
    value: function buildUrl(pvname, operator, interval, from, to) {
      var pv = "";

      if (operator === "raw" || interval === "") {
        pv = "pv=" + pvname;
      } else if (_lodash["default"].includes(["", undefined], operator)) {
        // Default Operator
        pv = "pv=mean_" + interval + "(" + pvname + ")";
      } else if (_lodash["default"].includes(this.operatorList, operator)) {
        pv = "pv=" + operator + "_" + interval + "(" + pvname + ")";
      } else {
        throw new Error("Data Processing Operator is invalid.");
      }

      var url = this.url + '/data/getData.json?' + pv + '&from=' + from.toISOString() + '&to=' + to.toISOString();
      return url;
    }
  }, {
    key: "doMultiUrlRequests",
    value: function doMultiUrlRequests(urls) {
      var _this4 = this;

      var requests = _lodash["default"].map(urls, function (url) {
        return _this4.doRequest({
          url: url,
          method: "GET"
        });
      });

      return this.q.all(requests);
    }
  }, {
    key: "responseParse",
    value: function responseParse(responses) {
      var deferred = this.q.defer();

      var target_data = _lodash["default"].map(responses, function (response) {
        var td = _lodash["default"].map(response.data, function (target_res) {
          var timesiries = _lodash["default"].map(target_res.data, function (datapoint) {
            return [datapoint.val, datapoint.secs * 1000 + _lodash["default"].floor(datapoint.nanos / 1000000)];
          });

          var target_data = {
            "target": target_res.meta["name"],
            "datapoints": timesiries
          };
          return target_data;
        });

        return td;
      });

      deferred.resolve(_lodash["default"].flatten(target_data));
      return deferred.promise;
    }
  }, {
    key: "setAlias",
    value: function setAlias(data, target) {
      var deferred = this.q.defer();

      _lodash["default"].forEach(data, function (d) {
        if (target.alias !== undefined && target.alias !== "") {
          d.target = target.alias;
        }
      });

      deferred.resolve(data);
      return deferred.promise;
    }
  }, {
    key: "applyFunctions",
    value: function applyFunctions(data, target) {
      var deferred = this.q.defer();

      if (target.functions === undefined) {
        deferred.resolve(data);
        return deferred.promise;
      } // Apply transformation functions


      var transformFunctions = bindFunctionDefs(target.functions, 'Transform');
      data = _lodash["default"].map(data, function (timeseries) {
        timeseries.datapoints = sequence(transformFunctions)(timeseries.datapoints);
        return timeseries;
      });
      deferred.resolve(data);
      return deferred.promise;
    }
  }, {
    key: "testDatasource",
    value: function testDatasource() {
      return {
        status: "success",
        message: "Data source is working",
        title: "Success"
      }; //return this.doRequest({
      //  url: this.url_mgmt + '/bpl/getAppliancesInCluster',
      //  method: 'GET',
      //}).then(response => {
      //  if (response.status === 200) {
      //    return { status: "success", message: "Data source is working", title: "Success" };
      //  }
      //});
    }
  }, {
    key: "PVNamesFindQuery",
    value: function PVNamesFindQuery(query) {
      var str = this.templateSrv.replace(query, null, 'regex');

      if (!str) {
        var deferred = this.q.defer();
        deferred.resolve([]);
        return deferred.promise;
      }

      var url = this.url + "/bpl/getMatchingPVs?limit=100&regex=" + str;
      return this.doRequest({
        url: url,
        method: 'GET'
      }).then(function (res) {
        return res.data;
      });
    }
  }, {
    key: "doRequest",
    value: function doRequest(options) {
      options.withCredentials = this.withCredentials;
      options.headers = this.headers;
      var result = this.backendSrv.datasourceRequest(options);
      return result;
    }
  }, {
    key: "buildQueryParameters",
    value: function buildQueryParameters(options) {
      var _this5 = this;

      //remove placeholder targets and undefined targets
      options.targets = _lodash["default"].filter(options.targets, function (target) {
        return target.target !== '' && typeof target.target !== 'undefined';
      });

      if (options.targets.length <= 0) {
        return options;
      }

      var from = new Date(options.range.from);
      var to = new Date(options.range.to);
      var range_msec = to.getTime() - from.getTime();

      var interval_sec = _lodash["default"].floor(range_msec / (options.maxDataPoints * 1000));

      var interval = "";

      if (interval_sec >= 1) {
        interval = String(interval_sec);
      }

      var targets = _lodash["default"].map(options.targets, function (target) {
        return {
          target: _this5.templateSrv.replace(target.target, options.scopedVars, 'regex'),
          refId: target.refId,
          hide: target.hide,
          alias: target.alias,
          operator: target.operator,
          from: from,
          to: to,
          interval: interval,
          functions: target.functions,
          regex: target.regex
        };
      });

      options.targets = targets;
      return options;
    }
  }]);

  return ArchiverapplianceDatasource;
}();

exports.ArchiverapplianceDatasource = ArchiverapplianceDatasource;

function bindFunctionDefs(functionDefs, category) {
  var aggregationFunctions = _lodash["default"].map(aafunc.getCategories()[category], 'name');

  var aggFuncDefs = _lodash["default"].filter(functionDefs, function (func) {
    return _lodash["default"].includes(aggregationFunctions, func.def.name);
  });

  return _lodash["default"].map(aggFuncDefs, function (func) {
    var funcInstance = aafunc.createFuncInstance(func.def, func.params);
    return funcInstance.bindFunction(_dataProcessor["default"].aaFunctions);
  });
}

function sequence(funcsArray) {
  return function (result) {
    for (var i = 0; i < funcsArray.length; i++) {
      result = funcsArray[i].call(this, result);
    }

    return result;
  };
}
//# sourceMappingURL=datasource.js.map
