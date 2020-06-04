import _ from 'lodash';
import { getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import {
  DataQueryResponse,
  DataQueryRequest,
  DataSourceInstanceSettings,
  DataSourceApi,
  MutableDataFrame,
  FieldType,
  getFieldDisplayName,
} from '@grafana/data';
import dataProcessor from './dataProcessor';
import * as aafunc from './aafunc';
import {
  AAQuery,
  AADataSourceOptions,
  TargetQuery,
  AADataQueryResponse,
  FunctionDescriptor,
  operatorList,
} from './types';

export class DataSource extends DataSourceApi<AAQuery, AADataSourceOptions> {
  url?: string | undefined;
  name: string;
  withCredentials?: boolean;
  headers: { [key: string]: string };

  constructor(instanceSettings: DataSourceInstanceSettings<AADataSourceOptions>) {
    super(instanceSettings);
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = { 'Content-Type': 'application/json' };
    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers.Authorization = instanceSettings.basicAuth;
    }
  }

  // Called from Grafana panels to get data
  async query(options: DataQueryRequest<AAQuery>): Promise<DataQueryResponse> {
    const rawTargets = this.buildQueryParameters(options);

    // Remove hidden target from query
    const targets = _.filter(rawTargets, t => !t.hide);

    if (targets.length <= 0) {
      return Promise.resolve({ data: [] });
    }

    const targetProcesses = _.map(targets, target => this.targetProcess(target));

    return Promise.all(targetProcesses).then(dataFramesArray => this.postProcess(dataFramesArray));
  }

  targetProcess(target: TargetQuery) {
    return this.buildUrls(target)
      .then((urls: string[]) => this.doMultiUrlRequests(urls))
      .then(responses => this.responseParse(responses))
      .then(dataFrames => this.setAlias(dataFrames, target))
      .then(dataFrames => this.applyFunctions(dataFrames, target));
  }

  postProcess(dataFramesArray: MutableDataFrame[][]) {
    const dataFrames = _.flatten(dataFramesArray);

    return { data: dataFrames };
  }

  buildUrls(target: TargetQuery): Promise<string[]> {
    // Get Option values
    const maxNumPVs = Number(target.options.maxNumPVs) || 100;
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

  responseParse(responses: AADataQueryResponse[]) {
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

  async setAlias(dataFrames: MutableDataFrame[], target: TargetQuery): Promise<MutableDataFrame[]> {
    if (!target.alias) {
      return Promise.resolve(dataFrames);
    }

    let pattern: RegExp;
    if (target.aliasPattern) {
      pattern = new RegExp(target.aliasPattern, '');
    }

    const newDataFrames = _.map(dataFrames, dataFrame => {
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

      return new MutableDataFrame({
        ...dataFrame,
        fields: [dataFrame.fields[0], newValfield],
      });
    });

    return Promise.resolve(newDataFrames);
  }

  applyFunctions(dataFrames: MutableDataFrame[], target: TargetQuery) {
    if (target.functions === undefined) {
      return Promise.resolve(dataFrames);
    }

    return this.applyFunctionDefs(target.functions, ['Transform', 'Filter Series'], dataFrames);
  }

  // Called from Grafana data source configuration page to make sure the connection is working
  async testDatasource() {
    return { status: 'success', message: 'Data source is working', title: 'Success' };
  }

  pvNamesFindQuery(query: string | undefined | null, maxPvs: number) {
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
    const templateSrv = getTemplateSrv();
    const replacedQuery = templateSrv.replace(query, undefined, 'regex');
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

  doRequest(options: {
    method?: string;
    url: any;
    requestId?: any;
    withCredentials?: any;
    headers?: any;
    inspect?: any;
  }) {
    const newOptions = { ...options };
    newOptions.withCredentials = this.withCredentials;
    newOptions.headers = this.headers;

    const result = getBackendSrv().datasourceRequest(newOptions);
    return result;
  }

  buildQueryParameters(options: DataQueryRequest<AAQuery>) {
    const templateSrv = getTemplateSrv();
    const query = { ...options };

    // remove placeholder targets and undefined targets
    query.targets = _.filter(query.targets, target => target.target !== '' && typeof target.target !== 'undefined');

    if (query.targets.length <= 0) {
      return [];
    }

    const from = new Date(String(query.range.from));
    const to = new Date(String(query.range.to));
    const rangeMsec = to.getTime() - from.getTime();
    const maxDataPoints = query.maxDataPoints || 2000;
    const intervalSec = _.floor(rangeMsec / (maxDataPoints * 1000));

    const interval = intervalSec >= 1 ? String(intervalSec) : '';

    const targets: TargetQuery[] = _.map(query.targets, target => {
      // Replace parameters with variables for each functions
      const functions = _.map(target.functions, func => {
        const newFunc = func;
        newFunc.params = _.map(newFunc.params, param => templateSrv.replace(param, query.scopedVars, 'regex'));
        return newFunc;
      });

      return {
        target: templateSrv.replace(target.target, query.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        alias: templateSrv.replace(target.alias, query.scopedVars, 'regex'),
        operator: templateSrv.replace(target.operator, query.scopedVars, 'regex'),
        functions,
        regex: target.regex,
        aliasPattern: target.aliasPattern,
        options: this.getOptions(target.functions),
        from,
        to,
        interval,
      };
    });

    return targets;
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

  applyFunctionDefs(functionDefs: FunctionDescriptor[], categories: string[], dataFrames: MutableDataFrame[]) {
    const applyFuncDefs = this.pickFuncDefsFromCategories(functionDefs, categories);

    const promises = _.reduce(
      applyFuncDefs,
      (prevPromise, func) =>
        prevPromise.then(res => {
          const funcInstance = aafunc.createFuncInstance(func.def, func.params);
          const bindedFunc = funcInstance.bindFunction(dataProcessor.aaFunctions);

          return Promise.resolve(bindedFunc(res));
        }),
      Promise.resolve(dataFrames)
    );

    return promises;
  }

  getOptions(functionDefs: FunctionDescriptor[]) {
    const appliedOptionFuncs = this.pickFuncDefsFromCategories(functionDefs, ['Options']);

    const options = _.reduce(
      appliedOptionFuncs,
      (optionMap: { [key: string]: string }, func) => {
        [optionMap[func.def.name]] = func.params;
        return optionMap;
      },
      {}
    );

    return options;
  }

  pickFuncDefsFromCategories(functionDefs: FunctionDescriptor[], categories: string[]) {
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
