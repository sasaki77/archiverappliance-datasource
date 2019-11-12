'use strict';

System.register(['lodash'], function (_export, _context) {
  "use strict";

  var _, _createClass, ArchiverapplianceDatasource;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [function (_lodash) {
      _ = _lodash.default;
    }],
    execute: function () {
      _createClass = function () {
        function defineProperties(target, props) {
          for (var i = 0; i < props.length; i++) {
            var descriptor = props[i];
            descriptor.enumerable = descriptor.enumerable || false;
            descriptor.configurable = true;
            if ("value" in descriptor) descriptor.writable = true;
            Object.defineProperty(target, descriptor.key, descriptor);
          }
        }

        return function (Constructor, protoProps, staticProps) {
          if (protoProps) defineProperties(Constructor.prototype, protoProps);
          if (staticProps) defineProperties(Constructor, staticProps);
          return Constructor;
        };
      }();

      _export('ArchiverapplianceDatasource', ArchiverapplianceDatasource = function () {
        function ArchiverapplianceDatasource(instanceSettings, $q, backendSrv, templateSrv) {
          _classCallCheck(this, ArchiverapplianceDatasource);

          this.type = instanceSettings.type;
          this.url = instanceSettings.url;
          this.name = instanceSettings.name;
          this.q = $q;
          this.backendSrv = backendSrv;
          this.templateSrv = templateSrv;
          this.withCredentials = instanceSettings.withCredentials;
          this.headers = { 'Content-Type': 'application/json' };
          if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
            this.headers['Authorization'] = instanceSettings.basicAuth;
          }

          var jsonData = instanceSettings.jsonData || {};

          this.url_mgmt = instanceSettings.jsonData.url_mgmt;
          this.operatorList = ["firstSample", "lastSample", "firstFill", "lastFill", "mean", "min", "max", "count", "ncount", "nth", "median", "std", "jitter", "ignoreflyers", "flyers", "variance", "popvariance", "kurtosis", "skewness", "raw"];
        }

        _createClass(ArchiverapplianceDatasource, [{
          key: 'query',
          value: function query(options) {
            var _this = this;

            var query = this.buildQueryParameters(options);
            query.targets = query.targets.filter(function (t) {
              return !t.hide;
            });

            if (query.targets.length <= 0) {
              return this.q.when({ data: [] });
            }

            var targets = query.targets.map(function (target) {
              return _this.targetProcess(target, options);
            });

            return this.q.all(targets).then(function (data) {
              return _this.postProcess(data);
            });
          }
        }, {
          key: 'targetProcess',
          value: function targetProcess(target, options) {
            var _this2 = this;

            return this.buildUrl(target, options).then(function (url) {
              return _this2.doRequest({
                url: url,
                method: 'GET'
              });
            }).then(function (res) {
              return _this2.responseParse(res);
            }).then(function (data) {
              return _this2.setAlias(data, target);
            });
          }
        }, {
          key: 'postProcess',
          value: function postProcess(data) {
            var d = data.reduce(function (result, d) {
              result = result.concat(d);
              return result;
            }, []);

            return { data: d };
          }
        }, {
          key: 'buildUrl',
          value: function buildUrl(target, options) {
            var deferred = this.q.defer();

            var interval = "";
            if (options.intervalMs > 1000) {
              interval = String(options.intervalMs / 1000 - 1);
            }

            var pv = "";
            if (["raw", "", undefined].includes(target.operator) || interval === "") {
              pv = "pv=" + target.target;
            } else if (this.operatorList.includes(target.operator)) {
              pv = "pv=" + target.operator + "_" + interval + "(" + target.target + ")";
            } else {
              deferred.reject(Error("Data Processing Operator is invalid."));
            }

            var from = new Date(options.range.from);
            var to = new Date(options.range.to);
            var url = this.url + '/data/getData.json?' + pv + '&from=' + from.toISOString() + '&to=' + to.toISOString();

            deferred.resolve(url);
            return deferred.promise;
          }
        }, {
          key: 'responseParse',
          value: function responseParse(response) {
            var deferred = this.q.defer();

            var target_data = response.data.map(function (target_res) {
              var timesiries = target_res.data.map(function (datapoint) {
                return [datapoint.val, datapoint.secs * 1000 + Math.floor(datapoint.nanos / 1000000)];
              });
              var target_data = { "target": target_res.meta["name"], "datapoints": timesiries };
              return target_data;
            });

            deferred.resolve(target_data);
            return deferred.promise;
          }
        }, {
          key: 'setAlias',
          value: function setAlias(data, target) {
            var deferred = this.q.defer();

            data.forEach(function (d) {
              if (target.alias !== undefined && target.alias !== "") {
                d.target = target.alias;
              }
            });

            deferred.resolve(data);
            return deferred.promise;
          }
        }, {
          key: 'testDatasource',
          value: function testDatasource() {
            return { status: "success", message: "Data source is working", title: "Success" };
            //return this.doRequest({
            //  url: this.url_mgmt + '/bpl/getAppliancesInCluster',
            //  method: 'GET',
            //}).then(response => {
            //  if (response.status === 200) {
            //    return { status: "success", message: "Data source is working", title: "Success" };
            //  }
            //});
          }
        }, {
          key: 'metricFindQuery',
          value: function metricFindQuery(query) {
            var str = this.templateSrv.replace(query, null, 'regex');

            if (str) {
              var s = str.toString().split('=');
              var target = s[1] || '';
              var name = s[0] || '';
            } else {
              var target = '';
              var name = '';
            }

            var interpolated = {
              target: target
            };

            interpolated.name = name;

            return this.doRequest({
              url: this.url + '/search',
              data: interpolated,
              method: 'POST'
            }).then(this.mapToTextValue);
          }
        }, {
          key: 'mapToTextValue',
          value: function mapToTextValue(result) {
            return _.map(result.data, function (d, i) {
              if (d && d.text && d.value) {
                return { text: d.text, value: d.value };
              } else if (_.isObject(d)) {
                return { text: d, value: i };
              }
              return { text: d, value: d };
            });
          }
        }, {
          key: 'doRequest',
          value: function doRequest(options) {
            options.withCredentials = this.withCredentials;
            options.headers = this.headers;

            var result = this.backendSrv.datasourceRequest(options);
            return result;
          }
        }, {
          key: 'buildQueryParameters',
          value: function buildQueryParameters(options) {
            var _this3 = this;

            //remove placeholder targets and undefined targets
            options.targets = _.filter(options.targets, function (target) {
              return target.target !== '' && typeof target.target !== 'undefined';
            });

            var targets = _.map(options.targets, function (target) {
              return {
                target: _this3.templateSrv.replace(target.target, options.scopedVars, 'regex'),
                refId: target.refId,
                hide: target.hide,
                alias: target.alias,
                operator: target.operator
              };
            });

            options.targets = targets;

            return options;
          }
        }]);

        return ArchiverapplianceDatasource;
      }());

      _export('ArchiverapplianceDatasource', ArchiverapplianceDatasource);
    }
  };
});
//# sourceMappingURL=datasource.js.map
