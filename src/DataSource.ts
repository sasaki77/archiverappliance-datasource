import _ from 'lodash';
import { Observable, from } from 'rxjs';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import {
  DataQueryResponse,
  DataQueryRequest,
  DataSourceInstanceSettings,
} from '@grafana/data';

import {
  AAQuery,
  AADataSourceOptions,
  TargetQuery,
} from './types';
import { getOptions } from './aafunc';
import { doQuery } from './query';
import { AAclient } from './aaclient'
import { parseTargetPV } from 'pvnameParser';
import { StreamQuery } from './streamQuery';

export class DataSource extends DataSourceWithBackend<AAQuery, AADataSourceOptions> {
  name: string;
  useBackend?: boolean | undefined;
  defaultOperator?: string;
  useLiveUpdate?: boolean | undefined;
  liveUpdateURI?: string | undefined;
  aaclient: AAclient;
  streamQuery: StreamQuery;

  constructor(instanceSettings: DataSourceInstanceSettings<AADataSourceOptions>) {
    super(instanceSettings);
    let url = "http://localhost:17668/retrieval";
    if (instanceSettings.url) {
      url = instanceSettings.url;
    }
    this.name = instanceSettings.name;
    this.useBackend = instanceSettings.jsonData.useBackend;
    this.defaultOperator = instanceSettings.jsonData.defaultOperator;
    this.useLiveUpdate = instanceSettings.jsonData.useLiveUpdate;
    this.liveUpdateURI = instanceSettings.jsonData.liveUpdateURI || "ws://localhost:8080/pvws/pv";
    this.aaclient = new AAclient(url, instanceSettings.withCredentials || false);
    this.streamQuery = new StreamQuery(this.aaclient)
  }

  // Called from Grafana panels to get data
  query(options: DataQueryRequest<AAQuery>): Observable<DataQueryResponse> {
    // Remove hidden target from query
    const query = { ...options };
    query.targets = _.filter(query.targets, (t) => !t.hide);

    const query_replaced = this.replaceVariables(query);
    const targets = this.buildQueryParameters(query_replaced);

    // There're no target query
    if (targets.length <= 0) {
      return new Observable<DataQueryResponse>((subscriber) => {
        subscriber.next({ data: [] });
      });
    }

    const stream = _.filter(targets, (t) => t.stream);

    // No stream query
    if (stream.length === 0 || !options.rangeRaw || options.rangeRaw.to !== 'now') {
      if (this.useBackend) {
        return super.query(query_replaced);
      }
      return from(doQuery(this.aaclient, targets));
    }

    // Stream query
    return this.streamQuery.runStream(targets, stream, options.intervalMs);
  }

  // Called from Grafana data source configuration page to make sure the connection is working
  async testDatasource() {
    return this.aaclient.testDatasource();
  }

  pvNamesFindQuery(query: string | undefined | null, maxPvs: number) {
    return this.aaclient.pvNamesFindQuery(query, maxPvs);
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
    const parsedPVs = parseTargetPV(pvQuery);

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

    const pvnamesPromise = _.map(parsedPVs, (targetQuery) => this.pvNamesFindQuery(targetQuery, limitNum));

    return Promise.all(pvnamesPromise).then((pvnamesArray) => {
      const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, limitNum);
      return _.map(pvnames, (pvname) => ({ text: pvname }));
    });
  }

  replaceVariables(options: DataQueryRequest<AAQuery>) {
    const templateSrv = getTemplateSrv();
    const query = { ...options };

    query.targets = _.map(query.targets, (target) => {
      const t = { ...target };

      t.functions = _.map(target.functions, (func) => {
        const newFunc = func;
        newFunc.params = _.map(newFunc.params, (param) => templateSrv.replace(param, query.scopedVars, 'regex'));
        return newFunc;
      });
      t.target = templateSrv.replace(target.target, query.scopedVars, 'regex');
      t.alias = templateSrv.replace(target.alias, query.scopedVars, 'regex');
      t.operator = templateSrv.replace(target.operator, query.scopedVars, 'regex');
      t.strmInt = templateSrv.replace(target.strmInt, query.scopedVars, 'regex');
      t.strmCap = templateSrv.replace(target.strmCap, query.scopedVars, 'regex');

      return t;
    });

    return query;
  }

  buildQueryParameters(options: DataQueryRequest<AAQuery>) {
    const query = { ...options };

    // remove placeholder targets and undefined targets
    query.targets = _.filter(query.targets, (target) => target.target !== '' && typeof target.target !== 'undefined');

    if (query.targets.length <= 0) {
      return [];
    }

    const from = new Date(String(query.range.from));
    const to_ = new Date(String(query.range.to));
    const rangeMsec = to_.getTime() - from.getTime();

    // If "from" == "to" in seconds then "to" should be "to + 1 second"
    const to = rangeMsec >= 1 ? to_ : new Date(to_.getTime() + 1000);

    const maxDataPoints = query.maxDataPoints || 2000;
    const intervalSec = _.floor(rangeMsec / (maxDataPoints * 1000));

    const targets: TargetQuery[] = _.map(query.targets, (target) => {
      const options = getOptions(target.functions);
      const interval = intervalSec >= 1 ? String(intervalSec) : options.disableAutoRaw === 'true' ? '1' : '';
      const operator = target.operator || this.defaultOperator || "mean";

      return {
        //target: templateSrv.replace(target.target, query.scopedVars, 'regex'),
        target: target.target,
        refId: target.refId,
        hide: target.hide,
        alias: target.alias,
        operator,
        stream: target.stream,
        strmInt: target.strmInt,
        strmCap: target.strmCap,
        functions: target.functions,
        regex: target.regex,
        live: target.live,
        aliasPattern: target.aliasPattern,
        options,
        from,
        to,
        interval,
      };
    });

    return targets;
  }
}
