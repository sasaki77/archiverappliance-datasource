import { DataFrame, getFieldDisplayName, DataSourceInstanceSettings, DataQueryRequest } from '@grafana/data';
import { from } from 'rxjs';
import * as runtime from '@grafana/runtime';
import { DataSource } from '../DataSource';
import * as aafunc from '../aafunc';
import { AAQuery, AADataSourceOptions } from '../types';

const fetchMock = jest.fn().mockResolvedValue(createDefaultResponse());

jest.spyOn(runtime, 'getBackendSrv').mockImplementation(() => {
  return { fetch: fetchMock } as any as runtime.BackendSrv;
});

jest.spyOn(runtime, 'getTemplateSrv').mockImplementation(() => {
  return { replace: jest.fn().mockImplementation((query) => query) } as any as runtime.TemplateSrv;
});

beforeEach(() => {
  fetchMock.mockClear();
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

describe('Archiverappliance Functions', () => {
  let ds: DataSource;

  beforeEach(() => {
    const instanceSettings = {
      url: 'url_header:',
      jsonData: {
        useBackend: false,
      },
    } as unknown as DataSourceInstanceSettings<AADataSourceOptions>;
    ds = new DataSource(instanceSettings);
  });

  it('should apply the target functions to the query results', (done) => {
    fetchMock.mockImplementation((request) =>
      from([
        {
          data: [
            {
              meta: { name: 'PV', PREC: '0' },
              data: [
                { millis: 1262304001456, val: 1 },
                { millis: 1262304002789, val: 2 },
              ],
            },
          ],
        },
      ])
    );

    const query = {
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('scale'), ['100'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown as DataQueryRequest<AAQuery>;

    ds.query(query).subscribe((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: DataFrame = result.data[0];
      const pvname = getFieldDisplayName(dataFrame.fields[1], dataFrame);
      const timesArray = dataFrame.fields[0].values;
      const valArray = dataFrame.fields[1].values;

      expect(pvname).toBe('PV');
      expect(timesArray).toHaveLength(2);
      expect(valArray).toHaveLength(2);
      expect(timesArray[0]).toBe(1262304001456);
      expect(timesArray[1]).toBe(1262304002789);
      expect(valArray[0]).toBe(100);
      expect(valArray[1]).toBe(200);
      done();
    });
  });

  it('should return option variables if option functions are applied', (done) => {
    const options = {
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('maxNumPVs'), ['1000'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown as DataQueryRequest<AAQuery>;

    const targets = ds.buildQueryParameters(options);

    expect(targets).toHaveLength(1);
    expect(targets[0].options.maxNumPVs).toBe('1000');
    done();
  });

  it('should return 1 second interval when interval time is less than 1 second and disableAutoRaw is true', (done) => {
    const options = {
      targets: [
        {
          target: 'PV1',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('disableAutoRaw'), ['true'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    } as unknown as DataQueryRequest<AAQuery>;

    const targets = ds.buildQueryParameters(options);

    expect(targets).toHaveLength(1);
    expect(targets[0].interval).toBe('1');
    done();
  });

  it('should return non extrapolation data when disableExtrapol func is set', (done) => {
    fetchMock.mockImplementation((request) =>
      from([
        {
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
        },
      ])
    );

    const query = {
      targets: [
        {
          target: 'PV',
          refId: 'A',
          operator: 'raw',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('disableExtrapol'), ['true'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown as DataQueryRequest<AAQuery>;

    ds.query(query).subscribe((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: DataFrame = result.data[0];
      const timesArray = dataFrame.fields[0].values;
      const valArray = dataFrame.fields[1].values;

      expect(valArray).toHaveLength(3);
      expect(timesArray).toHaveLength(3);
      expect(valArray[2]).toBe(2);
      expect(timesArray[2]).toBe(1262304002789);
      done();
    });
  });
});
