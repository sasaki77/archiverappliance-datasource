import _ from 'lodash';
import dataProcessor from './dataProcessor';
import * as aafunc from './aafunc';

export class ArchiverapplianceDatasource {
  constructor(instanceSettings, backendSrv, templateSrv) {
    this.type = instanceSettings.type;
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.backendSrv = backendSrv;
    this.templateSrv = templateSrv;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = { 'Content-Type': 'application/json' };
    if (
      typeof instanceSettings.basicAuth === 'string'
      && instanceSettings.basicAuth.length > 0
    ) {
      this.headers.Authorization = instanceSettings.basicAuth;
    }

    this.operatorList = [
      'firstSample', 'lastSample', 'firstFill', 'lastFill', 'mean', 'min',
      'max', 'count', 'ncount', 'nth', 'median', 'std', 'jitter',
      'ignoreflyers', 'flyers', 'variance',
      'popvariance', 'kurtosis', 'skewness', 'raw',
    ];
  }

  query(options) {
    const query = this.buildQueryParameters(options);
    query.targets = _.filter(query.targets, (t) => !t.hide);

    if (query.targets.length <= 0) {
      return Promise.resolve({ data: [] });
    }

    const targetProcesses = _.map(query.targets, (target) => (
      this.targetProcess(target)
    ));

    return (
      Promise.all(targetProcesses)
        .then((timeseriesDataArray) => this.postProcess(timeseriesDataArray))
    );
  }

  targetProcess(target) {
    return (
      this.buildUrls(target)
        .then((urls) => this.doMultiUrlRequests(urls))
        .then((responses) => this.responseParse(responses))
        .then((timeseriesData) => this.setAlias(timeseriesData, target))
        .then((timeseriesData) => this.applyFunctions(timeseriesData, target))
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

      return Promise.resolve([targetQuery]);
    });

    return Promise.all(pvnamesPromise)
      .then((pvnamesArray) => (
        new Promise((resolve, reject) => {
          const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, 100);
          let urls;

          try {
            urls = _.map(pvnames, (pvname) => (
              this.buildUrl(
                pvname,
                target.operator,
                target.interval,
                target.from,
                target.to,
              )
            ));
          } catch (e) {
            reject(e);
          }

          resolve(urls);
        })
      ));
  }

  buildUrl(pvname, operator, interval, from, to) {
    let pv = '';
    if (operator === 'raw' || interval === '') {
      pv = `${pvname}`;
    } else if (_.includes(['', undefined], operator)) {
      // Default Operator
      pv = `mean_${interval}(${pvname})`;
    } else if (_.includes(this.operatorList, operator)) {
      pv = `${operator}_${interval}(${pvname})`;
    } else {
      throw new Error('Data Processing Operator is invalid.');
    }

    const url = `${this.url}/data/getData.json?pv=${encodeURIComponent(pv)}&from=${from.toISOString()}&to=${to.toISOString()}`;

    return url;
  }

  doMultiUrlRequests(urls) {
    const requests = _.map(urls, (url) => (
      this.doRequest({ url, method: 'GET' })
    ));

    return Promise.all(requests);
  }

  responseParse(responses) {
    const timeSeriesDataArray = _.map(responses, (response) => {
      const timeSeriesData = _.map(response.data, (targetRes) => {
        const timesiries = _.map(targetRes.data, (datapoint) => (
          [
            datapoint.val,
            datapoint.secs * 1000 + _.floor(datapoint.nanos / 1000000),
          ]
        ));
        const timeseries = { target: targetRes.meta.name, datapoints: timesiries };
        return timeseries;
      });
      return timeSeriesData;
    });

    return Promise.resolve(_.flatten(timeSeriesDataArray));
  }

  setAlias(timeseriesData, target) {
    if (!target.alias) {
      return Promise.resolve(timeseriesData);
    }

    let pattern;
    if (target.aliasPattern) {
      pattern = new RegExp(target.aliasPattern, '');
    }

    const newTimeseriesData = _.map(timeseriesData, (timeseries) => {
      if (pattern) {
        const alias = timeseries.target.replace(pattern, target.alias);
        return { target: alias, datapoints: timeseries.datapoints };
      }

      return { target: target.alias, datapoints: timeseries.datapoints };
    });

    return Promise.resolve(newTimeseriesData);
  }

  applyFunctions(timeseriesData, target) {
    if (target.functions === undefined) {
      return Promise.resolve(timeseriesData);
    }

    return this.bindFunctionDefs(target.functions, ['Transform', 'Filter Series'], timeseriesData);
  }

  testDatasource() {
    return { status: 'success', message: 'Data source is working', title: 'Success' };
  }

  pvNamesFindQuery(query) {
    if (!query) {
      return Promise.resolve([]);
    }

    const url = `${this.url}/bpl/getMatchingPVs?limit=100&regex=${encodeURIComponent(query)}`;

    return this.doRequest({
      url,
      method: 'GET',
    }).then((res) => res.data);
  }

  metricFindQuery(query) {
    const replacedQuery = this.templateSrv.replace(query, null, 'regex');
    const parsedQuery = this.parseTargetQuery(replacedQuery);

    const pvnamesPromise = _.map(parsedQuery, (targetQuery) => (
      this.pvNamesFindQuery(targetQuery)
    ));

    return Promise.all(pvnamesPromise).then((pvnamesArray) => {
      const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, 100);
      return _.map(pvnames, (pvname) => ({ text: pvname }));
    });
  }

  doRequest(options) {
    const newOptions = { ...options };
    newOptions.withCredentials = this.withCredentials;
    newOptions.headers = this.headers;

    const result = this.backendSrv.datasourceRequest(newOptions);
    return result;
  }

  buildQueryParameters(options) {
    const query = { ...options };

    // remove placeholder targets and undefined targets
    query.targets = _.filter(query.targets, (target) => (
      target.target !== '' && typeof target.target !== 'undefined'
    ));

    if (query.targets.length <= 0) {
      return query;
    }

    const from = new Date(query.range.from);
    const to = new Date(query.range.to);
    const rangeMsec = to.getTime() - from.getTime();
    const intervalSec = _.floor(rangeMsec / (query.maxDataPoints * 1000));

    const interval = (intervalSec >= 1) ? String(intervalSec) : '';

    const targets = _.map(query.targets, (target) => (
      {
        target: this.templateSrv.replace(target.target, query.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        alias: target.alias,
        operator: target.operator,
        functions: target.functions,
        regex: target.regex,
        aliasPattern: target.aliasPattern,
        from,
        to,
        interval,
      }
    ));

    query.targets = targets;

    return query;
  }

  parseTargetQuery(targetQuery) {
    /*
     * ex) targetQuery = ABC(1|2|3)EFG(5|6)
     *     then
     *     splitQueries = ['ABC','(1|2|3'), 'EFG', '(5|6)']
     *     queries = [
     *     ABC1EFG5, ABC1EFG6, ABC2EFG6,
     *     ABC2EFG6, ABC3EFG5, ABC3EFG6
     *     ]
     */
    const splitQueries = _.split(targetQuery, /(\(.*?\))/);
    let queries = [''];

    _.forEach(splitQueries, (splitQuery, i) => {
      // Fixed string like 'ABC'
      if (i % 2 === 0) {
        queries = _.map(queries, (query) => `${query}${splitQuery}`);
        return;
      }

      // Regex OR string like '(1|2|3)'
      const orElems = _.split(_.trim(splitQuery, '()'), '|');

      const newQueries = _.map(queries, (query) => (
        _.map(orElems, (orElem) => `${query}${orElem}`)
      ));
      queries = _.flatten(newQueries);
    });

    return queries;
  }

  bindFunctionDefs(functionDefs, categories, data) {
    const allCategorisedFuncDefs = aafunc.getCategories();

    const requiredCategoryFuncNames = _.reduce(categories, (funcNames, category) => (
      _.concat(funcNames, _.map(allCategorisedFuncDefs[category], 'name'))
    ), []);

    const applyFuncDefs = _.filter(functionDefs, (func) => (
      _.includes(requiredCategoryFuncNames, func.def.name)
    ));

    const promises = _.reduce(applyFuncDefs, (prevPromise, func) => (
      prevPromise.then((res) => {
        const funcInstance = aafunc.createFuncInstance(func.def, func.params);
        const bindedFunc = funcInstance.bindFunction(dataProcessor.aaFunctions);

        return Promise.resolve(bindedFunc(res));
      })
    ), Promise.resolve(data));

    return promises;
  }
}
