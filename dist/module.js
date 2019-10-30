'use strict';

System.register(['./datasource', './query_ctrl', './config_ctrl', './annotation_ctrl'], function (_export, _context) {
  "use strict";

  var ArchiverapplianceDatasource, ArchiverapplianceDatasourceQueryCtrl, ArchiverapplianceConfigCtrl, ArchiverapplianceAnnotationsQueryCtrl, ArchiverapplianceQueryOptionsCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [function (_datasource) {
      ArchiverapplianceDatasource = _datasource.ArchiverapplianceDatasource;
    }, function (_query_ctrl) {
      ArchiverapplianceDatasourceQueryCtrl = _query_ctrl.ArchiverapplianceDatasourceQueryCtrl;
    }, function (_config_ctrl) {
      ArchiverapplianceConfigCtrl = _config_ctrl.ArchiverapplianceConfigCtrl;
    }, function (_annotation_ctrl) {
      ArchiverapplianceAnnotationsQueryCtrl = _annotation_ctrl.ArchiverapplianceAnnotationsQueryCtrl;
    }],
    execute: function () {
      _export('QueryOptionsCtrl', ArchiverapplianceQueryOptionsCtrl = function ArchiverapplianceQueryOptionsCtrl() {
        _classCallCheck(this, ArchiverapplianceQueryOptionsCtrl);
      });

      ArchiverapplianceQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

      _export('Datasource', ArchiverapplianceDatasource);

      _export('QueryCtrl', ArchiverapplianceDatasourceQueryCtrl);

      _export('ConfigCtrl', ArchiverapplianceConfigCtrl);

      _export('QueryOptionsCtrl', ArchiverapplianceQueryOptionsCtrl);

      _export('AnnotationsQueryCtrl', ArchiverapplianceAnnotationsQueryCtrl);
    }
  };
});
//# sourceMappingURL=module.js.map
