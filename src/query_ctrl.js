import {QueryCtrl} from 'app/plugins/sdk';
import './css/query-editor.css!'
import * as aafunc from './aafunc';

export class ArchiverapplianceDatasourceQueryCtrl extends QueryCtrl {

  constructor($scope, $injector)  {
    super($scope, $injector);

    this.scope = $scope;
    this.target.type = this.target.type || 'timeserie';
    this.target.functions = this.target.functions || [];

    this.getPVNames = _.bind(this.getPVNames_, this);
    this.getOperators = _.bind(this.getOperators_, this);
  }

  addFunction(funcDef) {
    var newFunc = aafunc.createFuncInstance(funcDef);
    newFunc.added = true;
    this.target.functions.push(newFunc);

    this.moveAliasFuncLast();

    if (newFunc.params.length && newFunc.added ||
        newFunc.def.params.length === 0) {
      this.targetChanged();
    }
  }

  removeFunction(func) {
    this.target.functions = _.without(this.target.functions, func);
    this.targetChanged();
  }

  moveFunction(func, offset) {
    const index = this.target.functions.indexOf(func);
    _.move(this.target.functions, index, index + offset);
    this.targetChanged();
  }

  moveAliasFuncLast() {
    var aliasFunc = _.find(this.target.functions, func => {
      return func.def.category === 'Alias';
    });

    if (aliasFunc) {
      this.target.functions = _.without(this.target.functions, aliasFunc);
      this.target.functions.push(aliasFunc);
    }
  }

  getPVNames_(query, callback) {
    const str = "*" + query + "*";
    this.datasource.PVNamesFindQuery(str).then( res => {
        callback(res);
    });
  }

  getOperators_(query) {
    return this.datasource.operatorList;
  }

  toggleEditorMode() {
    this.target.rawQuery = !this.target.rawQuery;
  }

  targetChanged() {
    this.panelCtrl.refresh(); // Asks the panel to refresh data.
  }

  onKeyup(e) {
    if(e.keyCode === 13){
      e.target.blur();
    }
  }
}

ArchiverapplianceDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
