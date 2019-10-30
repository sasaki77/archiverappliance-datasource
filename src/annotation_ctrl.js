export class ArchiverapplianceAnnotationsQueryCtrl {

  constructor($scope, $injector)  {
    this.scope = $scope;
    this.annotation.param_names = this.datasource.annParam_names;
    this.annotation.param_vals = this.annotation.param_vals || {};
  }

}

ArchiverapplianceAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html'
