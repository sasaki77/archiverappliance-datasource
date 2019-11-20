import {ArchiverapplianceDatasource} from './datasource';
import {ArchiverapplianceDatasourceQueryCtrl} from './query_ctrl';
import {ArchiverapplianceConfigCtrl} from './config_ctrl';
import {ArchiverapplianceAnnotationsQueryCtrl} from './annotation_ctrl';
import './aa-add-func';
import './func_editor';

class ArchiverapplianceQueryOptionsCtrl {}
ArchiverapplianceQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

export {
  ArchiverapplianceDatasource as Datasource,
  ArchiverapplianceDatasourceQueryCtrl as QueryCtrl,
  ArchiverapplianceConfigCtrl as ConfigCtrl,
  ArchiverapplianceQueryOptionsCtrl as QueryOptionsCtrl,
  ArchiverapplianceAnnotationsQueryCtrl as AnnotationsQueryCtrl
};
