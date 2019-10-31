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

  }

  query(options) {
    var query = this.buildQueryParameters(options);
    query.targets = query.targets.filter(t => !t.hide);

    if (query.targets.length <= 0) {
      return this.q.when({data: []});
    }

    let pvs = new Array();
    query.targets.forEach(function(target) {
        pvs.push("pv=" + target.target);
    });

    let from = new Date(options.range.from);
    let to = new Date(options.range.to);

    return this.doRequest({
      url: this.url + '/data/getDataForPVs.json?' + pvs.join('&') + '&from=' + from.toISOString() + '&to=' + to.toISOString(),
      data: query,
      method: 'GET'
    }).then(this.responseParse);
  }

  responseParse(response) {
    let data = new Array();
    response.data.forEach(function(td) {
        let timesiries = new Array();
        td.data.forEach(function(d) {
            timesiries.push([d.val, d.secs*1000+Math.floor(d.nanos/1000000)]);
        });
        let d = {"target": td.meta["name"], "datapoints": timesiries};
        data.push(d);
    });
    return {data: data};
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
      return (target.target !== 'select metric' && typeof target.target !== 'undefined');
    });

    var targets = _.map(options.targets, target => {
      return {
        target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
      };
    });

    options.targets = targets;

    return options;
  }
}
