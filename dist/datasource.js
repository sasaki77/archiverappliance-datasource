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
        }

        _createClass(ArchiverapplianceDatasource, [{
          key: 'query',
          value: function query(options) {
            var query = this.buildQueryParameters(options);
            query.targets = query.targets.filter(function (t) {
              return !t.hide;
            });

            if (query.targets.length <= 0) {
              return this.q.when({ data: [] });
            }

            var pvs = new Array();
            query.targets.forEach(function (target) {
              pvs.push("pv=" + target.target);
            });

            var from = new Date(options.range.from);
            var to = new Date(options.range.to);

            return this.doRequest({
              url: this.url + '/data/getDataForPVs.json?' + pvs.join('&') + '&from=' + from.toISOString() + '&to=' + to.toISOString(),
              data: query,
              method: 'GET'
            }).then(this.responseParse);
          }
        }, {
          key: 'responseParse',
          value: function responseParse(response) {
            var data = new Array();
            response.data.forEach(function (td) {
              var timesiries = new Array();
              td.data.forEach(function (d) {
                timesiries.push([d.val, d.secs * 1000 + Math.floor(d.nanos / 1000000)]);
              });
              var d = { "target": td.meta["name"], "datapoints": timesiries };
              data.push(d);
            });
            return { data: data };
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
            var _this = this;

            //remove placeholder targets and undefined targets
            options.targets = _.filter(options.targets, function (target) {
              return target.target !== '' && typeof target.target !== 'undefined';
            });

            var targets = _.map(options.targets, function (target) {
              return {
                target: _this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
                refId: target.refId,
                hide: target.hide
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
