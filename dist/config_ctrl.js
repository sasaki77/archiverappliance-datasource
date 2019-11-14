"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.ArchiverapplianceConfigCtrl = void 0;

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var ArchiverapplianceConfigCtrl = function ArchiverapplianceConfigCtrl($scope, $injector) {
  _classCallCheck(this, ArchiverapplianceConfigCtrl);

  this.scope = $scope;
  this.current.jsonData.entityLabel = this.current.jsonData.entityLabel || "entity";
};

exports.ArchiverapplianceConfigCtrl = ArchiverapplianceConfigCtrl;
ArchiverapplianceConfigCtrl.templateUrl = 'partials/config.html';
//# sourceMappingURL=config_ctrl.js.map
