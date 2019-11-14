"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.ArchiverapplianceAnnotationsQueryCtrl = void 0;

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var ArchiverapplianceAnnotationsQueryCtrl = function ArchiverapplianceAnnotationsQueryCtrl($scope, $injector) {
  _classCallCheck(this, ArchiverapplianceAnnotationsQueryCtrl);

  this.scope = $scope;
  this.annotation.param_names = this.datasource.annParam_names;
  this.annotation.param_vals = this.annotation.param_vals || {};
};

exports.ArchiverapplianceAnnotationsQueryCtrl = ArchiverapplianceAnnotationsQueryCtrl;
ArchiverapplianceAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';
//# sourceMappingURL=annotation_ctrl.js.map
