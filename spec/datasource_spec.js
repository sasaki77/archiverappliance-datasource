/* eslint import/no-unresolved: [0] */

import _ from 'lodash';
import { expect } from 'chai';
import { ArchiverapplianceDatasource } from '../datasource';

describe('ArchiverapplianceDatasource', () => {
  const ctx = {};

  beforeEach(() => {
    ctx.instanceSettings = {
      url: 'url_header:',
    };
    ctx.backendSrv = {};
    ctx.templateSrv = {};
    ctx.ds = new ArchiverapplianceDatasource(
      ctx.instanceSettings,
      ctx.backendSrv,
      ctx.templateSrv,
    );
  });

  it('should return an empty array when no targets are set', (done) => {
    ctx.ds.query({ targets: [] }).then((result) => {
      expect(result.data).to.have.length(0);
      done();
    });
  });

  it('should return an valid url', (done) => {
    const target = {
      target: 'PV1',
      interval: '9',
      from: new Date('2010-01-01T00:00:00.000Z'),
      to: new Date('2010-01-01T00:00:30.000Z'),
    };

    ctx.ds.buildUrls(target).then((url) => {
      expect(url[0]).to.equal('url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      done();
    });
  });

  it('should return an valid multi urls', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: ['PV1', 'PV2'],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    const target = {
      target: 'PV*',
      interval: '9',
      from: new Date('2010-01-01T00:00:00.000Z'),
      to: new Date('2010-01-01T00:00:30.000Z'),
      regex: true,
    };

    ctx.ds.buildUrls(target).then((url) => {
      expect(url[0]).to.equal('url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[1]).to.equal('url_header:/data/getData.json?pv=mean_9(PV2)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      done();
    });
  });

  it('should return an valid unique urls', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: ['PV1', 'PV2', 'PV1'],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    const target = {
      target: 'PV*',
      interval: '9',
      from: new Date('2010-01-01T00:00:00.000Z'),
      to: new Date('2010-01-01T00:00:30.000Z'),
      regex: true,
    };

    ctx.ds.buildUrls(target).then((url) => {
      expect(url[0]).to.equal('url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[1]).to.equal('url_header:/data/getData.json?pv=mean_9(PV2)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      done();
    });
  });

  it('should return an 100 urls', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: _.map(_.range(1000), (num) => String(num)),
      })
    );

    ctx.templateSrv.replace = (data) => data;

    const target = {
      target: 'PV*',
      interval: '9',
      from: new Date('2010-01-01T00:00:00.000Z'),
      to: new Date('2010-01-01T00:00:30.000Z'),
      regex: true,
    };

    ctx.ds.buildUrls(target).then((url) => {
      expect(url).to.have.length(100);
      done();
    });
  });

  it('should return an valid multi urls when regex OR target', (done) => {
    ctx.templateSrv.replace = (data) => data;

    const target = {
      target: 'PV(A|B|C):(1|2):test',
      interval: '9',
      from: new Date('2010-01-01T00:00:00.000Z'),
      to: new Date('2010-01-01T00:00:30.000Z'),
    };

    ctx.ds.buildUrls(target).then((url) => {
      expect(url).to.have.length(6);
      expect(url[0]).to.equal('url_header:/data/getData.json?pv=mean_9(PVA%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[1]).to.equal('url_header:/data/getData.json?pv=mean_9(PVA%3A2%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[2]).to.equal('url_header:/data/getData.json?pv=mean_9(PVB%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[3]).to.equal('url_header:/data/getData.json?pv=mean_9(PVB%3A2%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[4]).to.equal('url_header:/data/getData.json?pv=mean_9(PVC%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(url[5]).to.equal('url_header:/data/getData.json?pv=mean_9(PVC%3A2%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      done();
    });
  });

  it('should return an Error when invalid data processing is required', (done) => {
    const target = {
      target: 'PV1',
      operator: 'invalid',
      interval: '9',
      from: new Date('2010-01-01T00:00:00.000Z'),
      to: new Date('2010-01-01T00:00:30.000Z'),
    };

    ctx.ds.buildUrls(target).then(() => {
    }).catch(() => {
      done();
    });
  });

  it('should return an data processing url', (done) => {
    const from = new Date('2010-01-01T00:00:00.000Z');
    const to = new Date('2010-01-01T00:00:30.000Z');
    const targets = [
      { target: 'PV1', operator: 'mean', interval: '9', from, to },
      { target: 'PV2', operator: 'raw', interval: '9', from, to },
      { target: 'PV3', operator: '', interval: '9', from, to },
      { target: 'PV4', interval: '9', from, to },
    ];

    const urlProcs = targets.map((target) => ctx.ds.buildUrls(target));

    Promise.all(urlProcs).then((urls) => {
      expect(urls).to.have.length(4);
      expect(urls[0][0]).to.equal('url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(urls[1][0]).to.equal('url_header:/data/getData.json?pv=PV2&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(urls[2][0]).to.equal('url_header:/data/getData.json?pv=mean_9(PV3)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      expect(urls[3][0]).to.equal('url_header:/data/getData.json?pv=mean_9(PV4)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z');
      done();
    });
  });

  it('should return valid interval time in integer', (done) => {
    ctx.templateSrv.replace = (data) => data;

    const options = {
      targets: [{ target: 'PV1', refId: 'A' }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T01:00:01.000Z') },
      maxDataPoints: 1800,
    };

    const query = ctx.ds.buildQueryParameters(options);

    expect(query.targets).to.have.length(1);
    expect(query.targets[0].interval).to.equal('2');
    done();
  });

  it('should return no interval data when interval time is less than 1 second', (done) => {
    ctx.templateSrv.replace = (data) => data;

    const options = {
      targets: [{ target: 'PV1', refId: 'A' }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    };

    const query = ctx.ds.buildQueryParameters(options);

    expect(query.targets).to.have.length(1);
    expect(query.targets[0].interval).to.equal('');
    done();
  });

  it('should return filtered array when target is empty or undefined', (done) => {
    ctx.templateSrv.replace = (data) => data;

    const options = {
      targets: [
        { target: 'PV', refId: 'A' },
        { target: '', refId: 'B' },
        { target: undefined, refId: 'C' },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    };

    const query = ctx.ds.buildQueryParameters(options);

    expect(query.targets).to.have.length(1);
    done();
  });

  it('should return the server results when a target is set', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: 'PV', PREC: '0' },
            data: [
              { secs: 1262304000, val: 0, nanos: 123000000, severity: 0, status: 0 },
              { secs: 1262304001, val: 1, nanos: 456000000, severity: 0, status: 0 },
              { secs: 1262304002, val: 2, nanos: 789000000, severity: 0, status: 0 },
            ],
          },
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    const query = {
      targets: [{ target: 'PV', refId: 'A' }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then((result) => {
      expect(result.data).to.have.length(1);
      const series = result.data[0];
      expect(series.target).to.equal('PV');
      expect(series.datapoints).to.have.length(3);
      expect(series.datapoints[0][0]).to.equal(0);
      expect(series.datapoints[0][1]).to.equal(1262304000123);
      done();
    });
  });

  it('should return the server results with alias', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => {
      const pv = request.url.slice(33, 36);
      return Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: pv, PREC: '0' },
            data: [],
          },
        ],
      });
    };

    ctx.templateSrv.replace = (data) => data;

    const query = {
      targets: [
        { target: 'PV1', refId: 'A', alias: 'alias' },
        { target: 'PV2', refId: 'B', alias: '' },
        { target: 'PV3', refId: 'C', alias: undefined },
        { target: 'PV4', refId: 'D' },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then((result) => {
      expect(result.data).to.have.length(4);
      expect(result.data[0].target).to.equal('alias');
      expect(result.data[1].target).to.equal('PV2');
      expect(result.data[2].target).to.equal('PV3');
      expect(result.data[3].target).to.equal('PV4');
      done();
    });
  });

  it('should return the server results with alias pattern', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: 'header:PV1', PREC: '0' },
            data: [],
          },
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    const query = {
      targets: [
        {
          target: 'header:PV1',
          refId: 'A',
          alias: '$2:$1',
          aliasPattern: '(.*):(.*)',
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then((result) => {
      expect(result.data).to.have.length(1);
      expect(result.data[0].target).to.equal('PV1:header');
      done();
    });
  });

  it('should return the pv name results when a target is null', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          'metric_0',
          'metric_1',
          'metric_2',
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    ctx.ds.pvNamesFindQuery(null).then((result) => {
      expect(result).to.have.length(0);
      done();
    });
  });

  it('should return the pv name results when a target is undefined', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          'metric_0',
          'metric_1',
          'metric_2',
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    ctx.ds.pvNamesFindQuery(undefined).then((result) => {
      expect(result).to.have.length(0);
      done();
    });
  });

  it('should return the pv name results when a target is empty', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          'metric_0',
          'metric_1',
          'metric_2',
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    ctx.ds.pvNamesFindQuery('').then((result) => {
      expect(result).to.have.length(0);
      done();
    });
  });

  it('should return the pv name results when a target is set', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          'metric_0',
          'metric_1',
          'metric_2',
        ],
      })
    );

    ctx.ds.pvNamesFindQuery('metric').then((result) => {
      expect(result).to.have.length(3);
      expect(result[0]).to.equal('metric_0');
      expect(result[1]).to.equal('metric_1');
      expect(result[2]).to.equal('metric_2');
      done();
    });
  });

  it('should return the pv name results for metricFindQuery', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          'metric_0',
          'metric_1',
          'metric_2',
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    ctx.ds.metricFindQuery('metric').then((result) => {
      expect(result).to.have.length(3);
      expect(result[0].text).to.equal('metric_0');
      expect(result[1].text).to.equal('metric_1');
      expect(result[2].text).to.equal('metric_2');
      done();
    });
  });

  it('should return the pv name results for metricFindQuery with regex OR', (done) => {
    ctx.backendSrv.datasourceRequest = (request) => (
      Promise.resolve({
        _request: request,
        data: [
          unescape(_.split(request.url, /regex=(.*)/)[1]),
        ],
      })
    );

    ctx.templateSrv.replace = (data) => data;

    ctx.ds.metricFindQuery('PV(A|B|C):(1|2):test').then((result) => {
      expect(result).to.have.length(6);
      expect(result[0].text).to.equal('PVA:1:test');
      expect(result[1].text).to.equal('PVA:2:test');
      expect(result[2].text).to.equal('PVB:1:test');
      expect(result[3].text).to.equal('PVB:2:test');
      expect(result[4].text).to.equal('PVC:1:test');
      expect(result[5].text).to.equal('PVC:2:test');
      done();
    });
  });
});
