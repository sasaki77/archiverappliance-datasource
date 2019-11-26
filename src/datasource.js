import _ from "lodash";
import dataProcessor from "./dataProcessor";
import * as aafunc from './aafunc';

export class ArchiverapplianceDatasource {

  constructor(instanceSettings, $q, backendSrv, templateSrv) {
    this.type = instanceSettings.type;
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.q = $q;
    this.backendSrv = backendSrv;
    this.templateSrv = templateSrv;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = {'Content-Type': 'application/json'};
    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers['Authorization'] = instanceSettings.basicAuth;
    }

    const jsonData = instanceSettings.jsonData || {};

    this.url_mgmt = instanceSettings.jsonData.url_mgmt;
    this.operatorList = ["firstSample", "lastSample", "firstFill", "lastFill", "mean", "min", "max",
        "count", "ncount", "nth", "median", "std", "jitter", "ignoreflyers", "flyers", "variance",
        "popvariance", "kurtosis", "skewness", "raw"];
  }

  query(options) {
    var query = this.buildQueryParameters(options);
    query.targets = _.filter(query.targets, t => !t.hide);

    if (query.targets.length <= 0) {
      return this.q.when({data: []});
    }

    const targets = _.map( query.targets, target => {
      return this.targetProcess(target, options);
    });

   return this.q.all(targets).then( data => this.postProcess(data) );
  }

  targetProcess(target, options) {
      return (
          this.buildUrl(target, options)
          .then( url => this.doRequest({ url: url, method: 'GET' }) )
          .then( res => this.responseParse(res) )
          .then( data => this.setAlias(data, target) )
          .then( data => this.applyFunctions(data, target) )
      );
  }

  postProcess(data) {
    const d = _.flatten( data );

    return {data: d};
  }

  buildUrl(target) {
    let deferred = this.q.defer();

    let pv = ""
    if ( target.operator === "raw" || target.interval === "") {
      pv = "pv=" + target.target;
    } else if ( _.includes(["", undefined], target.operator) ) {
      // Default Operator
      pv = "pv=mean_" + target.interval + "(" + target.target + ")";
    } else if ( _.includes(this.operatorList, target.operator) ) {
      pv = "pv=" + target.operator + "_" + target.interval + "(" + target.target + ")";
    } else {
      deferred.reject(Error("Data Processing Operator is invalid."));
    }

    const url = this.url + '/data/getData.json?' + pv + '&from=' + target.from.toISOString() + '&to=' + target.to.toISOString();

    deferred.resolve(url);
    return deferred.promise;
  }

  responseParse(response) {
    let deferred = this.q.defer();

    const target_data = _.map( response.data, target_res => {
      const timesiries = _.map( target_res.data, datapoint => {
          return [datapoint.val, datapoint.secs*1000+_.floor(datapoint.nanos/1000000)];
      });
      const target_data = {"target": target_res.meta["name"], "datapoints": timesiries};
      return target_data;
    });

    deferred.resolve(target_data);
    return deferred.promise;
  }

  setAlias(data, target) {
    let deferred = this.q.defer();

    _.forEach( data, d => {
      if( target.alias !== undefined && target.alias !== "" ) {
        d.target = target.alias;
      }
    });

    deferred.resolve(data);
    return deferred.promise;
  }

  applyFunctions(data, target) {
    let deferred = this.q.defer();

    if(target.functions === undefined){
      deferred.resolve(data);
      return deferred.promise;
    }

    // Apply transformation functions
    const transformFunctions = bindFunctionDefs(target.functions, 'Transform');
    data = _.map(data, timeseries => {
      timeseries.datapoints = sequence(transformFunctions)(timeseries.datapoints);
      return timeseries;
    });

    deferred.resolve(data);
    return deferred.promise;
  }

  testDatasource() {
    return { status: "success", message: "Data source is working", title: "Success" };
    //return this.doRequest({
    //  url: this.url_mgmt + '/bpl/getAppliancesInCluster',
    //  method: 'GET',
    //}).then(response => {
    //  if (response.status === 200) {
    //    return { status: "success", message: "Data source is working", title: "Success" };
    //  }
    //});
  }

  PVNamesFindQuery(query) {
    const str = this.templateSrv.replace(query, null, 'regex');

    if(!str) {
        let deferred = this.q.defer();
        deferred.resolve([]);
        return deferred.promise;
    }

    const url = this.url + "/bpl/getMatchingPVs?limit=100&pv=" + str;

    return this.doRequest({
      url: url,
      method: 'GET',
    }).then( res => {
        return res.data;
    });
  }

  doRequest(options) {
    options.withCredentials = this.withCredentials;
    options.headers = this.headers;

    const result = this.backendSrv.datasourceRequest(options);
    return result;
  }

  buildQueryParameters(options) {
    //remove placeholder targets and undefined targets
    options.targets = _.filter(options.targets, target => {
      return (target.target !== '' && typeof target.target !== 'undefined');
    });

    if (options.targets.length <= 0) {
      return options;
    }

    const from = new Date(options.range.from);
    const to = new Date(options.range.to);
    const range_msec = to.getTime() - from.getTime();
    const interval_sec =  _.floor(range_msec / ( options.maxDataPoints * 1000));

    let interval = "";
    if ( interval_sec >= 1 ) {
        interval = String(interval_sec);
    }

    var targets = _.map(options.targets, target => {
      return {
        target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        alias: target.alias,
        operator: target.operator,
        from: from,
        to: to,
        interval: interval,
        functions: target.functions
      };
    });

    options.targets = targets;

    return options;
  }
}

function bindFunctionDefs(functionDefs, category) {
  var aggregationFunctions = _.map(aafunc.getCategories()[category], 'name');
  var aggFuncDefs = _.filter(functionDefs, function(func) {
    return _.includes(aggregationFunctions, func.def.name);
  });

  return _.map(aggFuncDefs, func => {
    let funcInstance = aafunc.createFuncInstance(func.def, func.params);
    return funcInstance.bindFunction(dataProcessor.aaFunctions);
  });
}

function sequence(funcsArray) {
  return function(result) {
      for (let i = 0; i < funcsArray.length; i++) {
          result = funcsArray[i].call(this, result);
      }
      return result;
  };
}
