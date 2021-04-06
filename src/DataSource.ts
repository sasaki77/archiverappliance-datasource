import _ from 'lodash';
import { Observable, Subscriber, from } from 'rxjs';
import { v4 as uuidv4 } from 'uuid';
import ms from 'ms';
import { DataSourceWithBackend, getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import {
  CircularDataFrame,
  DataQueryResponse,
  DataQueryRequest,
  DataSourceInstanceSettings,
  MutableDataFrame,
  FieldType,
  getFieldDisplayName,
  LoadingState,
} from '@grafana/data';
import {
  AAQuery,
  AADataSourceOptions,
  TargetQuery,
  AADataQueryData,
  AADataQueryResponse,
  operatorList,
  isNumberArray,
} from './types';
import { applyFunctionDefs, getOptions, getToScalarFuncs } from './aafunc';
import { locateOuterParen, permuteQuery, splitLowestLevelOnly, selectiveInsert } from './utils';

export class DataSource extends DataSourceWithBackend<AAQuery, AADataSourceOptions> {
  url?: string | undefined;
  name: string;
  useBackend?: boolean | undefined;
  withCredentials?: boolean;
  headers: { [key: string]: string };
  timerIDs: { [key: string]: any };

  constructor(instanceSettings: DataSourceInstanceSettings<AADataSourceOptions>) {
    super(instanceSettings);
    this.url = instanceSettings.url;
    this.name = instanceSettings.name;
    this.useBackend = instanceSettings.jsonData.useBackend;
    this.withCredentials = instanceSettings.withCredentials;
    this.headers = { 'Content-Type': 'application/json' };
    this.timerIDs = {};
    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
      this.headers.Authorization = instanceSettings.basicAuth;
    }
  }

  // Called from Grafana panels to get data
  query(options: DataQueryRequest<AAQuery>): Observable<DataQueryResponse> {
    const rawTargets = this.buildQueryParameters(options);

    // Remove hidden target from query
    const targets = _.filter(rawTargets, (t) => !t.hide);

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
        return super.query(options);
      }
      return from(this.doQuery(targets));
    }

    // Stream query
    return new Observable<DataQueryResponse>((subscriber) => {
      // Create new targets to disable auto Extrapolation
      const t = _.map(targets, (target) => {
        return {
          ...target,
          options: {
            ...target.options,
            disableExtrapol: 'true',
          },
        };
      });

      const id = uuidv4();
      const cirFrames: { [key: string]: CircularDataFrame<any> } = {};

      this.doQueryStream(t, cirFrames).then((data) => {
        subscriber.next(data);

        const interval = (stream[0].strmInt && ms(stream[0].strmInt)) || options.intervalMs;

        // Create new targets to update interval time
        const new_t = _.map(t, (target) => {
          const t_int = target.interval ? Math.floor(interval / 1000).toFixed() : '';
          const int = interval >= 1000 ? t_int : '';

          return {
            ...target,
            interval: int,
          };
        });

        this.timerIDs[id] = setTimeout(this.timerLoop, interval, subscriber, new_t, id, cirFrames, interval);
      });

      return () => {
        this.timerClear(id);
      };
    });
  }

  timerLoop = async (
    subscriber: Subscriber<DataQueryResponse>,
    targets: TargetQuery[],
    id: string,
    frames: { [key: string]: CircularDataFrame },
    interval: number
  ) => {
    this.updateTargetDate(targets);
    const data = await this.doQueryStream(targets, frames);

    subscriber.next(data);
    if (id in this.timerIDs) {
      this.timerIDs[id] = setTimeout(this.timerLoop, interval, subscriber, targets, id, frames, interval);
    }
  };

  updateTargetDate(targets: TargetQuery[]) {
    return _.map(targets, (target) => {
      // AA should probably not able to return latest data near the "now".
      // So, the time range is set from 2 secs ago from last update date and to 500 msecs ago from "now".
      target.from = new Date(target.to.getTime() - 2000);
      target.to = new Date(Date.now() - 500);
      return target;
    });
  }

  timerClear(id: string) {
    if (id in this.timerIDs) {
      clearTimeout(this.timerIDs[id]);
    }
    this.timerIDs = {};
  }

  doQuery(targets: TargetQuery[]): Promise<{ data: Array<MutableDataFrame<any>> }> {
    // Create promises to buil URLs for each targets: [[URLs for target 1], [URLs for target 2] , ...]
    const urlsArray = _.map(targets, (target) => this.buildUrls(target));

    // Wait for building URLs then create target data
    const targetProcesses = Promise.all(urlsArray).then((urlsArray) => {
      // Create promises to retrieve data for each targets: [[Responses for target 1], [Reponses for target 2] , ...]
      const responsePromisesArray = this.createUrlRequests(urlsArray);

      // Data processing for each targets: [[Processed data for target 1], [Processed data for target 2], ...]
      const targetProcesses = _.map(responsePromisesArray, (responsePromises, i) => {
        return Promise.all(responsePromises).then((responses) => this.targetProcess(responses, targets[i]));
      });

      // Wait all target data processings
      return Promise.all(targetProcesses);
    });

    return targetProcesses.then((dataFramesArray) => this.postProcess(dataFramesArray));
  }

  doQueryStream(targets: TargetQuery[], frames: { [key: string]: CircularDataFrame }): Promise<DataQueryResponse> {
    // Create promises to buil URLs for each targets: [[URLs for target 1], [URLs for target 2] , ...]
    const urlsArray = _.map(targets, (target) => this.buildUrls(target));

    // Wait for building URLs then create target data
    const targetProcesses = Promise.all(urlsArray).then((urlsArray) => {
      // Create promises to retrieve data for each targets: [[Responses for target 1], [Reponses for target 2] , ...]
      const responsePromisesArray = this.createUrlRequests(urlsArray);

      // Data processing for each targets: [[Processed data for target 1], [Processed data for target 2], ...]
      const targetProcesses = _.map(responsePromisesArray, (responsePromises, i) => {
        return Promise.all(responsePromises)
          .then((responses) => this.responseParse(responses, targets[i]))
          .then((dataFrames) => this.mergeResToCirFrames(dataFrames, frames, targets[i]))
          .then((dataFrames) => this.setAlias(dataFrames, targets[i]))
          .then((dataFrames) => this.applyFunctions(dataFrames, targets[i]));
      });

      // Wait all target data processings
      return Promise.all(targetProcesses);
    });

    return targetProcesses.then((dataFramesArray) => this.streamPostProcess(dataFramesArray));
  }

  mergeResToCirFrames(
    dataFrames: MutableDataFrame[],
    cirFrames: { [key: string]: CircularDataFrame },
    target: TargetQuery
  ): Promise<MutableDataFrame[]> {
    const to = target.to.getTime();
    const d = _.filter(dataFrames, (frame) => frame.name !== undefined);

    const frames = _.map(d, (frame) => {
      if (frame.name === undefined) {
        return frame;
      }

      // Create frame for new arrival data
      if (!(frame.name in cirFrames)) {
        cirFrames[frame.name] = this.createStreamFrame(target, frame);
        return cirFrames[frame.name];
      }

      const last_time = cirFrames[frame.name].get(cirFrames[frame.name].length - 1)['time'];

      // Update frame data
      for (let i = 0; i < frame.length; i++) {
        const fields = frame.get(i);
        if (fields['time'] <= last_time || fields['time'] > to) {
          continue;
        }
        cirFrames[frame.name].add(fields);
      }
      return cirFrames[frame.name];
    });

    return Promise.resolve(frames);
  }

  createStreamFrame(target: TargetQuery, dataFrame: MutableDataFrame) {
    const c = parseInt(target.strmCap, 10);
    const cap = dataFrame.refId ? c || dataFrame.length : dataFrame.length;

    const new_frame = new CircularDataFrame({
      append: 'tail',
      capacity: cap,
    });

    new_frame.name = dataFrame.name;
    for (const field of dataFrame.fields) {
      new_frame.addField(field);
    }

    return new_frame;
  }

  targetProcess(responses: any, target: TargetQuery) {
    return this.responseParse(responses, target)
      .then((dataFrames) => this.setAlias(dataFrames, target))
      .then((dataFrames) => this.applyFunctions(dataFrames, target));
  }

  postProcess(dataFramesArray: MutableDataFrame[][]) {
    const dataFrames = _.flatten(dataFramesArray);

    return { data: dataFrames };
  }

  streamPostProcess(dataFramesArray: MutableDataFrame[][]) {
    const dataFrames = _.flatten(dataFramesArray);

    return { data: dataFrames, state: LoadingState.Streaming };
  }

  buildUrls(target: TargetQuery): Promise<string[]> {
    // Get Option values
    const maxNumPVs = Number(target.options.maxNumPVs) || 100;
    const binInterval = target.options.binInterval || target.interval;

    const targetPVs = this.parseTargetPV(target.target);

    // Create Promise to fetch PV names
    const pvnamesPromise = _.map(targetPVs, (targetPV) => {
      if (target.regex) {
        return this.pvNamesFindQuery(targetPV, maxNumPVs);
      }

      return Promise.resolve([targetPV]);
    });

    return Promise.all(pvnamesPromise).then(
      (pvnamesArray) =>
        new Promise((resolve, reject) => {
          const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, maxNumPVs);
          let urls: string[] = [];

          try {
            urls = _.map(pvnames, (pvname) =>
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
      // raw Operator or last Operator or interval is less than 1 sec
      if (operator === 'raw' || operator === 'last' || interval === '') {
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

    const from_str = operator === 'last' ? to.toISOString() : from.toISOString();

    const url = `${this.url}/data/getData.qw?pv=${encodeURIComponent(pv)}&from=${from_str}&to=${to.toISOString()}`;

    return url;
  }

  createUrlRequests(urlsArray: string[][]) {
    const requestHash: { [key: string]: Promise<any> } = {};

    const requestsArray = _.map(urlsArray, (urls) => {
      const requests = _.map(urls, (url) => {
        if (!(url in requestHash)) {
          requestHash[url] = this.doRequest({ url, method: 'GET' });
        }
        return requestHash[url];
      });
      return requests;
    });

    return requestsArray;
  }

  responseParse(responses: AADataQueryResponse[], target: TargetQuery) {
    const dataFramesArray = _.map(responses, (response) => {
      const dataFrames = _.map(response.data, (targetRes) => {
        if (targetRes.meta.waveform) {
          const toScalarFuncs = getToScalarFuncs(target.functions);
          if (toScalarFuncs.length > 0) {
            return this.parseArrayResponseToScalar(targetRes, toScalarFuncs, target);
          }
          return this.parseArrayResponse(targetRes, target);
        }
        return this.parseScalarResponse(targetRes, target);
      });

      return _.flatten(dataFrames);
    });

    const dataFrames = _.flatten(dataFramesArray);

    // Except for raw operator or extrapolation is disabled
    if ((target.operator !== 'raw' && target.interval !== '') || target.options.disableExtrapol === 'true') {
      return Promise.resolve(dataFrames);
    }

    // Extrapolation for raw operator
    const to_msec = target.to.getTime();
    const extrapolationDataFrames = _.map(dataFrames, (dataframe) => {
      const latestval = dataframe.get(dataframe.length - 1);
      const addval = { ...latestval, time: to_msec };

      dataframe.add(addval);

      return dataframe;
    });

    return Promise.resolve(extrapolationDataFrames);
  }

  parseArrayResponse(targetRes: AADataQueryData, target: TargetQuery) {
    // Type check for columnValues
    if (!isNumberArray(targetRes)) {
      return new MutableDataFrame();
    }

    const columnValues = _.map(targetRes.data, (datapoint) => datapoint.val);

    const rowValues = _.unzip(columnValues);
    const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
    const fields = [{ name: 'time', type: FieldType.time, values: times }];

    // Add fields for each waveform elements
    _.reduce(
      rowValues,
      (fields, val, i) => {
        const field = {
          name: `${targetRes.meta.name}[${i}]`,
          type: FieldType.number,
          values: val,
        };
        fields.push(field);
        return fields;
      },
      fields
    );

    const frame = new MutableDataFrame({
      refId: target.refId,
      name: targetRes.meta.name,
      fields,
    });

    return frame;
  }

  parseArrayResponseToScalar(
    targetRes: AADataQueryData,
    toScalarFuncs: Array<{ func: any; label: string }>,
    target: TargetQuery
  ) {
    // Type check for columnValues
    if (!isNumberArray(targetRes)) {
      return new MutableDataFrame();
    }

    const frames = _.map(toScalarFuncs, (func) => {
      const values = _.map(targetRes.data, (datapoint) => func.func(datapoint.val));
      const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
      const frame = new MutableDataFrame({
        refId: target.refId,
        name: targetRes.meta.name,
        fields: [
          { name: 'time', type: FieldType.time, values: times },
          {
            name: 'value',
            type: FieldType.number,
            values: values,
            config: { displayName: `${targetRes.meta.name} (${func.label})` },
          },
        ],
      });
      return frame;
    });

    return frames;
  }

  parseScalarResponse(targetRes: AADataQueryData, target: TargetQuery): MutableDataFrame {
    const values = _.map(targetRes.data, (datapoint) => datapoint.val);
    const times = _.map(targetRes.data, (datapoint) => datapoint.millis);
    const frame = new MutableDataFrame({
      refId: target.refId,
      name: targetRes.meta.name,
      fields: [
        { name: 'time', type: FieldType.time, values: times },
        { name: 'value', type: FieldType.number, values: values, config: { displayName: targetRes.meta.name } },
      ],
    });
    return frame;
  }

  async setAlias(dataFrames: MutableDataFrame[], target: TargetQuery): Promise<MutableDataFrame[]> {
    if (!target.alias) {
      return Promise.resolve(dataFrames);
    }

    let pattern: RegExp;
    if (target.aliasPattern) {
      pattern = new RegExp(target.aliasPattern, '');
    }

    const newDataFrames = _.map(dataFrames, (dataFrame) => {
      const valfields = _.filter(dataFrame.fields, (field) => field.name !== 'time');

      const newValfields = _.map(valfields, (valfield) => {
        const displayName = getFieldDisplayName(valfield, dataFrame);
        const alias = pattern ? displayName.replace(pattern, target.alias) : target.alias;

        return {
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
      });

      return new MutableDataFrame({
        ...dataFrame,
        fields: [dataFrame.fields[0]].concat(newValfields),
      });
    });

    return Promise.resolve(newDataFrames);
  }

  applyFunctions(dataFrames: MutableDataFrame[], target: TargetQuery) {
    if (target.functions === undefined) {
      return Promise.resolve(dataFrames);
    }

    return applyFunctionDefs(target.functions, dataFrames);
  }

  // Called from Grafana data source configuration page to make sure the connection is working
  async testDatasource() {
    return this.doRequest({
      url: `${this.url}/bpl/getVersion`,
      method: 'GET',
    }).then((response) => {
      if (response.status === 200) {
        return { status: 'success', message: 'Data source is working', title: 'Success' };
      }

      return {
        status: 'error',
        title: 'Failed',
        message: response.message,
      };
    });
  }

  pvNamesFindQuery(query: string | undefined | null, maxPvs: number) {
    if (!query) {
      return Promise.resolve([]);
    }

    const url = `${this.url}/bpl/getMatchingPVs?limit=${maxPvs}&regex=${encodeURIComponent(query)}`;

    return this.doRequest({
      url,
      method: 'GET',
    }).then((res) => res.data);
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

    const pvnamesPromise = _.map(parsedPVs, (targetQuery) => this.pvNamesFindQuery(targetQuery, limitNum));

    return Promise.all(pvnamesPromise).then((pvnamesArray) => {
      const pvnames = _.slice(_.uniq(_.flatten(pvnamesArray)), 0, limitNum);
      return _.map(pvnames, (pvname) => ({ text: pvname }));
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
    query.targets = _.filter(query.targets, (target) => target.target !== '' && typeof target.target !== 'undefined');

    if (query.targets.length <= 0) {
      return [];
    }

    const from = new Date(String(query.range.from));
    const to = new Date(String(query.range.to));
    const rangeMsec = to.getTime() - from.getTime();
    const maxDataPoints = query.maxDataPoints || 2000;
    const intervalSec = _.floor(rangeMsec / (maxDataPoints * 1000));

    const targets: TargetQuery[] = _.map(query.targets, (target) => {
      // Replace parameters with variables for each functions
      const functions = _.map(target.functions, (func) => {
        const newFunc = func;
        newFunc.params = _.map(newFunc.params, (param) => templateSrv.replace(param, query.scopedVars, 'regex'));
        return newFunc;
      });

      const options = getOptions(target.functions);
      const interval = intervalSec >= 1 ? String(intervalSec) : options.disableAutoRaw === 'true' ? '1' : '';

      return {
        target: templateSrv.replace(target.target, query.scopedVars, 'regex'),
        refId: target.refId,
        hide: target.hide,
        alias: templateSrv.replace(target.alias, query.scopedVars, 'regex'),
        operator: templateSrv.replace(target.operator, query.scopedVars, 'regex'),
        stream: target.stream,
        strmInt: templateSrv.replace(target.strmInt, query.scopedVars, 'regex'),
        strmCap: templateSrv.replace(target.strmCap, query.scopedVars, 'regex'),
        functions,
        regex: target.regex,
        aliasPattern: target.aliasPattern,
        options,
        from,
        to,
        interval,
      };
    });

    return targets;
  }

  parseTargetPV(targetPV: string): string[] {
    // ex) A(1(2))(3|4)B
    // parenPhraseData = {phrases: [(1(2)), (3|4)], idxs: [[1,6], [7,11]]}
    // phraseParts = [[1(2)], [3,4]]
    // 1stresult = [A1(2)3B, A1(2)4B]
    // 2ndresult = [A123B, A124B]
    const parenPhraseData = locateOuterParen(targetPV);

    // Identify parenthesis-bound sections
    const parenPhrases = parenPhraseData.phrases;

    // If there are no sub-phrases in this string, return immediately to prevent further recursion
    if (parenPhrases.length === 0) {
      return [targetPV];
    }

    // A list of all the possible phrases
    const phraseParts = _.map(parenPhrases, (phrase, i) => {
      const stripedPhrase = phrase.slice(1, phrase.length - 1);
      return splitLowestLevelOnly(stripedPhrase);
    });

    // list of all the configurations for the in-order phrases to be inserted
    const phraseCase = permuteQuery(phraseParts);

    //// Build results by substituting all phrase combinations in place for 1st-level substitutions
    const result = _.map(phraseCase, (phrase) => selectiveInsert(targetPV, parenPhraseData.idxs, phrase));

    // For any phrase that has sub-phrases in need of parsing, call this function again on the sub-phrase and append the results to the end of the current output.
    result.forEach((chunk, pos) => {
      const parseAttempt = this.parseTargetPV(chunk);
      if (parseAttempt.length > 1) {
        result.splice(pos, 1); // pop partially parsed entry
        result.push(...parseAttempt); // add new entires at the end of the list.
      }
    });

    return result;
  }
}
