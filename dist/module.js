'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.AnnotationsQueryCtrl = exports.QueryOptionsCtrl = exports.ConfigCtrl = exports.QueryCtrl = exports.Datasource = undefined;

var _datasource = require('./datasource');

var _query_ctrl = require('./query_ctrl');

var _config_ctrl = require('./config_ctrl');

var _annotation_ctrl = require('./annotation_ctrl');

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var ArchiverapplianceQueryOptionsCtrl = function ArchiverapplianceQueryOptionsCtrl() {
  _classCallCheck(this, ArchiverapplianceQueryOptionsCtrl);
};

ArchiverapplianceQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

exports.Datasource = _datasource.ArchiverapplianceDatasource;
exports.QueryCtrl = _query_ctrl.ArchiverapplianceDatasourceQueryCtrl;
exports.ConfigCtrl = _config_ctrl.ArchiverapplianceConfigCtrl;
exports.QueryOptionsCtrl = ArchiverapplianceQueryOptionsCtrl;
exports.AnnotationsQueryCtrl = _annotation_ctrl.ArchiverapplianceAnnotationsQueryCtrl;
//# sourceMappingURL=module.js.map
