import _ from 'lodash';
import dataProcessor from './dataProcessor';
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
    if (
        typeof instanceSettings.basicAuth === 'string'
        && instanceSettings.basicAuth.length > 0
    ) {
      this.headers['Authorization'] = instanceSettings.basicAuth;
    }

    this.operatorList = [
        'firstSample', 'lastSample', 'firstFill', 'lastFill', 'mean', 'min',
        'max', 'count', 'ncount', 'nth', 'median', 'std', 'jitter',
        'ignoreflyers', 'flyers', 'variance',
        'popvariance', 'kurtosis', 'skewness', 'raw'
    ];
  }

  query(options) {
    let query = this.buildQueryParameters(options);
    query.targets = _.filter(query.targets, t => !t.hide);

    if (query.targets.length <= 0) {
      return this.q.when({ data: [] });
    }

    const targetProcesses = _.map(query.targets, (target) => {
      return this.targetProcess(target, options);
    });

   return this.q.all(targetProcesses)
          .then((timeseriesDataArray) => this.postProcess(timeseriesDataArray));
  }

  targetProcess(target, options) {
    return (
      this.buildUrls(target, options)
      .then(urls           => this.doMultiUrlRequests(urls))
      .then(responses      => this.responseParse(responses))
      .then(timeseriesData => this.setAlias(timeseriesData, target))
      .then(timeseriesData => this.applyFunctions(timeseriesData, target))
    );
  }

  postProcess(timeseriesDataArray) {
    const timeseriesData = _.flatten(timeseriesDataArray);

    return { data: timeseriesData };
  }

  buildUrls(target) {
    const targetQueries = this.parseTargetQuery(target.target);

    const pvnamesPromise = _.map(targetQueries, (targetQuery) => {
      if (target.regex) {
        return this.pvNamesFindQuery(targetQuery);
      }

      return this.q.when([targetQuery]);
    });

    return this.q.all(pvnamesPromise)
          .then((pvnamesArray) => {
            const pvnames = _.slice(_.uniq(_.flattenDeep(pvnamesArray)), 0, 100);
            let deferred = this.q.defer();
            let urls;

            try {
              urls = _.map( pvnames, (pvname) => {
                return this.buildUrl(
                  pvname,
                  target.operator,
                  target.interval,
                  target.from,
                  target.to
                );
              });
            } catch (e) {
              deferred.reject(e);
            }

            deferred.resolve(urls);
            return deferred.promise;
          });
  }

  buildUrl(pvname, operator, interval, from, to) {
    let pv = ''
    if (operator === 'raw' || interval === '') {
      pv = ['pv=', pvname].join('');
    } else if (_.includes(['', undefined], operator)) {
      // Default Operator
      pv = ['pv=mean_', interval, '(', pvname, ')'].join('');
    } else if (_.includes(this.operatorList, operator) ) {
      pv = ['pv=', operator, '_', interval, '(', pvname, ')'].join('');
    } else {
      throw new Error('Data Processing Operator is invalid.');
    }

    const url = [
      this.url,
      '/data/getData.json?',
      pv,
      '&from=',
      from.toISOString(),
      '&to=',
      to.toISOString()
    ].join('');

    return url;
  }

  doMultiUrlRequests(urls) {
    const requests = _.map(urls, (url) => {
      return this.doRequest({ url: url, method: 'GET' });
    });

    return this.q.all(requests);
  }

  responseParse(responses) {
    let deferred = this.q.defer();

    const timeSeriesDataArray = _.map(responses, (response) => {
      const timeSeriesData = _.map(response.data, (target_res) => {
        const timesiries = _.map( target_res.data, (datapoint) => {
          return [
            datapoint.val,
            datapoint.secs * 1000 + _.floor(datapoint.nanos / 1000000)
          ];
        });
        const timeseries = { target: target_res.meta['name'], datapoints: timesiries };
        return timeseries;
      });
      return timeSeriesData;
    });

    deferred.resolve(_.flatten(timeSeriesDataArray));
    return deferred.promise;
  }

  setAlias(timeseriesData, target) {
    let deferred = this.q.defer();

    if (!target.alias) {
      deferred.resolve(timeseriesData);
      return deferred.promise;
    }

    let pattern;
    if (target.aliasPattern) {
      pattern = new RegExp(target.aliasPattern, '');
    }

    let newTimeseriesData = _.map(timeseriesData, (timeseries) => {
      if (pattern) {
        const alias = timeseries.target.replace(pattern, target.alias);
        return { target: alias, datapoints: timeseries.datapoints };
      }

      return { target: target.alias, datapoints: timeseries.datapoints };
    });

    deferred.resolve(newTimeseriesData);
    return deferred.promise;
  }

  applyFunctions(timeseriesData, target) {
    let deferred = this.q.defer();

    if (target.functions === undefined) {
      deferred.resolve(timeseriesData);
      return deferred.promise;
    }

    // Apply transformation functions
    const transformFunctions = bindFunctionDefs(target.functions, 'Transform');
    timeseriesData = _.map(timeseriesData, (timeseries) => {
      timeseries.datapoints = sequence(transformFunctions)(timeseries.datapoints);
      return timeseries;
    });

    deferred.resolve(timeseriesData);
    return deferred.promise;
  }

  testDatasource() {
    return { status: 'success', message: 'Data source is working', title: 'Success' };
    //return this.doRequest({
    //  url: this.url_mgmt + '/bpl/getAppliancesInCluster',
    //  method: 'GET',
    //}).then(response => {
    //  if (response.status === 200) {
    //    return { status: 'success', message: 'Data source is working', title: 'Success' };
    //  }
    //});
  }

  pvNamesFindQuery(query) {
    if (!query) {
      let deferred = this.q.defer();
      deferred.resolve([]);
      return deferred.promise;
    }

    const url = [
      this.url,
      '/bpl/getMatchingPVs?limit=100&regex=',
      encodeURIComponent(query)
    ].join('');

    return this.doRequest({
      url: url,
      method: 'GET',
    }).then( (res) => {
      return res.data;
    });
  }

  metricFindQuery(query) {
    const replacedQuery = this.templateSrv.replace(query, null, 'regex');
    const parsedQuery = this.parseTargetQuery(replacedQuery);

    const pvnamesPromise = _.map(parsedQuery, (targetQuery) => {
      return this.pvNamesFindQuery(targetQuery);
    });

    return this.q.all(pvnamesPromise).then((pvnamesArray) => {
      const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, 100);
      return _.map(pvnames, (pvname) => {
        return { text: pvname };
      });
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
    options.targets = _.filter(options.targets, (target) => {
      return (target.target !== '' && typeof target.target !== 'undefined');
    });

    if (options.targets.length <= 0) {
      return options;
    }

    const from = new Date(options.range.from);
    const to = new Date(options.range.to);
    const rangeMsec = to.getTime() - from.getTime();
    const intervalSec =  _.floor(rangeMsec / ( options.maxDataPoints * 1000));

    let interval = '';
    if ( intervalSec >= 1 ) {
      interval = String(intervalSec);
    }

    const targets = _.map(options.targets, (target) => {
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
        aliasPattern: target.aliasPattern
      };
    });

    options.targets = targets;

    return options;
  }

  parseTargetQuery(query) {
    /*
     * ex) query = ABC(1|2|3)EFG(5|6)
     *     then
     *     splitQueries = ['ABC','(1|2|3'), 'EFG', '(5|6)']
     *     queries = [
     *     ABC1EFG5, ABC1EFG6, ABC2EFG6,
     *     ABC2EFG6, ABC3EFG5, ABC3EFG6
     *     ]
     */
    const splitQueries = _.split(query, /(\(.*?\))/);
    let queries = [''];

    _.forEach(splitQueries, (splitQuery, i) => {
      // Fixed string like 'ABC'
      if (i % 2 === 0) {
        queries = _.map(queries, (query) => {
          return [query, splitQuery].join('');
        });
        return;
      }

      // Regex OR string like '(1|2|3)'
      const orElems = _.split(_.trim(splitQuery, '()'), '|');

      const newQueries = _.map(queries, (query) => {
        return _.map(orElems, (orElem) => {
          return [query, orElem].join('');
        });
      });
      queries = _.flatten(newQueries);
    });

    return queries;
  }
}

function bindFunctionDefs(functionDefs, category) {
  const aggregationFunctions = _.map(aafunc.getCategories()[category], 'name');
  const aggFuncDefs = _.filter(functionDefs, function(func) {
    return _.includes(aggregationFunctions, func.def.name);
  });

  return _.map(aggFuncDefs, func => {
    let funcInstance = aafunc.createFuncInstance(func.def, func.params);
    return funcInstance.bindFunction(dataProcessor.aaFunctions);
  });
}

function sequence(funcsArray) {
  return function (result) {
    for (let i = 0; i < funcsArray.length; i++) {
      result = funcsArray[i].call(this, result);
    }
    return result;
  };
}
