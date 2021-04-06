import range from 'lodash/range';
import split from 'lodash/split';
import { min, max } from 'lodash';
import {
  MutableDataFrame,
  getFieldDisplayName,
  DataSourceInstanceSettings,
  DataQueryRequest,
  LoadingState,
} from '@grafana/data';
import * as runtime from '@grafana/runtime'
import { DataSource } from '../DataSource';
import { AADataSourceOptions, TargetQuery, AAQuery } from '../types';
import { take, toArray } from 'rxjs/operators';

const datasourceRequestMock = jest.fn().mockResolvedValue(createDefaultResponse());

jest.spyOn(runtime, 'getBackendSrv').mockImplementation(
  () => {
    return {datasourceRequest: datasourceRequestMock} as any as runtime.BackendSrv;
  }
);

jest.spyOn(runtime, 'getTemplateSrv').mockImplementation(
  () => {
      return {replace: jest.fn().mockImplementation((query) => query)} as any as runtime.TemplateSrv;
  }
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
          { millis: 1262304000123, val: 0 },
          { millis: 1262304001456, val: 1 },
          { millis: 1262304002789, val: 2 },
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
      jsonData: {
        useBackend: false
      },
    } as unknown) as DataSourceInstanceSettings<AADataSourceOptions>;
    ds = new DataSource(instanceSettings);
  });

  describe('testDatasource tests', () => {
    it('should return success message', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          status: 200,
          message: 'Success',
        })
      );

      ds.testDatasource().then((result: any) => {
        expect(result.status).toBe('success');
        expect(result.message).toBe('Data source is working');
        expect(result.title).toBe('Success');
        done();
      });
    });

    it('should return error message', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          status: 404,
          message: 'Bad gateway',
        })
      );

      ds.testDatasource().then((result: any) => {
        expect(result.status).toBe('error');
        expect(result.message).toBe('Bad gateway');
        expect(result.title).toBe('Failed');
        done();
      });
    });
  });

  describe('Build URL tests', () => {
    it('should return an valid url', (done) => {
      const target = ({
        target: 'PV1',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url[0]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an valid multi urls', (done) => {
      datasourceRequestMock.mockImplementation((requesty) =>
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
          'url_header:/data/getData.qw?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[1]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PV2)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an valid unique urls', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
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
          'url_header:/data/getData.qw?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[1]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PV2)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an 100 urls', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: range(1000).map((num) => String(num)),
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

    it('should return an required number of urls', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: range(1000).map((num) => String(num)),
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

    it('should return an required bin interval url', (done) => {
      const target = ({
        target: 'PV1',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: { binInterval: 100 },
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url[0]).toBe(
          'url_header:/data/getData.qw?pv=mean_100(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an valid multi urls when regex OR target', (done) => {
      const target = ({
        target: 'PV(A|B):(1|2(3|4)):test',
        interval: '9',
        from: new Date('2010-01-01T00:00:00.000Z'),
        to: new Date('2010-01-01T00:00:30.000Z'),
        options: {},
      } as unknown) as TargetQuery;

      ds.buildUrls(target).then((url: any) => {
        expect(url).toHaveLength(6);
        expect(url[0]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PVA%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[1]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PVB%3A1%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[2]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PVA%3A23%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[3]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PVA%3A24%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[4]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PVB%3A23%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(url[5]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PVB%3A24%3Atest)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });

    it('should return an Error when invalid data processing is required', (done) => {
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

    it('should return an data processing url', (done) => {
      const from = new Date('2010-01-01T00:00:00.000Z');
      const to = new Date('2010-01-01T00:00:30.000Z');
      const options = {};
      const targets = [
        ({ target: 'PV1', operator: 'mean', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV2', operator: 'raw', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV3', operator: '', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV4', interval: '9', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV5', operator: 'max', interval: '', from, to, options } as unknown) as TargetQuery,
        ({ target: 'PV6', operator: 'last', interval: '9', from, to, options } as unknown) as TargetQuery,
      ];

      const urlProcs = targets.map((target) => ds.buildUrls(target));

      Promise.all(urlProcs).then((urls) => {
        expect(urls).toHaveLength(6);
        expect(urls[0][0]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[1][0]).toBe(
          'url_header:/data/getData.qw?pv=PV2&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[2][0]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PV3)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[3][0]).toBe(
          'url_header:/data/getData.qw?pv=mean_9(PV4)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[4][0]).toBe(
          'url_header:/data/getData.qw?pv=PV5&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z'
        );
        expect(urls[5][0]).toBe(
          'url_header:/data/getData.qw?pv=PV6&from=2010-01-01T00:00:30.000Z&to=2010-01-01T00:00:30.000Z'
        );
        done();
      });
    });
  });

  describe('Build query parameters tests', () => {
    it('should return valid interval time in integer', (done) => {
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

    it('should return no interval data when interval time is less than 1 second', (done) => {
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

    it('should return filtered array when target is empty or undefined', (done) => {
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
    it('should return an empty array when no targets are set', (done) => {
      ds.query(({ targets: [] } as unknown) as DataQueryRequest<AAQuery>).subscribe((result: any) => {
        expect(result.data).toHaveLength(0);
        done();
      });
    });

    it('should return the server results when a target is set', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: 1262304000123, val: 0 },
                { millis: 1262304001456, val: 1 },
                { millis: 1262304002789, val: 2 },
              ],
            },
          ],
        })
      );

      const query = ({
        targets: [{ target: 'PV', refId: 'A' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();
        const seriesName = dataFrame.name;
        const displayName = getFieldDisplayName(dataFrame.fields[1], dataFrame);

        expect(seriesName).toBe('PV');
        expect(displayName).toBe('PV');
        expect(valArray).toHaveLength(3);
        expect(timesArray).toHaveLength(3);
        expect(valArray[0]).toBe(0);
        expect(timesArray[0]).toBe(1262304000123);
        done();
      });
    });

    it('should return the server results once for each url', (done) => {
      let counter = 0;
      datasourceRequestMock.mockImplementation((request) => {
        counter += 1;
        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [{ millis: 1262304000123, val: counter }],
            },
          ],
        });
      });

      const query = ({
        targets: [
          { target: 'PV1', refId: 'A' },
          { target: 'PV1', refId: 'B' },
          { target: 'PV1', refId: 'C', operator: 'max' },
          { target: 'PV2', refId: 'D' },
        ],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(4);
        const dataFrame1: MutableDataFrame = result.data[0];
        const dataFrame2: MutableDataFrame = result.data[1];
        const dataFrame3: MutableDataFrame = result.data[2];
        const dataFrame4: MutableDataFrame = result.data[3];
        const valArray1 = dataFrame1.fields[1].values.toArray();
        const valArray2 = dataFrame2.fields[1].values.toArray();
        const valArray3 = dataFrame3.fields[1].values.toArray();
        const valArray4 = dataFrame4.fields[1].values.toArray();

        expect(min([valArray1[0], valArray2[0], valArray3[0], valArray4[0]])).toBe(1);
        expect(max([valArray1[0], valArray2[0], valArray3[0], valArray4[0]])).toBe(3);
        done();
      });
    });

    it('should return array data when waveform is true', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'header:PV1', PREC: '0', waveform: true },
              data: [
                { millis: 1262304000123, val: [1, 2, 3] },
                { millis: 1262304001456, val: [4, 5, 6] },
                { millis: 1262304002789, val: [7, 8, 9, 10] },
              ],
            },
          ],
        })
      );

      const query = ({
        targets: [
          {
            target: 'header:PV1',
            refId: 'A',
          },
        ],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        expect(dataFrame.fields).toHaveLength(5);

        const seriesName = dataFrame.name;
        expect(seriesName).toBe('header:PV1');

        const timesArray = dataFrame.fields[0].values.toArray();
        expect(timesArray[0]).toBe(1262304000123);

        const name1 = getFieldDisplayName(dataFrame.fields[1], dataFrame);
        const name2 = getFieldDisplayName(dataFrame.fields[2], dataFrame);
        const name3 = getFieldDisplayName(dataFrame.fields[3], dataFrame);
        expect(name1).toBe('header:PV1[0]');
        expect(name2).toBe('header:PV1[1]');
        expect(name3).toBe('header:PV1[2]');

        const valArray1 = dataFrame.fields[1].values.toArray();
        const valArray2 = dataFrame.fields[2].values.toArray();
        const valArray3 = dataFrame.fields[3].values.toArray();
        const valArray4 = dataFrame.fields[4].values.toArray();
        expect(valArray1).toHaveLength(4);
        expect(valArray2).toHaveLength(4);
        expect(valArray3).toHaveLength(4);
        expect(valArray4).toHaveLength(4);

        expect(valArray1[0]).toBe(1);
        expect(valArray1[1]).toBe(4);
        expect(valArray1[2]).toBe(7);
        expect(valArray1[3]).toBe(7);

        expect(valArray2[0]).toBe(2);
        expect(valArray2[1]).toBe(5);
        expect(valArray2[2]).toBe(8);
        expect(valArray2[3]).toBe(8);

        expect(valArray4[0]).toBe(undefined);
        expect(valArray4[1]).toBe(undefined);
        expect(valArray4[2]).toBe(10);
        expect(valArray4[3]).toBe(10);
        done();
      });
    });

    it('should return the server results with alias', (done) => {
      datasourceRequestMock.mockImplementation((request) => {
        const pv = request.url.slice(31, 34);
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

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(4);
        const dataFrameArray: MutableDataFrame[] = result.data;
        const seriesName1 = dataFrameArray[0].name;
        const seriesName2 = dataFrameArray[1].name;
        const seriesName3 = dataFrameArray[2].name;
        const seriesName4 = dataFrameArray[3].name;
        const pv1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
        const pv2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
        const pv3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);
        const pv4 = getFieldDisplayName(dataFrameArray[3].fields[1], dataFrameArray[3]);

        expect(seriesName1).toBe('PV1');
        expect(seriesName2).toBe('PV2');
        expect(seriesName3).toBe('PV3');
        expect(seriesName4).toBe('PV4');
        expect(pv1).toBe('alias');
        expect(pv2).toBe('PV2');
        expect(pv3).toBe('PV3');
        expect(pv4).toBe('PV4');
        done();
      });
    });

    it('should return the server results with alias pattern', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
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

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const seriesName = dataFrame.name;
        const alias = getFieldDisplayName(dataFrame.fields[1], dataFrame);
        expect(seriesName).toBe('header:PV1');
        expect(alias).toBe('PV1:header');
        done();
      });
    });

    it('should return the server results with alias pattern for array data', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'header:PV1', PREC: '0', waveform: true },
              data: [{ millis: 1262304000123, val: [1, 2, 3] }],
            },
          ],
        })
      );

      const query = ({
        targets: [
          {
            target: 'header:PV1',
            refId: 'A',
            alias: '$1',
            aliasPattern: 'header:PV1(.*)',
          },
        ],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const seriesName = dataFrame.name;
        const alias1 = getFieldDisplayName(dataFrame.fields[1], dataFrame);
        const alias2 = getFieldDisplayName(dataFrame.fields[2], dataFrame);
        const alias3 = getFieldDisplayName(dataFrame.fields[3], dataFrame);
        expect(seriesName).toBe('header:PV1');
        expect(alias1).toBe('[0]');
        expect(alias2).toBe('[1]');
        expect(alias3).toBe('[2]');
        done();
      });
    });

    it('should return extrapolation data when operator is set as raw', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: 1262304000123, val: 0 },
                { millis: 1262304001456, val: 1 },
                { millis: 1262304002789, val: 2 },
              ],
            },
          ],
        })
      );

      const query = ({
        targets: [{ target: 'PV', refId: 'A', operator: 'raw' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
        maxDataPoints: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toHaveLength(4);
        expect(timesArray).toHaveLength(4);
        expect(valArray[3]).toBe(2);
        expect(timesArray[3]).toBe(1262390400000);
        done();
      });
    });

    it('should return extrapolation data when time range is narrow', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: 1262304000123, val: 0 },
                { millis: 1262304001456, val: 1 },
                { millis: 1262304002789, val: 2 },
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

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toHaveLength(4);
        expect(timesArray).toHaveLength(4);
        expect(valArray[3]).toBe(2);
        expect(timesArray[3]).toBe(1262304030000);
        done();
      });
    });

    it('should return normal data when the stream is enabled but rangeRaw is absent', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: 1262304000123, val: 0 },
                { millis: 1262304001456, val: 1 },
                { millis: 1262304002789, val: 2 },
              ],
            },
          ],
        })
      );

      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
        intervalMs: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        expect(result.state).toEqual(expect.not.objectContaining({ state: LoadingState.Streaming }));
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toHaveLength(4);
        expect(timesArray).toHaveLength(4);
        expect(valArray[3]).toBe(2);
        expect(timesArray[3]).toBe(1262304030000);
        done();
      });
    });

    it('should return normal data when the stream is enabled but rangeRaw is not now', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: 1262304000123, val: 0 },
                { millis: 1262304001456, val: 1 },
                { millis: 1262304002789, val: 2 },
              ],
            },
          ],
        })
      );

      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '' }],
        range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
        maxDataPoints: 1000,
        intervalMs: 1000,
        rangeRaw: { to: 'now-5m' },
      } as unknown) as DataQueryRequest<AAQuery>;

      ds.query(query).subscribe((result: any) => {
        expect(result.data).toHaveLength(1);
        expect(result.state).toEqual(expect.not.objectContaining({ state: LoadingState.Streaming }));
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toHaveLength(4);
        expect(timesArray).toHaveLength(4);
        expect(valArray[3]).toBe(2);
        expect(timesArray[3]).toBe(1262304030000);
        done();
      });
    });

    it('should return stream data when the stream is enabled', (done) => {
      datasourceRequestMock.mockImplementation((request) => {
        const from_str = unescape(split(request.url, /from=(.*Z)&to/)[1]);
        const to_str = unescape(split(request.url, /to=(.*Z)/)[1]);

        const from_ms = new Date(from_str).getTime();
        const to_ms = new Date(to_str).getTime();

        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: from_ms + 2001, val: 0 },
                { millis: Math.floor((from_ms + 2000 + to_ms) / 2), val: 1 },
                { millis: to_ms, val: 2 },
              ],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '', strmCap: '9' }],
        range: { from: new Date(now - 1000 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 1000,
        intervalMs: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(3), toArray());

      d.subscribe((results: any[]) => {
        expect(results).toHaveLength(3);
        const result = results[2];
        expect(result.data).toHaveLength(1);
        expect(result.state).toEqual(LoadingState.Streaming);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toEqual([0, 1, 2, 0, 1, 2, 0, 1, 2]);
        expect(timesArray).toHaveLength(9);

        const diff = timesArray[8] - timesArray[5];
        expect(diff).toBeGreaterThanOrEqual(1000);
        expect(diff).toBeLessThan(2000);

        done();
      });
    });

    it('should return stream data with strmInt while without strmCap', (done) => {
      jest.setTimeout(10000);
      datasourceRequestMock.mockImplementation((request) => {
        const from_str = unescape(split(request.url, /from=(.*Z)&to/)[1]);
        const to_str = unescape(split(request.url, /to=(.*Z)/)[1]);

        const from_ms = new Date(from_str).getTime();
        const to_ms = new Date(to_str).getTime();

        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: from_ms - 1, val: 0 },
                { millis: Math.floor((from_ms + to_ms) / 2), val: 1 },
                { millis: to_ms, val: 2 },
              ],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '3000' }],
        range: { from: new Date(now - 1000 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 1000,
        intervalMs: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(3), toArray());

      d.subscribe((results: any[]) => {
        expect(results).toHaveLength(3);
        const result = results[2];
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toEqual([2, 1, 2]);
        expect(timesArray).toHaveLength(3);

        const diff = timesArray[2] - timesArray[0];
        expect(diff).toBeGreaterThanOrEqual(2000);
        expect(diff).toBeLessThan(4000);

        done();
      });
    });

    it('should return stream data with unit strmInt while without strmCap', (done) => {
      jest.setTimeout(10000);
      datasourceRequestMock.mockImplementation((request) => {
        const from_str = unescape(split(request.url, /from=(.*Z)&to/)[1]);
        const to_str = unescape(split(request.url, /to=(.*Z)/)[1]);

        const from_ms = new Date(from_str).getTime();
        const to_ms = new Date(to_str).getTime();

        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: from_ms - 1, val: 0 },
                { millis: Math.floor((from_ms + to_ms) / 2), val: 1 },
                { millis: to_ms, val: 2 },
              ],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '3s' }],
        range: { from: new Date(now - 1000 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 1000,
        intervalMs: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(3), toArray());

      d.subscribe((results: any[]) => {
        expect(results).toHaveLength(3);
        const result = results[2];
        expect(result.data).toHaveLength(1);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toEqual([2, 1, 2]);
        expect(timesArray).toHaveLength(3);

        const diff = timesArray[2] - timesArray[0];
        expect(diff).toBeGreaterThanOrEqual(2000);
        expect(diff).toBeLessThan(4000);

        done();
      });
    });

    it('should ignore out of range data on stream', (done) => {
      datasourceRequestMock.mockImplementation((request) => {
        const from_str = unescape(split(request.url, /from=(.*Z)&to/)[1]);
        const to_str = unescape(split(request.url, /to=(.*Z)/)[1]);

        const from_ms = new Date(from_str).getTime();
        const to_ms = new Date(to_str).getTime();

        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: from_ms + 2000, val: 0 },
                { millis: from_ms + 2002, val: 1 },
                { millis: to_ms, val: 2 },
                { millis: to_ms + 1, val: 3 },
              ],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '', strmCap: '12' }],
        range: { from: new Date(now - 1000 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 1000,
        intervalMs: 1000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(3), toArray());

      d.subscribe((results: any[]) => {
        expect(results).toHaveLength(3);
        const result = results[2];
        expect(result.data).toHaveLength(1);
        expect(result.state).toEqual(LoadingState.Streaming);
        const dataFrame: MutableDataFrame = result.data[0];
        const timesArray = dataFrame.fields[0].values.toArray();
        const valArray = dataFrame.fields[1].values.toArray();

        expect(valArray).toEqual([0, 1, 2, 3, 1, 2, 1, 2]);
        expect(timesArray).toHaveLength(8);

        done();
      });
    });

    it('should return raw url when strmInt is less than 1000', (done) => {
      let i = 0;
      datasourceRequestMock.mockImplementation((request) => {
        if (i === 0) {
          expect(request.url.includes('mean_10')).toBeTruthy();
        } else {
          expect(request.url.includes('mean')).toBeFalsy();
        }
        i++;
        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '999', strmCap: '12' }],
        range: { from: new Date(now - 1000 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 100,
        intervalMs: 10000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(2), toArray());

      d.subscribe((results: any[]) => {
        done();
      });
    });

    it('should return mean url when strmInt is greater than 1000', (done) => {
      let i = 0;
      datasourceRequestMock.mockImplementation((request) => {
        if (i === 0) {
          expect(request.url.includes('mean_10')).toBeTruthy();
        } else {
          expect(request.url.includes('mean_1')).toBeTruthy();
        }
        i++;
        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '1500', strmCap: '12' }],
        range: { from: new Date(now - 1000 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 100,
        intervalMs: 10000,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(2), toArray());

      d.subscribe((results: any[]) => {
        done();
      });
    });

    it('should return raw url when the range is narrow', (done) => {
      datasourceRequestMock.mockImplementation((request) => {
        expect(request.url.includes('mean')).toBeFalsy();
        return Promise.resolve({
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [],
            },
          ],
        });
      });

      const now = Date.now();
      const query = ({
        targets: [{ target: 'PV', refId: 'A', stream: true, strmInt: '1500', strmCap: '12' }],
        range: { from: new Date(now - 100 * 1000), to: new Date(now) },
        rangeRaw: { to: 'now' },
        maxDataPoints: 1000,
        intervalMs: 100,
      } as unknown) as DataQueryRequest<AAQuery>;

      const d = ds.query(query).pipe(take(2), toArray());

      d.subscribe((results: any[]) => {
        done();
      });
    });
  });

  describe('PV name find query tests', () => {
    it('should return the pv name results when a target is null', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery(null, 100).then((result: any) => {
        expect(result).toHaveLength(0);
        done();
      });
    });

    it('should return the pv name results when a target is undefined', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery(undefined, 100).then((result: any) => {
        expect(result).toHaveLength(0);
        done();
      });
    });

    it('should return the pv name results when a target is empty', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
        Promise.resolve({
          data: ['metric_0', 'metric_1', 'metric_2'],
        })
      );

      ds.pvNamesFindQuery('', 100).then((result: any) => {
        expect(result).toHaveLength(0);
        done();
      });
    });

    it('should return the pv name results when a target is set', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
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
    it('should return the pv name results for metricFindQuery', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
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

    it('should return the pv name results for metricFindQuery with regex OR', (done) => {
      datasourceRequestMock.mockImplementation((request) =>
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

    it('should return the pv name results for metricFindQuery with limit parameter', (done) => {
      datasourceRequestMock.mockImplementation((request) => {
        const params = new URLSearchParams(request.url.split('?')[1]);
        const limit = parseInt(String(params.get('limit')), 10);
        const pvname = params.get('regex');
        const data = [`${pvname}${limit}`];

        return Promise.resolve({
          _request: request,
          data,
        });
      });

      ds.metricFindQuery('PV?limit=5').then((result: any) => {
        expect(result).toHaveLength(1);
        expect(result[0].text).toBe('PV5');
        done();
      });
    });

    it('should return the pv name results for metricFindQuery with invalid limit parameter', (done) => {
      datasourceRequestMock.mockImplementation((request) => {
        const params = new URLSearchParams(request.url.split('?')[1]);
        const limit = parseInt(String(params.get('limit')), 10);
        const pvname = params.get('regex');
        const data = [`${pvname}${limit}`];

        return Promise.resolve({
          _request: request,
          data,
        });
      });

      ds.metricFindQuery('PV?limit=a').then((result: any) => {
        expect(result).toHaveLength(1);
        expect(result[0].text).toBe('PV100');
        done();
      });
    });
  });
});
