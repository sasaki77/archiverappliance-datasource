import {QueryCtrl} from 'app/plugins/sdk';
import './css/query-editor.css!'

export class ArchiverapplianceDatasourceQueryCtrl extends QueryCtrl {

  constructor($scope, $injector)  {
    super($scope, $injector);

    this.scope = $scope;
    this.target.type = this.target.type || 'timeserie';

    this.getOperators = _.bind(this.getOperators_, this);
  }

  getOptions(query, name) {
    //return this.datasource.metricFindQuery(name + '=' + query || '');
    return [];
  }

  getOperators_(query) {
    return this.datasource.operatorList;
  }

  toggleEditorMode() {
    this.target.rawQuery = !this.target.rawQuery;
  }

  onChangeInternal() {
    this.panelCtrl.refresh(); // Asks the panel to refresh data.
  }

  onKeyup(e) {
    if(e.keyCode === 13){
      e.target.blur();
    }
  }
}

ArchiverapplianceDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
