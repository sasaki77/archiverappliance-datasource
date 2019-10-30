export class ArchiverapplianceConfigCtrl {

  constructor($scope, $injector)  {
    this.scope = $scope;
    this.current.jsonData.entityLabel = this.current.jsonData.entityLabel || "entity";
  }
}

ArchiverapplianceConfigCtrl.templateUrl = 'partials/config.html';
