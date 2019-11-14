"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
Object.defineProperty(exports, "Datasource", {
  enumerable: true,
  get: function get() {
    return _datasource.ArchiverapplianceDatasource;
  }
});
Object.defineProperty(exports, "QueryCtrl", {
  enumerable: true,
  get: function get() {
    return _query_ctrl.ArchiverapplianceDatasourceQueryCtrl;
  }
});
Object.defineProperty(exports, "ConfigCtrl", {
  enumerable: true,
  get: function get() {
    return _config_ctrl.ArchiverapplianceConfigCtrl;
  }
});
Object.defineProperty(exports, "AnnotationsQueryCtrl", {
  enumerable: true,
  get: function get() {
    return _annotation_ctrl.ArchiverapplianceAnnotationsQueryCtrl;
  }
});
exports.QueryOptionsCtrl = void 0;

var _datasource = require("./datasource");

var _query_ctrl = require("./query_ctrl");

var _config_ctrl = require("./config_ctrl");

var _annotation_ctrl = require("./annotation_ctrl");

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var ArchiverapplianceQueryOptionsCtrl = function ArchiverapplianceQueryOptionsCtrl() {
  _classCallCheck(this, ArchiverapplianceQueryOptionsCtrl);
};

exports.QueryOptionsCtrl = ArchiverapplianceQueryOptionsCtrl;
ArchiverapplianceQueryOptionsCtrl.templateUrl = 'partials/query.options.html';
//# sourceMappingURL=module.js.map
