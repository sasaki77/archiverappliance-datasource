'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var ArchiverapplianceAnnotationsQueryCtrl = exports.ArchiverapplianceAnnotationsQueryCtrl = function ArchiverapplianceAnnotationsQueryCtrl($scope, $injector) {
  _classCallCheck(this, ArchiverapplianceAnnotationsQueryCtrl);

  this.scope = $scope;
  this.annotation.param_names = this.datasource.annParam_names;
  this.annotation.param_vals = this.annotation.param_vals || {};
};

ArchiverapplianceAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';
//# sourceMappingURL=annotation_ctrl.js.map
