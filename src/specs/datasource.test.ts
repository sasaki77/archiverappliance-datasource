import range from 'lodash/range';
import split from 'lodash/split';
import { MutableDataFrame, getFieldDisplayName, DataSourceInstanceSettings, DataQueryRequest } from '@grafana/data';
import { DataSource } from '../DataSource';
import { AADataSourceOptions, TargetQuery, AAQuery } from 'types';

const datasourceRequestMock = jest.fn().mockResolvedValue(createDefaultResponse());

jest.mock(
  '@grafana/runtime',
  () => ({
    getBackendSrv: () => ({
      datasourceRequest: datasourceRequestMock,
    }),
    getTemplateSrv: () => ({
      replace: jest.fn().mockImplementation(query => query),
    }),
  }),
  { virtual: true }
);

beforeEach(() => {
  datasourceRequestMock.mockClear();
});

function createDefaultResponse() {
  return {
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
  };
}

describe('Archiverappliance Datasource', () => {
  let ds: DataSource;

  beforeEach(() => {
    const instanceSettings = ({
      url: 'url_header:',
    } as unknown) as DataSourceInstanceSettings<AADataSourceOptions>;
    ds = new DataSource(instanceSettings);
  });

  describe('Build URL tests', () => {
    it('should return an valid url', done => {
      const target = ({
        target: 'PV1',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url[0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an valid multi urls', done => {
      datasourceRequestMock.mockImplementation(requesty =>
        Promise.resolve({
          data: ['PV1', 'PV2'],
        })
      );
      const target = ({
        target: 'PV*',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        regex: true,
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url[0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[1]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV2)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an valid unique urls', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: ['PV1', 'PV2', 'PV1'],
        })
      );

      const target = ({
        target: 'PV*',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        regex: true,
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url[0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[1]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV2)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an 100 urls', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: range(1000).map(num => String(num)),
        })
      );

      const target = ({
        target: 'PV*',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        regex: true,
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url).toHaveLength(100);
        done();
      });
    });

    it('should return an required number of urls', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: range(1000).map(num => String(num)),
        })
      );

      const target = ({
        target: 'PV*',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        regex: true,
        options: { maxNumPVs: 300 },
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url).toHaveLength(300);
        done();
      });
    });

    it('should return an required bin interval url', done => {
      const target = ({
        target: 'PV1',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: { binInterval: 100 },
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url[0]).toBe(
          'url_header:/data/getData.json?pv=mean_100(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an valid multi urls when regex OR target', done => {
      const target = ({
        target: 'PV(A|B|C):(1|2):test',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url).toHaveLength(6);
        expect(url[0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PVA%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[1]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PVA%3A2%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[2]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PVB%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[3]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PVB%3A2%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[4]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PVC%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[5]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PVC%3A2%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an Error when invalid data processing is required', done => {
      const target = ({
        target: 'PV1',
        operator: 'invalid',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target)
        .then(() => {})
        .catch(() => {
          done();
        });
    });

    it('should return an data processing url', done => {
      const from = new Date('2010-01-01T00:00:00.000Z');
      const to = new Date('2010-01-01T00:00:30.000Z');
      const options = {};
      const targets = [
        ({ target: 'PV1', operator: 'mean', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV2', operator: 'raw', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV3', operator: '', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV4', interval: '9', from, to, options } as unknown) as TargetQuery,
      ];

      const urlProcs = targets.map(target => ds.buildUrls(target));

      Promise.all(urlProcs).then(urls => {
        expect(urls).toHaveLength(4);
        expect(urls[0][0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[1][0]).toBe(
          'url_header:/data/getData.json?pv=PV2&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[2][0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV3)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[3][0]).toBe(
          'url_header:/data/getData.json?pv=mean_9(PV4)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });
  });

  describe('Build query parameters tests', () => {
    it('should return valid interval time in integer', done => {
      const options = ({
        targets: [{ target: 'PV1', refId: 'A' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T01:00:01.000Z') },
        maxDataPoints: 1800,
      } as unknown) as DataQueryRequest<AAQuery>;

      const targets = ds.buildQueryParameters(options);

      expect(targets).toHaveLength(1);
      expect(targets[0].interval).toBe('2');
      done();
    });

    it('should return no interval data when interval time is less than 1 second', done => {
      const options = ({
        targets: [{ target: 'PV1', refId: 'A' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const targets = ds.buildQueryParameters(options);

      expect(targets).toHaveLength(1);
      expect(targets[0].interval).toBe('');
      done();
    });

    it('should return filtered array when target is empty or undefined', done => {
      const options = ({
        targets: [
          { target: 'PV', refId: 'A' },
          { target: '', refId: 'B' },
          { target: undefined, refId: 'C' },
        ],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const targets = ds.buildQueryParameters(options);

      expect(targets).toHaveLength(1);
      done();
    });
  });

  describe('Query tests', () => {
    it('should return an empty array when no targets are set', done => {
      ds.query(({ targets: [] } as unknown) as DataQueryRequest<AAQuery>).then((result: any) => {
        expect(result.data).toHaveLength(0);
        done();
      });
    });

    it('should return the server results when a target is set', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
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

      const query = ({
        targets: [{ target: 'PV', refId: 'A' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).then((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();
        const displayName = getFieldDisplayName(dataFrame.fields[1], dataFrame);

        expect(displayName).toBe('PV');
        expect(valArray).toHaveLength(3);
        expect(timesArray).toHaveLength(3);
        expect(valArray[0]).toBe(0);
        expect(timesArray[0]).toBe(1262304000123);
        done();
      });
    });

    it('should return the server results with alias', done => {
      datasourceRequestMock.mockImplementation(request => {
        const pv = request.url.slice(33, 36);
        Promise.resolve({
          status: 'success',
          data: { data: ['value1', 'value2', 'value3'] },
        });
        return Promise.resolve({
          data: [
            {
              meta: { name: pv, PREC: '0' },
              data: [],
            },
          ],
        });
      });

      const query = ({
        targets: [
          { target: 'PV1', refId: 'A', alias: 'alias' },
          { target: 'PV2', refId: 'B', alias: '' },
          { target: 'PV3', refId: 'C', alias: undefined },
          { target: 'PV4', refId: 'D' },
        ],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).then((result: any) => {
        expect(result.data).toHaveLength(4);
        const dataFrameArray: MutableDataFrame[] = result.data;
        const pv1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
        const pv2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
        const pv3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);
        const pv4 = getFieldDisplayName(dataFrameArray[3].fields[1], dataFrameArray[3]);

        expect(pv1).toBe('alias');
        expect(pv2).toBe('PV2');
        expect(pv3).toBe('PV3');
        expect(pv4).toBe('PV4');
        done();
      });
    });

    it('should return the server results with alias pattern', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'header:PV1', PREC: '0' },
              data: [],
            },
          ],
        })
      );

      const query = ({
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
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).then((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const alias = getFieldDisplayName(dataFrame.fields[1], dataFrame);
        expect(alias).toBe('PV1:header');
        done();
      });
    });
  });

  describe('PV name find query tests', () => {
    it('should return the pv name results when a target is null', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery(null, 100).then((result: any) => {
        expect(result).toHaveLength(0);
        done();
      });
    });

    it('should return the pv name results when a target is undefined', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery(undefined, 100).then((result: any) => {
        expect(result).toHaveLength(0);
        done();
      });
    });

    it('should return the pv name results when a target is empty', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery('', 100).then((result: any) => {
        expect(result).toHaveLength(0);
        done();
      });
    });

    it('should return the pv name results when a target is set', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery('metric', 100).then((result: any) => {
        expect(result).toHaveLength(3);
        expect(result[0]).toBe('metric_0');
        expect(result[1]).toBe('metric_1');
        expect(result[2]).toBe('metric_2');
        done();
      });
    });
  });

  describe('Metric find query tests', () => {
    it('should return the pv name results for metricFindQuery', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          _request: request,
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.metricFindQuery('metric').then((result: any) => {
        expect(result).toHaveLength(3);
        expect(result[0].text).toBe('metric_0');
        expect(result[1].text).toBe('metric_1');
        expect(result[2].text).toBe('metric_2');
        done();
      });
    });

    it('should return the pv name results for metricFindQuery with regex OR', done => {
      datasourceRequestMock.mockImplementation(request =>
        Promise.resolve({
          _request: request,
          data: [unescape(split(request.url, /regex=(.*)/)[1])],
        })
      );

      ds.metricFindQuery('PV(A|B|C):(1|2):test').then((result: any) => {
        expect(result).toHaveLength(6);
        expect(result[0].text).toBe('PVA:1:test');
        expect(result[1].text).toBe('PVA:2:test');
        expect(result[2].text).toBe('PVB:1:test');
        expect(result[3].text).toBe('PVB:2:test');
        expect(result[4].text).toBe('PVC:1:test');
        expect(result[5].text).toBe('PVC:2:test');
        done();
      });
    });

    it('should return the pv name results for metricFindQuery with limit parameter', done => {
      datasourceRequestMock.mockImplementation(request => {
        const params = new URLSearchParams(request.url.split('?')[1]);
        const limit = parseInt(String(params.get('limit')), 10);
        const pvname = params.get('regex');
        const data = [...Array(limit).keys()].map(i => `${pvname}${i}`);

        return Promise.resolve({
          _request: request,
          data,
        });
      });

      ds.metricFindQuery('PV?limit=5').then((result: any) => {
        expect(result).toHaveLength(5);
        expect(result[0].text).toBe('PV0');
        expect(result[1].text).toBe('PV1');
        expect(result[2].text).toBe('PV2');
        expect(result[3].text).toBe('PV3');
        expect(result[4].text).toBe('PV4');
        done();
      });
    });

    it('should return the pv name results for metricFindQuery with invalid limit parameter', done => {
      datasourceRequestMock.mockImplementation(request => {
        const params = new URLSearchParams(request.url.split('?')[1]);
        const limit = parseInt(String(params.get('limit')), 10);
        const pvname = params.get('regex');
        const data = [...Array(limit).keys()].map(i => `${pvname}${i}`);

        return Promise.resolve({
          _request: request,
          data,
        });
      });

      ds.metricFindQuery('PV?limit=a').then((result: any) => {
        expect(result).toHaveLength(100);
        done();
      });
    });
  });
});
