import { getBackendSrv } from '@grafana/runtime';
import {
  DataQueryResponse,
  DataQueryRequest,
  DataSourceInstanceSettings,
  DataSourceApi,
  MutableDataFrame,
  FieldType,
  getFieldDisplayName,
} from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
import _ from 'lodash';

import { AAQuery, AADataSourceOptions, operatorList } from './types';
import dataProcessor from './dataProcessor';
import * as aafunc from './aafunc';

/*
 * Variable format descriptions
 * ---
 * timeseries = {
 *   "target":"PV1", // Used as legend in Grafana
 *   "datapoints":[
 *     [622, 1450754160000], // Metric value as a float, unixtimestamp in milliseconds
 *     [365, 1450754220000]
 *   ]
 * }
 * timeseriesData = [ timeseries, timeseries, ... ]
 * timeseriesDataArray = [ timeseriesData, timeseriesData, ... ]
 */

export class DataSource extends DataSourceApi<AAQuery, AADataSourceOptions> {
  url?: string | undefined;
  name: string;
  templateSrv: any;
  withCredentials?: boolean;
  headers: { [key: string]: string };

  constructor(instanceSettings: DataSourceInstanceSettings<AADataSourceOptions>) {
    super(instanceSettings);
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.templateSrv = getTemplateSrv();
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = { 'Content-Type': 'application/json' };
    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers.Authorization = instanceSettings.basicAuth;
    }
  }

  // Called from Grafana panels to get data
  async query(options: DataQueryRequest<AAQuery>): Promise<DataQueryResponse> {
    const query = this.buildQueryParameters(options);

    // Remove hidden target from query
    query.targets = _.filter(query.targets, t => !t.hide);

    if (query.targets.length <= 0) {
      return Promise.resolve({ data: [] });
    }

    const targetProcesses = _.map(query.targets, target => this.targetProcess(target));

    return Promise.all(targetProcesses).then(timeseriesDataArray => this.postProcess(timeseriesDataArray));
  }

  targetProcess(target: any[]) {
    return this.buildUrls(target)
      .then((urls: string[]) => this.doMultiUrlRequests(urls))
      .then(responses => this.responseParse(responses))
      .then(timeseriesData => this.setAlias(timeseriesData, target))
      .then(timeseriesData => this.applyFunctions(timeseriesData, target));
  }

  postProcess(timeseriesDataArray: any) {
    const timeseriesData = _.flatten(timeseriesDataArray);

    return { data: timeseriesData };
  }

  buildUrls(target: any): Promise<string[]> {
    // Get Option values
    const maxNumPVs = target.options.maxNumPVs || 100;
    const binInterval = target.options.binInterval || target.interval;

    const targetPVs = this.parseTargetPV(target.target);

    // Create Promise to fetch PV names
    const pvnamesPromise = _.map(targetPVs, targetPV => {
      if (target.regex) {
        return this.pvNamesFindQuery(targetPV, maxNumPVs);
      }

      return Promise.resolve([targetPV]);
    });

    return Promise.all(pvnamesPromise).then(
      pvnamesArray =>
        new Promise((resolve, reject) => {
          const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, maxNumPVs);
          let urls;

          try {
            urls = _.map(pvnames, pvname =>
              this.buildUrl(pvname, target.operator, binInterval, target.from, target.to)
            );
          } catch (e) {
            reject(e);
          }

          resolve(urls);
        })
    );
  }

  buildUrl(pvname: string, operator: string, interval: string, from: Date, to: Date) {
    const pv = (() => {
      // raw Operator
      if (operator === 'raw' || interval === '') {
        return `${pvname}`;
      }

      // Default Operator
      if (_.includes(['', undefined], operator)) {
        return `mean_${interval}(${pvname})`;
      }

      // Other Operator
      if (_.includes(operatorList, operator)) {
        return `${operator}_${interval}(${pvname})`;
      }

      throw new Error('Data Processing Operator is invalid.');
    })();

    const url = `${this.url}/data/getData.json?pv=${encodeURIComponent(
      pv
    )}&from=${from.toISOString()}&to=${to.toISOString()}`;

    return url;
  }

  doMultiUrlRequests(urls: string[]) {
    const requests = _.map(urls, url => this.doRequest({ url, method: 'GET' }));

    return Promise.all(requests);
  }

  responseParse(responses: any[]) {
    const dataFramesArray = _.map(responses, response => {
      const dataFrames = _.map(response.data, targetRes => {
        const values = _.map(targetRes.data, datapoint => datapoint.val);
        const times = _.map(targetRes.data, datapoint => datapoint.secs * 1000 + _.floor(datapoint.nanos / 1000000));
        const frame = new MutableDataFrame({
          fields: [
            { name: 'time', type: FieldType.time, values: times },
            { name: 'value', type: FieldType.number, values: values, config: { displayName: targetRes.meta.name } },
          ],
        });
        return frame;
      });
      return dataFrames;
    });

    return Promise.resolve(_.flatten(dataFramesArray));
  }

  setAlias(dataFrameArray: MutableDataFrame[], target: any) {
    if (!target.alias) {
      return Promise.resolve(dataFrameArray);
    }

    let pattern: RegExp;
    if (target.aliasPattern) {
      pattern = new RegExp(target.aliasPattern, '');
    }

    const newDataFrameArray = _.map(dataFrameArray, dataFrame => {
      let alias = target.alias;
      const valfield = dataFrame.fields[1];
      const displayName = getFieldDisplayName(valfield, dataFrame);

      if (pattern) {
        alias = displayName.replace(pattern, alias);
      }

      const newValfield = {
        ...valfield,
        config: {
          ...valfield.config,
          displayName: alias,
        },
        state: {
          ...valfield.state,
          displayName: alias,
        },
      };

      return {
        ...dataFrame,
        fields: [dataFrame.fields[0], newValfield],
      };
    });

    return Promise.resolve(newDataFrameArray);
  }

  applyFunctions(timeseriesData: any, target: any) {
    if (target.functions === undefined) {
      return Promise.resolve(timeseriesData);
    }

    return this.applyFunctionDefs(target.functions, ['Transform', 'Filter Series'], timeseriesData);
  }

  // Called from Grafana data source configuration page to make sure the connection is working
  async testDatasource() {
    return { status: 'success', message: 'Data source is working', title: 'Success' };
  }

  pvNamesFindQuery(query: string, maxPvs: number) {
    if (!query) {
      return Promise.resolve([]);
    }

    const url = `${this.url}/bpl/getMatchingPVs?limit=${maxPvs}&regex=${encodeURIComponent(query)}`;

    return this.doRequest({
      url,
      method: 'GET',
    }).then(res => res.data);
  }

  // Called from Grafana variables to get values
  metricFindQuery(query: string) {
    /*
     * query format:
     * ex1) PV:NAME:.*
     * ex2) PV:NAME:.*?limit=10
     */
    const replacedQuery = this.templateSrv.replace(query, null, 'regex');
    const [pvQuery, paramsQuery] = replacedQuery.split('?', 2);
    const parsedPVs = this.parseTargetPV(pvQuery);

    // Parse query parameters
    let limitNum = 100;
    if (paramsQuery) {
      const params: URLSearchParams = new URLSearchParams(paramsQuery);
      const limit_param: string | null = params.get('limit');
      if (limit_param) {
        const limit = parseInt(limit_param, 10);
        limitNum = Number.isInteger(limit) ? limit : 100;
      }
    }

    const pvnamesPromise = _.map(parsedPVs, targetQuery => this.pvNamesFindQuery(targetQuery, limitNum));

    return Promise.all(pvnamesPromise).then(pvnamesArray => {
      const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, limitNum);
      return _.map(pvnames, pvname => ({ text: pvname }));
    });
  }

  doRequest(options: any) {
    const newOptions = { ...options };
    newOptions.withCredentials = this.withCredentials;
    newOptions.headers = this.headers;

    const result = getBackendSrv().datasourceRequest(newOptions);
    return result;
  }

  buildQueryParameters(options: any) {
    /*
     * options argument format
     * ---
     * {
     *   ...
     *   "range": { "from": "2015-12-22T03:06:13.851Z", "to": "2015-12-22T06:48:24.137Z" },
     *   "interval": "5s",
     *   "targets": [
     *     { "refId":"A",
     *       "target":"PV:NAME:.*",
     *       "regex":true,
     *       "operator":"mean",
     *       "alias":"$3",
     *       "aliasPattern":"(.*):(.*)",
     *       "functions":[
     *         {
     *           "text":"top($top_num, max)",
     *           "params":[ "$top_num", "max" ],
     *           "def":{
     *             "category":"Filter Series",
     *             "defaultParams":[ 5, "avg" ],
     *             "name":"top",
     *             "params":[
     *               { "name":"number", "type":"int" },
     *               {
     *                 "name":"value",
     *                 "options":[ "avg", "min", "max", "absoluteMin", "absoluteMax", "sum" ],
     *                 "type":"string"
     *               }
     *             ]
     *           },
     *         }
     *       ],
     *     }
     *   ],
     *   "format": "json",
     *   "maxDataPoints": 2495 // decided by the panel
     *   ...
     * }
     */
    const query = { ...options };

    // remove placeholder targets and undefined targets
    query.targets = _.filter(query.targets, target => target.target !== '' && typeof target.target !== 'undefined');

    if (query.targets.length <= 0) {
      return query;
    }

    const from = new Date(query.range.from);
    const to = new Date(query.range.to);
    const rangeMsec = to.getTime() - from.getTime();
    const intervalSec = _.floor(rangeMsec / (query.maxDataPoints * 1000));

    const interval = intervalSec >= 1 ? String(intervalSec) : '';

    const targets = _.map(query.targets, target => {
      // Replace parameters with variables for each functions
      const functions = _.map(target.functions, func => {
        const newFunc = func;
        newFunc.params = _.map(newFunc.params, param => this.templateSrv.replace(param, query.scopedVars, 'regex'));
        return newFunc;
      });

      return {
        target: this.templateSrv.replace(target.target, query.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        alias: this.templateSrv.replace(target.alias, query.scopedVars, 'regex'),
        operator: this.templateSrv.replace(target.operator, query.scopedVars, 'regex'),
        functions,
        regex: target.regex,
        aliasPattern: target.aliasPattern,
        options: this.getOptions(target.functions),
        from,
        to,
        interval,
      };
    });

    query.targets = targets;

    return query;
  }

  parseTargetPV(targetPV: string) {
    /*
     * ex) targetPV = ABC(1|2|3)EFG(5|6)
     *     then
     *     splitQueries = ['ABC','(1|2|3'), 'EFG', '(5|6)']
     *     queries = [
     *     ABC1EFG5, ABC1EFG6, ABC2EFG6,
     *     ABC2EFG6, ABC3EFG5, ABC3EFG6
     *     ]
     */
    const splitQueries = _.split(targetPV, /(\(.*?\))/);
    let queries = [''];

    _.forEach(splitQueries, (splitQuery, i) => {
      // Fixed string like 'ABC'
      if (i % 2 === 0) {
        queries = _.map(queries, query => `${query}${splitQuery}`);
        return;
      }

      // Regex OR string like '(1|2|3)'
      const orElems = _.split(_.trim(splitQuery, '()'), '|');

      const newQueries = _.map(queries, query => _.map(orElems, orElem => `${query}${orElem}`));
      queries = _.flatten(newQueries);
    });

    return queries;
  }

  applyFunctionDefs(functionDefs: any, categories: string[], data: any) {
    const applyFuncDefs = this.pickFuncDefsFromCategories(functionDefs, categories);

    const promises = _.reduce(
      applyFuncDefs,
      (prevPromise, func) =>
        prevPromise.then(res => {
          const funcInstance = aafunc.createFuncInstance(func.def, func.params);
          const bindedFunc = funcInstance.bindFunction(dataProcessor.aaFunctions);

          return Promise.resolve(bindedFunc(res));
        }),
      Promise.resolve(data)
    );

    return promises;
  }

  getOptions(functionDefs: any) {
    const appliedOptionFuncs = this.pickFuncDefsFromCategories(functionDefs, ['Options']);

    const options = _.reduce(
      appliedOptionFuncs,
      (optionMap: any, func: any) => {
        [optionMap[func.def.name]] = func.params;
        return optionMap;
      },
      {}
    );

    return options;
  }

  pickFuncDefsFromCategories(functionDefs: any, categories: string[]) {
    const allCategorisedFuncDefs = aafunc.getCategories();

    const requiredCategoryFuncNames = _.reduce(
      categories,
      (funcNames: string[], category: string) => _.concat(funcNames, _.map(allCategorisedFuncDefs[category], 'name')),
      []
    );

    const pickedFuncDefs = _.filter(functionDefs, func => _.includes(requiredCategoryFuncNames, func.def.name));

    return pickedFuncDefs;
  }
}
