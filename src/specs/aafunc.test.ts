import split from 'lodash/split';
import {
  MutableDataFrame,
  FieldType,
  getFieldDisplayName,
  DataSourceInstanceSettings,
  DataQueryRequest,
} from '@grafana/data';
import { DataSource } from '../DataSource';
import * as aafunc from '../aafunc';
import { seriesFunctions } from '../dataProcessor';
import { AAQuery, AADataSourceOptions } from '../types';

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
    const instanceSettings = ({
      url: 'url_header:',
    } as unknown) as DataSourceInstanceSettings<AADataSourceOptions>;
    ds = new DataSource(instanceSettings);
  });

  it('should return the server results with scale function', done => {
    datasourceRequestMock.mockImplementation(request =>
      Promise.resolve({
        data: [
          {
            meta: { name: 'PV', PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
            ],
          },
        ],
      })
    );

    const query = ({
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('scale'), ['100'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: MutableDataFrame = result.data[0];
      const pvname = getFieldDisplayName(dataFrame.fields[1], dataFrame);
      const timesArray = dataFrame.fields[0].values.toArray();
      const valArray = dataFrame.fields[1].values.toArray();

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

  it('should return the server results with offset function', done => {
    datasourceRequestMock.mockImplementation(request =>
      Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: 'PV', PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
            ],
          },
        ],
      })
    );

    const query = ({
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('offset'), ['100'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: MutableDataFrame = result.data[0];
      const pvname = getFieldDisplayName(dataFrame.fields[1], dataFrame);
      const timesArray = dataFrame.fields[0].values.toArray();
      const valArray = dataFrame.fields[1].values.toArray();

      expect(pvname).toBe('PV');
      expect(timesArray).toHaveLength(2);
      expect(valArray).toHaveLength(2);
      expect(timesArray[0]).toBe(1262304001456);
      expect(timesArray[1]).toBe(1262304002789);
      expect(valArray[0]).toBe(101);
      expect(valArray[1]).toBe(102);
      done();
    });
  });

  it('should return the server results with delta function', done => {
    datasourceRequestMock.mockImplementation(request =>
      Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: 'PV', PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
            ],
          },
        ],
      })
    );

    const query = ({
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('delta'), [])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: MutableDataFrame = result.data[0];
      const pvname = getFieldDisplayName(dataFrame.fields[1], dataFrame);
      const timesArray = dataFrame.fields[0].values.toArray();
      const valArray = dataFrame.fields[1].values.toArray();

      expect(pvname).toBe('PV');
      expect(timesArray).toHaveLength(1);
      expect(valArray).toHaveLength(1);
      expect(timesArray[0]).toBe(1262304002789);
      expect(valArray[0]).toBe(1);
      done();
    });
  });

  it('should return the server results with fluctuation function', done => {
    datasourceRequestMock.mockImplementation(request =>
      Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: 'PV', PREC: '0' },
            data: [
              { millis: 1262304001456, val: 100 },
              { millis: 1262304002789, val: 200 },
              { millis: 1262304002789, val: 300 },
            ],
          },
        ],
      })
    );

    const query = ({
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('fluctuation'), [])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: MutableDataFrame = result.data[0];
      const pvname = getFieldDisplayName(dataFrame.fields[1], dataFrame);
      const timesArray = dataFrame.fields[0].values.toArray();
      const valArray = dataFrame.fields[1].values.toArray();

      expect(pvname).toBe('PV');
      expect(timesArray).toHaveLength(3);
      expect(valArray).toHaveLength(3);
      expect(valArray[0]).toBe(0);
      expect(valArray[1]).toBe(100);
      expect(valArray[2]).toBe(200);
      done();
    });
  });

  it('should return the server results with movingAverage function', done => {
    datasourceRequestMock.mockImplementation(request =>
      Promise.resolve({
        _request: request,
        data: [
          {
            meta: { name: 'PV', PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
              { millis: 1262304002789, val: 3 },
              { millis: 1262304002790, val: 4 },
              { millis: 1262304002791, val: 5 },
              { millis: 1262304002792, val: 6 },
              { millis: 1262304002793, val: 7 },
            ],
          },
        ],
      })
    );

    const query = ({
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('movingAverage'), ['3'])],
        },
        {
          target: 'PV',
          refId: 'B',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('movingAverage'), ['8'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(2);
      const dataFrame: MutableDataFrame = result.data[0];
      const pvname = getFieldDisplayName(dataFrame.fields[1], dataFrame);
      const timesArray = dataFrame.fields[0].values.toArray();
      const valArray = dataFrame.fields[1].values.toArray();

      expect(pvname).toBe('PV');
      expect(timesArray).toHaveLength(7);
      expect(valArray).toHaveLength(7);
      expect(valArray[0]).toBe(1);
      expect(valArray[1]).toBe(1.5);
      expect(valArray[2]).toBe(2);
      expect(valArray[6]).toBe(6);

      const dataFrame2: MutableDataFrame = result.data[1];
      const valArray2 = dataFrame2.fields[1].values.toArray();
      const timesArray2 = dataFrame2.fields[0].values.toArray();
      expect(timesArray2).toHaveLength(7);
      expect(valArray2).toHaveLength(7);
      expect(valArray2[6]).toBe(7);
      done();
    });
  });

  it('should return correct scalar data with toScalar funcs', done => {
    datasourceRequestMock.mockImplementation(request =>
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
          functions: [
            aafunc.createFuncDescriptor(aafunc.getFuncDef('toScalarByAvg'), []),
            aafunc.createFuncDescriptor(aafunc.getFuncDef('toScalarByMax'), []),
            aafunc.createFuncDescriptor(aafunc.getFuncDef('toScalarByMin'), []),
            aafunc.createFuncDescriptor(aafunc.getFuncDef('toScalarBySum'), []),
            aafunc.createFuncDescriptor(aafunc.getFuncDef('toScalarByMed'), []),
            aafunc.createFuncDescriptor(aafunc.getFuncDef('toScalarByStd'), []),
          ],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(6);
      const dataFrameAvg: MutableDataFrame = result.data[0];
      const dataFrameMax: MutableDataFrame = result.data[1];
      const dataFrameMin: MutableDataFrame = result.data[2];
      const dataFrameSum: MutableDataFrame = result.data[3];
      const dataFrameMed: MutableDataFrame = result.data[4];
      const dataFrameStd: MutableDataFrame = result.data[5];

      const seriesNameAvg = dataFrameAvg.name;
      const seriesNameMax = dataFrameMax.name;
      const seriesNameMin = dataFrameMin.name;
      const seriesNameSum = dataFrameSum.name;
      const seriesNameMed = dataFrameMed.name;
      const seriesNameStd = dataFrameStd.name;
      expect(seriesNameAvg).toBe('header:PV1');
      expect(seriesNameMax).toBe('header:PV1');
      expect(seriesNameMin).toBe('header:PV1');
      expect(seriesNameSum).toBe('header:PV1');
      expect(seriesNameMed).toBe('header:PV1');
      expect(seriesNameStd).toBe('header:PV1');

      const valArrayAvg = dataFrameAvg.fields[1].values.toArray();
      const valArrayMax = dataFrameMax.fields[1].values.toArray();
      const valArrayMin = dataFrameMin.fields[1].values.toArray();
      const valArraySum = dataFrameSum.fields[1].values.toArray();
      const valArrayMed = dataFrameMed.fields[1].values.toArray();
      const valArrayStd = dataFrameStd.fields[1].values.toArray();

      expect(valArrayAvg).toHaveLength(4);
      expect(valArrayMax).toHaveLength(4);
      expect(valArrayMin).toHaveLength(4);
      expect(valArraySum).toHaveLength(4);
      expect(valArrayMed).toHaveLength(4);
      expect(valArrayStd).toHaveLength(4);
      expect(valArrayAvg[0]).toBe(2);
      expect(valArrayMax[0]).toBe(3);
      expect(valArrayMin[0]).toBe(1);
      expect(valArraySum[0]).toBe(6);
      expect(valArrayMed[0]).toBe(2);
      expect(valArrayStd[0]).toBe(1);

      const nameAvg = getFieldDisplayName(dataFrameAvg.fields[1], dataFrameAvg);
      const nameMax = getFieldDisplayName(dataFrameMax.fields[1], dataFrameMax);
      const nameMin = getFieldDisplayName(dataFrameMin.fields[1], dataFrameMin);
      const nameSum = getFieldDisplayName(dataFrameSum.fields[1], dataFrameSum);
      const nameMed = getFieldDisplayName(dataFrameMed.fields[1], dataFrameMed);
      const nameStd = getFieldDisplayName(dataFrameStd.fields[1], dataFrameStd);
      expect(nameAvg).toBe('header:PV1 (avg)');
      expect(nameMax).toBe('header:PV1 (max)');
      expect(nameMin).toBe('header:PV1 (min)');
      expect(nameSum).toBe('header:PV1 (sum)');
      expect(nameMed).toBe('header:PV1 (median)');
      expect(nameStd).toBe('header:PV1 (std)');

      const timesArrayAvg = dataFrameAvg.fields[0].values.toArray();
      const timesArrayMax = dataFrameMax.fields[0].values.toArray();
      const timesArrayMin = dataFrameMin.fields[0].values.toArray();
      const timesArraySum = dataFrameSum.fields[0].values.toArray();
      const timesArrayMed = dataFrameMed.fields[0].values.toArray();
      const timesArrayStd = dataFrameStd.fields[0].values.toArray();
      expect(timesArrayAvg[0]).toBe(1262304000123);
      expect(timesArrayMax[0]).toBe(1262304000123);
      expect(timesArrayMin[0]).toBe(1262304000123);
      expect(timesArraySum[0]).toBe(1262304000123);
      expect(timesArrayMed[0]).toBe(1262304000123);
      expect(timesArrayStd[0]).toBe(1262304000123);

      done();
    });
  });

  it('should return the server results with top function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 1 },
              { millis: 1262304002789, val: 2 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 3 },
              { millis: 1262304002789, val: 4 },
              { millis: 1262304002789, val: 5 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 0 },
              { millis: 1262304002789, val: 0 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('top'), ['2', 'avg'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(2);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const timesArray1 = dataFrameArray[0].fields[0].values.toArray();
      const valArray1 = dataFrameArray[0].fields[1].values.toArray();

      expect(pvname1).toBe('PV2');
      expect(pvname2).toBe('PV1');
      expect(timesArray1).toHaveLength(3);
      expect(valArray1).toHaveLength(3);
      expect(valArray1[0]).toBe(3);
      expect(valArray1[1]).toBe(4);
      expect(valArray1[2]).toBe(5);
      done();
    });
  });

  it('should return aggregated value', () => {
    const data = [
      {
        target: 'min',
        name: 'min',
        values: [0, 1, 2, 3, 4],
        times: [0, 0, 0, 0, 0],
        datapoints: [
          [0, 0],
          [1, 0],
          [2, 0],
          [3, 0],
          [4, 0],
        ],
      },
      {
        target: 'max',
        name: 'max',
        values: [1, 1, 1, 1, 7],
        times: [0, 0, 0, 0, 0],
        datapoints: [
          [1, 0],
          [1, 0],
          [1, 0],
          [1, 0],
          [7, 0],
        ],
      },
      {
        target: 'avgsum',
        name: 'avgsum',
        values: [2, 3, 4, 5, 6],
        times: [0, 0, 0, 0, 0],
        datapoints: [
          [2, 0],
          [3, 0],
          [4, 0],
          [5, 0],
          [6, 0],
        ],
      },
    ];

    const timeseriesData: MutableDataFrame[] = data.map(
      d =>
        new MutableDataFrame({
          name: d.name,
          fields: [
            { name: 'time', type: FieldType.time, values: d.times },
            { name: 'value', type: FieldType.number, values: d.values, config: { displayName: d.name } },
          ],
        })
    );

    const topFunction = seriesFunctions.top;
    const bottomFunction = seriesFunctions.bottom;

    const minTopData = topFunction(3, 'min', timeseriesData);
    const maxTopData = topFunction(1, 'max', timeseriesData);
    const avgTopData = topFunction(1, 'avg', timeseriesData);
    const sumTopData = topFunction(1, 'sum', timeseriesData);

    const minBottomData = bottomFunction(3, 'min', timeseriesData);
    const maxBottomData = bottomFunction(1, 'max', timeseriesData);
    const avgBottomData = bottomFunction(1, 'avg', timeseriesData);
    const sumBottomData = bottomFunction(1, 'sum', timeseriesData);

    expect(minTopData).toHaveLength(3);
    expect(minTopData[0].name).toBe('avgsum');
    expect(minTopData[1].name).toBe('max');
    expect(minTopData[2].name).toBe('min');

    expect(maxTopData).toHaveLength(1);
    expect(maxTopData[0].name).toBe('max');

    expect(avgTopData).toHaveLength(1);
    expect(avgTopData[0].name).toBe('avgsum');

    expect(sumTopData).toHaveLength(1);
    expect(sumTopData[0].name).toBe('avgsum');

    expect(minBottomData).toHaveLength(3);
    expect(minBottomData[0].name).toBe('min');
    expect(minBottomData[1].name).toBe('max');
    expect(minBottomData[2].name).toBe('avgsum');

    expect(maxBottomData).toHaveLength(1);
    expect(maxBottomData[0].name).toBe('min');

    expect(avgBottomData).toHaveLength(1);
    expect(avgBottomData[0].name).toBe('min');

    expect(sumBottomData).toHaveLength(1);
    expect(sumBottomData[0].name).toBe('min');

    const absData = [
      {
        target: 'min',
        name: 'min',
        values: [0, 3],
        times: [0, 0],
        datapoints: [
          [0, 0],
          [3, 0],
        ],
      },
      {
        target: 'max',
        name: 'max',
        values: [-6, -5, -4, -3, -2],
        times: [0, 0, 0, 0, 0],
        datapoints: [
          [-6, 0],
          [-5, 0],
          [-4, 0],
          [-3, 0],
          [-2, 0],
        ],
      },
      {
        target: 'avgsum',
        name: 'avgsum',
        values: [3, 4],
        times: [0, 0],
        datapoints: [
          [3, 0],
          [4, 0],
        ],
      },
    ];

    const timeseriesDataAbs: MutableDataFrame[] = absData.map(
      d =>
        new MutableDataFrame({
          name: d.name,
          fields: [
            { name: 'time', type: FieldType.time, values: d.times },
            { name: 'value', type: FieldType.number, values: d.values, config: { displayName: d.name } },
          ],
        })
    );

    const absMinTopData = topFunction(1, 'absoluteMin', timeseriesDataAbs);
    const absMaxTopData = topFunction(1, 'absoluteMax', timeseriesDataAbs);
    const absMinBottomData = bottomFunction(1, 'absoluteMin', timeseriesDataAbs);
    const absMaxBottomData = bottomFunction(1, 'absoluteMax', timeseriesDataAbs);

    expect(absMinTopData).toHaveLength(1);
    expect(absMinTopData[0].name).toBe('avgsum');

    expect(absMaxTopData).toHaveLength(1);
    expect(absMaxTopData[0].name).toBe('max');

    expect(absMinBottomData).toHaveLength(1);
    expect(absMinBottomData[0].name).toBe('min');

    expect(absMaxBottomData).toHaveLength(1);
    expect(absMaxBottomData[0].name).toBe('min');
  });

  it('should return the server results with exclude function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      const pvdata = [
        {
          meta: { name: pvname, PREC: '0' },
          data: [{ millis: 1262304001456, val: 0 }],
        },
      ];

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PVA|PVB)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('exclude'), ['PV[0-9]'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(2);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);

      expect(pvname1).toBe('PVA');
      expect(pvname2).toBe('PVB');
      done();
    });
  });

  it('should return the server results with sortByMax function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 1 },
              { millis: 1262304002789, val: 2 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 3 },
              { millis: 1262304002789, val: 4 },
              { millis: 1262304002789, val: 5 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 0 },
              { millis: 1262304002789, val: 0 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('sortByMax'), ['desc'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(3);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const pvname3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);

      expect(pvname1).toBe('PV2');
      expect(pvname2).toBe('PV1');
      expect(pvname3).toBe('PV3');
      done();
    });
  });

  it('should return the server results with sortByMin function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
              { millis: 1262304002789, val: 3 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 3 },
              { millis: 1262304002789, val: 4 },
              { millis: 1262304002789, val: 5 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 0 },
              { millis: 1262304002789, val: 0 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('sortByMin'), ['asc'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(3);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const pvname3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);

      expect(pvname1).toBe('PV3');
      expect(pvname2).toBe('PV1');
      expect(pvname3).toBe('PV2');
      done();
    });
  });

  it('should return the server results with sortByAvg function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
              { millis: 1262304002789, val: 3 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 3 },
              { millis: 1262304002789, val: 4 },
              { millis: 1262304002789, val: 5 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 0 },
              { millis: 1262304002789, val: 0 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('sortByMin'), ['desc'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(3);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const pvname3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);

      expect(pvname1).toBe('PV2');
      expect(pvname2).toBe('PV1');
      expect(pvname3).toBe('PV3');
      done();
    });
  });

  it('should return the server results with sortBySum function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
              { millis: 1262304002789, val: 3 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 3 },
              { millis: 1262304002789, val: 4 },
              { millis: 1262304002789, val: 5 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 0 },
              { millis: 1262304002789, val: 0 },
              { millis: 1262304002789, val: 0 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('sortByMin'), ['asc'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(3);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const pvname3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);

      expect(pvname1).toBe('PV3');
      expect(pvname2).toBe('PV1');
      expect(pvname3).toBe('PV2');
      done();
    });
  });

  it('should return the server results with sortByAbsMax function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
              { millis: 1262304002789, val: 3 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 3 },
              { millis: 1262304002789, val: 4 },
              { millis: 1262304002789, val: 5 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: -10 },
              { millis: 1262304002789, val: 0 },
              { millis: 1262304002789, val: 0 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('sortByAbsMax'), ['desc'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(3);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const pvname3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);

      expect(pvname1).toBe('PV3');
      expect(pvname2).toBe('PV2');
      expect(pvname3).toBe('PV1');
      done();
    });
  });

  it('should return the server results with sortByAbsMin function', done => {
    datasourceRequestMock.mockImplementation(request => {
      const pvname = unescape(split(request.url, /pv=mean_[0-9].*\((.*?)\)&/)[1]);
      let pvdata = [];
      if (pvname === 'PV1') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: 1 },
              { millis: 1262304002789, val: 2 },
              { millis: 1262304002789, val: 3 },
            ],
          },
        ];
      } else if (pvname === 'PV2') {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: -6 },
              { millis: 1262304002789, val: 7 },
              { millis: 1262304002789, val: 8 },
            ],
          },
        ];
      } else {
        pvdata = [
          {
            meta: { name: pvname, PREC: '0' },
            data: [
              { millis: 1262304001456, val: -5 },
              { millis: 1262304002789, val: 10 },
              { millis: 1262304002789, val: 10 },
            ],
          },
        ];
      }

      return Promise.resolve({
        _request: request,
        data: pvdata,
      });
    });

    const query = ({
      targets: [
        {
          target: '(PV1|PV2|PV3)',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('sortByAbsMax'), ['desc'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(3);
      const dataFrameArray: MutableDataFrame[] = result.data;
      const pvname1 = getFieldDisplayName(dataFrameArray[0].fields[1], dataFrameArray[0]);
      const pvname2 = getFieldDisplayName(dataFrameArray[1].fields[1], dataFrameArray[1]);
      const pvname3 = getFieldDisplayName(dataFrameArray[2].fields[1], dataFrameArray[2]);

      expect(pvname1).toBe('PV3');
      expect(pvname2).toBe('PV2');
      expect(pvname3).toBe('PV1');
      done();
    });
  });

  it('should return option variables if option functions are applied', done => {
    const options = ({
      targets: [
        {
          target: 'PV',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('maxNumPVs'), ['1000'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-02T00:00:00.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    const targets = ds.buildQueryParameters(options);

    expect(targets).toHaveLength(1);
    expect(targets[0].options.maxNumPVs).toBe('1000');
    done();
  });

  it('should return 1 second interval when interval time is less than 1 second and disableAutoRaw is true', done => {
    const options = ({
      targets: [
        {
          target: 'PV1',
          refId: 'A',
          functions: [aafunc.createFuncDescriptor(aafunc.getFuncDef('disableAutoRaw'), ['true'])],
        },
      ],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z') },
      maxDataPoints: 1000,
    } as unknown) as DataQueryRequest<AAQuery>;

    const targets = ds.buildQueryParameters(options);

    expect(targets).toHaveLength(1);
    expect(targets[0].interval).toBe('1');
    done();
  });

  it('should return non extrapolation data when disableExtrapol func is set', done => {
    datasourceRequestMock.mockImplementation(request =>
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
    } as unknown) as DataQueryRequest<AAQuery>;

    ds.query(query).then((result: any) => {
      expect(result.data).toHaveLength(1);
      const dataFrame: MutableDataFrame = result.data[0];
      const timesArray = dataFrame.fields[0].values.toArray();
      const valArray = dataFrame.fields[1].values.toArray();

      expect(valArray).toHaveLength(3);
      expect(timesArray).toHaveLength(3);
      expect(valArray[2]).toBe(2);
      expect(timesArray[2]).toBe(1262304002789);
      done();
    });
  });
});
