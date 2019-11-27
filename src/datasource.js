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
          this.buildUrls(target, options)
          .then( urls => this.doMultiUrlRequests(urls) )
          .then( responses => this.responseParse(responses) )
          .then( data => this.setAlias(data, target) )
          .then( data => this.applyFunctions(data, target) )
      );
  }

  postProcess(data) {
    const d = _.flatten( data );

    return {data: d};
  }

  buildUrls(target) {
    let pvnames_promise;
    if (target.regex) {
      pvnames_promise = this.PVNamesFindQuery(target.target);
    } else {
      let pvnames = this.q.defer();
      pvnames.resolve([target.target]);
      pvnames_promise = pvnames.promise;
    }

    return pvnames_promise.then( pvnames => {
      let deferred = this.q.defer();
      let urls;
      try {
        urls = _.map( pvnames, pvname => {
          return this.buildUrl(pvname, target.operator, target.interval, target.from, target.to);
        });
      } catch (e) {
        deferred.reject(e);
      }
      deferred.resolve(urls);
      return deferred.promise;
    });
  }

  buildUrl(pvname, operator, interval, from, to) {
    let pv = ""
    if ( operator === "raw" || interval === "") {
      pv = "pv=" + pvname;
    } else if ( _.includes(["", undefined], operator) ) {
      // Default Operator
      pv = "pv=mean_" + interval + "(" + pvname + ")";
    } else if ( _.includes(this.operatorList, operator) ) {
      pv = "pv=" + operator + "_" + interval + "(" + pvname + ")";
    } else {
      throw new Error("Data Processing Operator is invalid.");
    }

    const url = this.url + '/data/getData.json?' + pv + '&from=' + from.toISOString() + '&to=' + to.toISOString();

    return url;
  }

  doMultiUrlRequests(urls) {
    const requests = _.map( urls, url => {
      return this.doRequest({ url: url, method: "GET" });
    });

    return this.q.all(requests);
  }

  responseParse(responses) {
    let deferred = this.q.defer();

    const target_data = _.map( responses, response => {
      const td = _.map( response.data, target_res => {
        const timesiries = _.map( target_res.data, datapoint => {
            return [datapoint.val, datapoint.secs*1000+_.floor(datapoint.nanos/1000000)];
        });
        const target_data = {"target": target_res.meta["name"], "datapoints": timesiries};
        return target_data;
      });
      return td;
    });

    deferred.resolve(_.flatten(target_data));
    return deferred.promise;
  }

  setAlias(data, target) {
    let deferred = this.q.defer();

    if( !target.alias ) {
      deferred.resolve(data);
      return deferred.promise;
    }

    let pattern;
    if (target.alias_regexp) {
      pattern = new RegExp(target.alias_regexp, "");
    }

    _.forEach( data, d => {
      if ( pattern ) {
        d.target = d.target.replace(pattern, target.alias);
      } else {
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

    const url = this.url + "/bpl/getMatchingPVs?limit=100&regex=" + encodeURIComponent(str);

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
        functions: target.functions,
        regex: target.regex,
        alias_regexp: target.alias_regexp
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
