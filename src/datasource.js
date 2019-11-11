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

    const urls = this.buildUrl(query, options);

    const requests = urls.map( url => {
      return this.doRequest({
        url: url,
        data: query,
        method: 'GET'
      })
    });

   return this.q.all(requests)
          .then( res => this.responseParse(res, query) );
  }

  buildUrl(query, options) {
    let interval = "";
    if ( options.intervalMs > 1000 ) {
      interval = String(options.intervalMs / 1000 - 1);
    }

    const pvs = query.targets.reduce( (pvs, target) => {
      if ( ["raw", "", undefined].includes(target.operator) || interval === "") {
        pvs.push("pv=" + target.target);
      } else if ( this.operatorList.includes(target.operator) ) {
        pvs.push("pv=" + target.operator + "_" + interval + "(" + target.target + ")");
      }
      return pvs;
    }, []);

    const from = new Date(options.range.from);
    const to = new Date(options.range.to);
    const urls = pvs.map( pv => {
      return this.url + '/data/getData.json?' + pv + '&from=' + from.toISOString() + '&to=' + to.toISOString();
    });

    return urls;
  }

  responseParse(responses, query) {
    let data = responses.reduce( (data, response) => {
      let targets_data = response.data.map( target_res => {
          const timesiries = target_res.data.map( datapoint => {
              return [datapoint.val, datapoint.secs*1000+Math.floor(datapoint.nanos/1000000)];
          });
          const target_data = {"target": target_res.meta["name"], "datapoints": timesiries};
          return target_data;
      });
      data = data.concat(targets_data);
      return data;
    }, []);

    this.setAlias(data, query.targets);

    return {data: data};
  }

  setAlias(data, targets) {
      let aliases = {};

      targets.forEach( target => {
        if( target.alias !== undefined && target.alias !== "" ) {
          aliases[target.target] = target.alias;
        }
      });

      data.forEach( d => {
        if( aliases[d.target] !== undefined ) {
          d.target = aliases[d.target];
        }
      });
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

    var targets = _.map(options.targets, target => {
      return {
        target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        alias: target.alias,
        operator: target.operator,
      };
    });

    options.targets = targets;

    return options;
  }
}
