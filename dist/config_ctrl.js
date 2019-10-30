"use strict";

System.register([], function (_export, _context) {
  "use strict";

  var ArchiverapplianceConfigCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [],
    execute: function () {
      _export("ArchiverapplianceConfigCtrl", ArchiverapplianceConfigCtrl = function ArchiverapplianceConfigCtrl($scope, $injector) {
        _classCallCheck(this, ArchiverapplianceConfigCtrl);

        this.scope = $scope;
        this.current.jsonData.entityLabel = this.current.jsonData.entityLabel || "entity";
      });

      _export("ArchiverapplianceConfigCtrl", ArchiverapplianceConfigCtrl);

      ArchiverapplianceConfigCtrl.templateUrl = 'partials/config.html';
    }
  };
});
//# sourceMappingURL=config_ctrl.js.map
