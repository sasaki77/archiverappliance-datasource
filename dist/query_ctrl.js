'use strict';

System.register(['app/plugins/sdk', './css/query-editor.css!'], function (_export, _context) {
  "use strict";

  var QueryCtrl, _createClass, ArchiverapplianceDatasourceQueryCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  function _possibleConstructorReturn(self, call) {
    if (!self) {
      throw new ReferenceError("this hasn't been initialised - super() hasn't been called");
    }

    return call && (typeof call === "object" || typeof call === "function") ? call : self;
  }

  function _inherits(subClass, superClass) {
    if (typeof superClass !== "function" && superClass !== null) {
      throw new TypeError("Super expression must either be null or a function, not " + typeof superClass);
    }

    subClass.prototype = Object.create(superClass && superClass.prototype, {
      constructor: {
        value: subClass,
        enumerable: false,
        writable: true,
        configurable: true
      }
    });
    if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass;
  }

  return {
    setters: [function (_appPluginsSdk) {
      QueryCtrl = _appPluginsSdk.QueryCtrl;
    }, function (_cssQueryEditorCss) {}],
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

      _export('ArchiverapplianceDatasourceQueryCtrl', ArchiverapplianceDatasourceQueryCtrl = function (_QueryCtrl) {
        _inherits(ArchiverapplianceDatasourceQueryCtrl, _QueryCtrl);

        function ArchiverapplianceDatasourceQueryCtrl($scope, $injector) {
          _classCallCheck(this, ArchiverapplianceDatasourceQueryCtrl);

          var _this = _possibleConstructorReturn(this, (ArchiverapplianceDatasourceQueryCtrl.__proto__ || Object.getPrototypeOf(ArchiverapplianceDatasourceQueryCtrl)).call(this, $scope, $injector));

          _this.scope = $scope;
          _this.target.type = _this.target.type || 'timeserie';

          return _this;
        }

        _createClass(ArchiverapplianceDatasourceQueryCtrl, [{
          key: 'getOptions',
          value: function getOptions(query, name) {
            //return this.datasource.metricFindQuery(name + '=' + query || '');
            return [];
          }
        }, {
          key: 'toggleEditorMode',
          value: function toggleEditorMode() {
            this.target.rawQuery = !this.target.rawQuery;
          }
        }, {
          key: 'onChangeInternal',
          value: function onChangeInternal() {
            this.panelCtrl.refresh(); // Asks the panel to refresh data.
          }
        }, {
          key: 'onKeyup',
          value: function onKeyup(e) {
            if (e.keyCode === 13) {
              e.target.blur();
            }
          }
        }]);

        return ArchiverapplianceDatasourceQueryCtrl;
      }(QueryCtrl));

      _export('ArchiverapplianceDatasourceQueryCtrl', ArchiverapplianceDatasourceQueryCtrl);

      ArchiverapplianceDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
    }
  };
});
//# sourceMappingURL=query_ctrl.js.map
