import _ from 'lodash';

import { getBackendSrv } from '@grafana/runtime';
import { lastValueFrom } from 'rxjs';

import { operatorList, TargetQuery, AADataQueryData } from 'types';
import { parseTargetPV } from 'pvnameParser';

export class AAclient {
  url: string;
  withCredentials: boolean;
  headers: { [key: string]: string };

  constructor(url: string, withCredentials: boolean) {
    this.url = url;
    this.withCredentials = withCredentials;
    this.headers = { 'Content-Type': 'application/json' };
  }

  async pvNamesFindQuery(query: string | undefined | null, maxPvs: number) {
    if (!query) {
      return Promise.resolve([]);
    }

    const url = `${this.url}/bpl/getMatchingPVs?limit=${maxPvs}&regex=${encodeURIComponent(query)}`;
    const options = this.makeRequestOption(url);
    const responseObservable = getBackendSrv().fetch<string[]>(options);

    const response = await lastValueFrom(responseObservable);
    return response.data;
  }

  async testDatasource() {
    const url = `${this.url}/bpl/getVersion`;
    const options = this.makeRequestOption(url);
    const responseObservable = getBackendSrv().fetch<string>(options);

    const response = await lastValueFrom(responseObservable);

    if (response.status === 200) {
      return { status: 'success', message: 'Data source is working' };
    }

    return {
      status: 'error',
      message: response.data,
    };
  }

  buildUrls(target: TargetQuery): Promise<string[]> {
    // Get Option values
    const maxNumPVs = Number(target.options.maxNumPVs) || 100;
    const binInterval = target.options.binInterval || target.interval;

    const targetPVs = parseTargetPV(target.target);

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
              this.buildUrl(this.url, pvname, target.operator, binInterval, target.from, target.to)
            );
          } catch (e) {
            reject(e);
          }

          resolve(urls);
        })
    );
  }

  createUrlRequests(urlsArray: string[][]) {
    const requestHash: { [key: string]: Promise<any> } = {};

    const requestsArray = _.map(urlsArray, (urls) => {
      const requests = _.map(urls, (url) => {
        if (!(url in requestHash)) {
          const options = this.makeRequestOption(url);
          const response = getBackendSrv().fetch<AADataQueryData[]>(options);
          requestHash[url] = lastValueFrom(response);
        }
        return requestHash[url];
      });
      return requests;
    });

    return requestsArray;
  }

  private makeRequestOption(url: string) {
    return { method: 'GET', url, headers: this.headers, withCredentials: this.withCredentials };
  }

  private buildUrl(baseUrl: string, pvname: string, operator: string, interval: string, from: Date, to: Date) {
    const pv = (() => {
      // raw Operator or last Operator or interval is less than 1 sec
      if (operator === 'raw' || operator === 'last' || interval === '') {
        return `${pvname}`;
      }

      // Operator is usually provided even if the user doesn't provide it because of the default operator
      // This code maintains compatibility with older versions
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

    const url = `${baseUrl}/data/getData.qw?pv=${encodeURIComponent(pv)}&from=${from_str}&to=${to.toISOString()}`;

    return url;
  }
}
