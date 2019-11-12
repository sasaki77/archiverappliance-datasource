import _ from "lodash";

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
    this.operatorList = [ "firstSample", "lastSample", "firstFill", "lastFill", "mean", "min", "max",
        "count", "ncount", "nth", "median", "std", "jitter", "ignoreflyers", "flyers", "variance",
        "popvariance", "kurtosis", "skewness", "raw"];
  }

  query(options) {
    var query = this.buildQueryParameters(options);
    query.targets = query.targets.filter(t => !t.hide);

    if (query.targets.length <= 0) {
      return this.q.when({data: []});
    }

    const targets = query.targets.map( target => {
      return this.targetProcess(target, options);
    });

   return this.q.all(targets).then( data => this.postProcess(data) );
  }

  targetProcess(target, options) {
      return this.buildUrl(target, options)
          .then( url => this.doRequest({
                  url: url,
                  method: 'GET'
              }))
          .then( res => this.responseParse(res) )
          .then( data => this.setAlias(data, target) );
  }

  postProcess(data) {
    const d = data.reduce( (result, d) => {
      result = result.concat(d);
      return result;
    }, []);

    return {data: d};
  }

  buildUrl(target) {
    let deferred = this.q.defer();

    let pv = ""
    if ( target.operator === "raw" || target.interval === "") {
      pv = "pv=" + target.target;
    } else if ( ["", undefined].includes(target.operator) ) {
      // Default Operator
      pv = "pv=mean_" + target.interval + "(" + target.target + ")";
    } else if ( this.operatorList.includes(target.operator) ) {
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

    const target_data = response.data.map( target_res => {
      const timesiries = target_res.data.map( datapoint => {
          return [datapoint.val, datapoint.secs*1000+Math.floor(datapoint.nanos/1000000)];
      });
      const target_data = {"target": target_res.meta["name"], "datapoints": timesiries};
      return target_data;
    });

    deferred.resolve(target_data);
    return deferred.promise;
  }

  setAlias(data, target) {
    let deferred = this.q.defer();

    data.forEach( d => {
      if( target.alias !== undefined && target.alias !== "" ) {
        d.target = target.alias;
      }
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

  metricFindQuery(query) {
    var str = this.templateSrv.replace(query, null, 'regex');

    if (str) {
      var s = str.toString().split('=');
      var target = (s[1] || '');
      var name = (s[0] || '');
    }
    else{
      var target = '';
      var name = '';
    }

    var interpolated = {
        target: target
    };

    interpolated.name = name;

    return this.doRequest({
      url: this.url + '/search',
      data: interpolated,
      method: 'POST',
    }).then(this.mapToTextValue);
  }

  mapToTextValue(result) {
    return _.map(result.data, (d, i) => {
      if (d && d.text && d.value) {
        return { text: d.text, value: d.value };
      } else if (_.isObject(d)) {
        return { text: d, value: i};
      }
      return { text: d, value: d };
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
    const interval_sec =  Math.floor(range_msec / ( options.maxDataPoints * 1000));

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
        interval: interval
      };
    });

    options.targets = targets;

    return options;
  }
}
