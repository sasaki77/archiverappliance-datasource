'use strict';

System.register([], function (_export, _context) {
  "use strict";

  var ArchiverapplianceAnnotationsQueryCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [],
    execute: function () {
      _export('ArchiverapplianceAnnotationsQueryCtrl', ArchiverapplianceAnnotationsQueryCtrl = function ArchiverapplianceAnnotationsQueryCtrl($scope, $injector) {
        _classCallCheck(this, ArchiverapplianceAnnotationsQueryCtrl);

        this.scope = $scope;
        this.annotation.param_names = this.datasource.annParam_names;
        this.annotation.param_vals = this.annotation.param_vals || {};
      });

      _export('ArchiverapplianceAnnotationsQueryCtrl', ArchiverapplianceAnnotationsQueryCtrl);

      ArchiverapplianceAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';
    }
  };
});
//# sourceMappingURL=annotation_ctrl.js.map
