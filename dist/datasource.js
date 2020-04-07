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

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } if (obj === null || _typeof(obj) !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _slicedToArray(arr, i) { return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _nonIterableRest(); }

function _nonIterableRest() { throw new TypeError("Invalid attempt to destructure non-iterable instance"); }

function _iterableToArrayLimit(arr, i) { if (!(Symbol.iterator in Object(arr) || Object.prototype.toString.call(arr) === "[object Arguments]")) { return; } var _arr = []; var _n = true; var _d = false; var _e = undefined; try { for (var _i = arr[Symbol.iterator](), _s; !(_n = (_s = _i.next()).done); _n = true) { _arr.push(_s.value); if (i && _arr.length === i) break; } } catch (err) { _d = true; _e = err; } finally { try { if (!_n && _i["return"] != null) _i["return"](); } finally { if (_d) throw _e; } } return _arr; }

function _arrayWithHoles(arr) { if (Array.isArray(arr)) return arr; }

function ownKeys(object, enumerableOnly) { var keys = Object.keys(object); if (Object.getOwnPropertySymbols) { var symbols = Object.getOwnPropertySymbols(object); if (enumerableOnly) symbols = symbols.filter(function (sym) { return Object.getOwnPropertyDescriptor(object, sym).enumerable; }); keys.push.apply(keys, symbols); } return keys; }

function _objectSpread(target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i] != null ? arguments[i] : {}; if (i % 2) { ownKeys(source, true).forEach(function (key) { _defineProperty(target, key, source[key]); }); } else if (Object.getOwnPropertyDescriptors) { Object.defineProperties(target, Object.getOwnPropertyDescriptors(source)); } else { ownKeys(source).forEach(function (key) { Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key)); }); } } return target; }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

var ArchiverapplianceDatasource =
/*#__PURE__*/
function () {
  function ArchiverapplianceDatasource(instanceSettings, backendSrv, templateSrv) {
    _classCallCheck(this, ArchiverapplianceDatasource);

    this.type = instanceSettings.type;
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.backendSrv = backendSrv;
    this.templateSrv = templateSrv;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = {
      'Content-Type': 'application/json'
    };

    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers.Authorization = instanceSettings.basicAuth;
    }

    this.operatorList = ['firstSample', 'lastSample', 'firstFill', 'lastFill', 'mean', 'min', 'max', 'count', 'ncount', 'nth', 'median', 'std', 'jitter', 'ignoreflyers', 'flyers', 'variance', 'popvariance', 'kurtosis', 'skewness', 'raw'];
  }

  _createClass(ArchiverapplianceDatasource, [{
    key: "query",
    value: function query(options) {
      var _this = this;

      var query = this.buildQueryParameters(options);
      query.targets = _lodash.default.filter(query.targets, function (t) {
        return !t.hide;
      });

      if (query.targets.length <= 0) {
        return Promise.resolve({
          data: []
        });
      }

      var targetProcesses = _lodash.default.map(query.targets, function (target) {
        return _this.targetProcess(target);
      });

      return Promise.all(targetProcesses).then(function (timeseriesDataArray) {
        return _this.postProcess(timeseriesDataArray);
      });
    }
  }, {
    key: "targetProcess",
    value: function targetProcess(target) {
      var _this2 = this;

      return this.buildUrls(target).then(function (urls) {
        return _this2.doMultiUrlRequests(urls);
      }).then(function (responses) {
        return _this2.responseParse(responses);
      }).then(function (timeseriesData) {
        return _this2.setAlias(timeseriesData, target);
      }).then(function (timeseriesData) {
        return _this2.applyFunctions(timeseriesData, target);
      });
    }
  }, {
    key: "postProcess",
    value: function postProcess(timeseriesDataArray) {
      var timeseriesData = _lodash.default.flatten(timeseriesDataArray);

      return {
        data: timeseriesData
      };
    }
  }, {
    key: "buildUrls",
    value: function buildUrls(target) {
      var _this3 = this;

      var targetQueries = this.parseTargetQuery(target.target);
      var maxNumPVs = 100;

      if (target.options.maxNumPVs) {
        maxNumPVs = target.options.maxNumPVs;
      }

      var pvnamesPromise = _lodash.default.map(targetQueries, function (targetQuery) {
        if (target.regex) {
          return _this3.pvNamesFindQuery(targetQuery, maxNumPVs);
        }

        return Promise.resolve([targetQuery]);
      });

      return Promise.all(pvnamesPromise).then(function (pvnamesArray) {
        return new Promise(function (resolve, reject) {
          var pvnames = _lodash.default.slice(_lodash.default.uniq(_lodash.default.flatten(pvnamesArray)), 0, maxNumPVs);

          var urls;

          try {
            urls = _lodash.default.map(pvnames, function (pvname) {
              return _this3.buildUrl(pvname, target.operator, target.interval, target.from, target.to);
            });
          } catch (e) {
            reject(e);
          }

          resolve(urls);
        });
      });
    }
  }, {
    key: "buildUrl",
    value: function buildUrl(pvname, operator, interval, from, to) {
      var pv = '';

      if (operator === 'raw' || interval === '') {
        pv = "".concat(pvname);
      } else if (_lodash.default.includes(['', undefined], operator)) {
        // Default Operator
        pv = "mean_".concat(interval, "(").concat(pvname, ")");
      } else if (_lodash.default.includes(this.operatorList, operator)) {
        pv = "".concat(operator, "_").concat(interval, "(").concat(pvname, ")");
      } else {
        throw new Error('Data Processing Operator is invalid.');
      }

      var url = "".concat(this.url, "/data/getData.json?pv=").concat(encodeURIComponent(pv), "&from=").concat(from.toISOString(), "&to=").concat(to.toISOString());
      return url;
    }
  }, {
    key: "doMultiUrlRequests",
    value: function doMultiUrlRequests(urls) {
      var _this4 = this;

      var requests = _lodash.default.map(urls, function (url) {
        return _this4.doRequest({
          url: url,
          method: 'GET'
        });
      });

      return Promise.all(requests);
    }
  }, {
    key: "responseParse",
    value: function responseParse(responses) {
      var timeSeriesDataArray = _lodash.default.map(responses, function (response) {
        var timeSeriesData = _lodash.default.map(response.data, function (targetRes) {
          var timesiries = _lodash.default.map(targetRes.data, function (datapoint) {
            return [datapoint.val, datapoint.secs * 1000 + _lodash.default.floor(datapoint.nanos / 1000000)];
          });

          var timeseries = {
            target: targetRes.meta.name,
            datapoints: timesiries
          };
          return timeseries;
        });

        return timeSeriesData;
      });

      return Promise.resolve(_lodash.default.flatten(timeSeriesDataArray));
    }
  }, {
    key: "setAlias",
    value: function setAlias(timeseriesData, target) {
      if (!target.alias) {
        return Promise.resolve(timeseriesData);
      }

      var pattern;

      if (target.aliasPattern) {
        pattern = new RegExp(target.aliasPattern, '');
      }

      var newTimeseriesData = _lodash.default.map(timeseriesData, function (timeseries) {
        if (pattern) {
          var alias = timeseries.target.replace(pattern, target.alias);
          return {
            target: alias,
            datapoints: timeseries.datapoints
          };
        }

        return {
          target: target.alias,
          datapoints: timeseries.datapoints
        };
      });

      return Promise.resolve(newTimeseriesData);
    }
  }, {
    key: "applyFunctions",
    value: function applyFunctions(timeseriesData, target) {
      if (target.functions === undefined) {
        return Promise.resolve(timeseriesData);
      }

      return this.bindFunctionDefs(target.functions, ['Transform', 'Filter Series'], timeseriesData);
    }
  }, {
    key: "testDatasource",
    value: function testDatasource() {
      return {
        status: 'success',
        message: 'Data source is working',
        title: 'Success'
      };
    }
  }, {
    key: "pvNamesFindQuery",
    value: function pvNamesFindQuery(query, maxPvs) {
      if (!query) {
        return Promise.resolve([]);
      }

      var url = "".concat(this.url, "/bpl/getMatchingPVs?limit=").concat(maxPvs, "&regex=").concat(encodeURIComponent(query));
      return this.doRequest({
        url: url,
        method: 'GET'
      }).then(function (res) {
        return res.data;
      });
    }
  }, {
    key: "metricFindQuery",
    value: function metricFindQuery(query) {
      var _this5 = this;

      var replacedQuery = this.templateSrv.replace(query, null, 'regex');
      var parsedQuery = this.parseTargetQuery(replacedQuery);

      var pvnamesPromise = _lodash.default.map(parsedQuery, function (targetQuery) {
        return _this5.pvNamesFindQuery(targetQuery, 100);
      });

      return Promise.all(pvnamesPromise).then(function (pvnamesArray) {
        var pvnames = _lodash.default.slice(_lodash.default.uniq(_lodash.default.flatten(pvnamesArray)), 0, 100);

        return _lodash.default.map(pvnames, function (pvname) {
          return {
            text: pvname
          };
        });
      });
    }
  }, {
    key: "doRequest",
    value: function doRequest(options) {
      var newOptions = _objectSpread({}, options);

      newOptions.withCredentials = this.withCredentials;
      newOptions.headers = this.headers;
      var result = this.backendSrv.datasourceRequest(newOptions);
      return result;
    }
  }, {
    key: "buildQueryParameters",
    value: function buildQueryParameters(options) {
      var _this6 = this;

      var query = _objectSpread({}, options); // remove placeholder targets and undefined targets


      query.targets = _lodash.default.filter(query.targets, function (target) {
        return target.target !== '' && typeof target.target !== 'undefined';
      });

      if (query.targets.length <= 0) {
        return query;
      }

      var from = new Date(query.range.from);
      var to = new Date(query.range.to);
      var rangeMsec = to.getTime() - from.getTime();

      var intervalSec = _lodash.default.floor(rangeMsec / (query.maxDataPoints * 1000));

      var interval = intervalSec >= 1 ? String(intervalSec) : '';

      var targets = _lodash.default.map(query.targets, function (target) {
        return {
          target: _this6.templateSrv.replace(target.target, query.scopedVars, 'regex'),
          refId: target.refId,
          hide: target.hide,
          alias: target.alias,
          operator: _this6.templateSrv.replace(target.operator, query.scopedVars, 'regex'),
          functions: target.functions,
          regex: target.regex,
          aliasPattern: target.aliasPattern,
          options: _this6.getOptions(target.functions),
          from: from,
          to: to,
          interval: interval
        };
      });

      query.targets = targets;
      return query;
    }
  }, {
    key: "parseTargetQuery",
    value: function parseTargetQuery(targetQuery) {
      /*
       * ex) targetQuery = ABC(1|2|3)EFG(5|6)
       *     then
       *     splitQueries = ['ABC','(1|2|3'), 'EFG', '(5|6)']
       *     queries = [
       *     ABC1EFG5, ABC1EFG6, ABC2EFG6,
       *     ABC2EFG6, ABC3EFG5, ABC3EFG6
       *     ]
       */
      var splitQueries = _lodash.default.split(targetQuery, /(\(.*?\))/);

      var queries = [''];

      _lodash.default.forEach(splitQueries, function (splitQuery, i) {
        // Fixed string like 'ABC'
        if (i % 2 === 0) {
          queries = _lodash.default.map(queries, function (query) {
            return "".concat(query).concat(splitQuery);
          });
          return;
        } // Regex OR string like '(1|2|3)'


        var orElems = _lodash.default.split(_lodash.default.trim(splitQuery, '()'), '|');

        var newQueries = _lodash.default.map(queries, function (query) {
          return _lodash.default.map(orElems, function (orElem) {
            return "".concat(query).concat(orElem);
          });
        });

        queries = _lodash.default.flatten(newQueries);
      });

      return queries;
    }
  }, {
    key: "bindFunctionDefs",
    value: function bindFunctionDefs(functionDefs, categories, data) {
      var allCategorisedFuncDefs = aafunc.getCategories();

      var requiredCategoryFuncNames = _lodash.default.reduce(categories, function (funcNames, category) {
        return _lodash.default.concat(funcNames, _lodash.default.map(allCategorisedFuncDefs[category], 'name'));
      }, []);

      var applyFuncDefs = _lodash.default.filter(functionDefs, function (func) {
        return _lodash.default.includes(requiredCategoryFuncNames, func.def.name);
      });

      var promises = _lodash.default.reduce(applyFuncDefs, function (prevPromise, func) {
        return prevPromise.then(function (res) {
          var funcInstance = aafunc.createFuncInstance(func.def, func.params);
          var bindedFunc = funcInstance.bindFunction(_dataProcessor.default.aaFunctions);
          return Promise.resolve(bindedFunc(res));
        });
      }, Promise.resolve(data));

      return promises;
    }
  }, {
    key: "getOptions",
    value: function getOptions(functionDefs) {
      var allCategorisedFuncDefs = aafunc.getCategories();

      var optionsFuncNames = _lodash.default.map(allCategorisedFuncDefs.Options, 'name');

      var applyFuncDefs = _lodash.default.filter(functionDefs, function (func) {
        return _lodash.default.includes(optionsFuncNames, func.def.name);
      });

      var options = _lodash.default.reduce(applyFuncDefs, function (optionMap, func) {
        var _func$params = _slicedToArray(func.params, 1);

        optionMap[func.def.name] = _func$params[0];
        return optionMap;
      }, {});

      return options;
    }
  }]);

  return ArchiverapplianceDatasource;
}();

exports.ArchiverapplianceDatasource = ArchiverapplianceDatasource;
//# sourceMappingURL=datasource.js.map
