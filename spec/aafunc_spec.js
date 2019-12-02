import _ from 'lodash';
import {Datasource} from '../module';
import Q from 'q';
import * as aafunc from '../aafunc';
import dataProcessor from '../dataProcessor';

describe('ArchiverapplianceFunc', function() {
  var ctx = {};

  beforeEach( function() {
    ctx.instanceSettings = {
      url: 'url_header:',
    };
    ctx.$q = Q;
    ctx.backendSrv = {};
    ctx.templateSrv = {};
    ctx.ds = new Datasource(ctx.instanceSettings, ctx.$q, ctx.backendSrv, ctx.templateSrv);
  });

  it('should return the server results with scale function', function(done) {
    ctx.backendSrv.datasourceRequest = function(request) {
      return ctx.$q.when({
        _request: request,
        data: [{
          meta: { 'name': 'PV' , 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 1, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 2, nanos: 789000000, severity:0, status:0 },
          ]
        }]
      });
    };

    ctx.templateSrv.replace = function(data) {
      return data;
    }

    const query = {
      targets: [{
        target: 'PV',
        refId: 'A',
        functions: [
            aafunc.createFuncInstance(aafunc.getFuncDef('scale'), [100])
        ]
      }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z')},
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then((result) => {
      expect(result.data).to.have.length(1);
      const series = result.data[0];
      expect(series.target).to.equal('PV');
      expect(series.datapoints).to.have.length(2);
      expect(series.datapoints[0][0]).to.equal(100);
      expect(series.datapoints[1][0]).to.equal(200);
      expect(series.datapoints[0][1]).to.equal(1262304001456);
      expect(series.datapoints[1][1]).to.equal(1262304002789);
      done();
    });
  });

  it('should return the server results with offset function', function(done) {
    ctx.backendSrv.datasourceRequest = function(request) {
      return ctx.$q.when({
        _request: request,
        data: [{
          meta: { 'name': 'PV' , 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 1, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 2, nanos: 789000000, severity:0, status:0 },
          ]
        }]
      });
    };

    ctx.templateSrv.replace = function(data) {
      return data;
    }

    const query = {
      targets: [{
        target: 'PV',
        refId: 'A',
        functions: [
            aafunc.createFuncInstance(aafunc.getFuncDef('offset'), [100])
        ]
      }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z')},
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then((result) => {
      expect(result.data).to.have.length(1);
      const series = result.data[0];
      expect(series.target).to.equal('PV');
      expect(series.datapoints).to.have.length(2);
      expect(series.datapoints[0][0]).to.equal(101);
      expect(series.datapoints[1][0]).to.equal(102);
      expect(series.datapoints[0][1]).to.equal(1262304001456);
      expect(series.datapoints[1][1]).to.equal(1262304002789);
      done();
    });
  });

  it('should return the server results with delta function', function(done) {
    ctx.backendSrv.datasourceRequest = function(request) {
      return ctx.$q.when({
        _request: request,
        data: [{
          meta: { 'name': 'PV' , 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 1, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 2, nanos: 789000000, severity:0, status:0 },
          ]
        }]
      });
    };

    ctx.templateSrv.replace = function(data) {
      return data;
    }

    const query = {
      targets: [{
        target: 'PV',
        refId: 'A',
        functions: [
          aafunc.createFuncInstance(aafunc.getFuncDef('delta'), [])
        ]
      }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z')},
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then((result) => {
      expect(result.data).to.have.length(1);
      const series = result.data[0];
      expect(series.target).to.equal('PV');
      expect(series.datapoints).to.have.length(1);
      expect(series.datapoints[0][0]).to.equal(1);
      expect(series.datapoints[0][1]).to.equal(1262304002789);
        done();
    });
  });


  it('should return the server results with fluctuation function', function(done) {
    ctx.backendSrv.datasourceRequest = function(request) {
      return ctx.$q.when({
        _request: request,
        data: [{
          meta: { 'name': 'PV' , 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 100, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 200, nanos: 789000000, severity:0, status:0 },
            { secs: 1262304002, val: 300, nanos: 789000000, severity:0, status:0 },
          ]
        }]
      });
    };

    ctx.templateSrv.replace = function(data) {
      return data;
    }

    const query = {
      targets: [{
        target: 'PV',
        refId: 'A',
        functions: [
            aafunc.createFuncInstance(aafunc.getFuncDef('fluctuation'), [])
        ]
      }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z')},
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then( result => {
      expect(result.data).to.have.length(1);
      const series = result.data[0];
      expect(series.target).to.equal('PV');
      expect(series.datapoints).to.have.length(3);
      expect(series.datapoints[0][0]).to.equal(0);
      expect(series.datapoints[1][0]).to.equal(100);
      expect(series.datapoints[2][0]).to.equal(200);
      done();
    });
  });

  it('should return the server results with top function', function(done) {
    ctx.backendSrv.datasourceRequest = function(request) {
      const pvname = unescape(_.split(request.url, /pv=(.*?)\&/)[1]);
      let data = [];
      if (pvname === 'PV1') {
        data = [{
          meta: { 'name': pvname, 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 0, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 1, nanos: 789000000, severity:0, status:0 },
            { secs: 1262304002, val: 2, nanos: 789000000, severity:0, status:0 }
          ]
        }];
      } else if (pvname === 'PV2') {
        data = [{
          meta: { 'name': pvname, 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 3, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 4, nanos: 789000000, severity:0, status:0 },
            { secs: 1262304002, val: 5, nanos: 789000000, severity:0, status:0 },
          ]
        }];
      } else {
        data = [{
          meta: { 'name': pvname, 'PREC': '0'},
          data: [
            { secs: 1262304001, val: 0, nanos: 456000000, severity:0, status:0 },
            { secs: 1262304002, val: 0, nanos: 789000000, severity:0, status:0 },
            { secs: 1262304002, val: 0, nanos: 789000000, severity:0, status:0 },
          ]
        }];
      }

      return ctx.$q.when({
        _request: request,
        data: data
      });
    };

    ctx.templateSrv.replace = function(data) {
      return data;
    }

    const query = {
      targets: [{
        target: '(PV1|PV2|PV3)',
        refId: 'A',
        functions: [
            aafunc.createFuncInstance(aafunc.getFuncDef('top'), [2, 'avg'])
        ]
      }],
      range: { from: new Date('2010-01-01T00:00:00.000Z'), to: new Date('2010-01-01T00:00:30.000Z')},
      maxDataPoints: 1000,
    };

    ctx.ds.query(query).then( result => {
      expect(result.data).to.have.length(2);
      const series1 = result.data[0];
      const series2 = result.data[1];
      expect(series1.target).to.equal('PV2');
      expect(series2.target).to.equal('PV1');
      expect(series1.datapoints).to.have.length(3);
      expect(series1.datapoints[0][0]).to.equal(3);
      expect(series1.datapoints[1][0]).to.equal(4);
      expect(series1.datapoints[2][0]).to.equal(5);
      done();
    });
  });

  it ('should return aggregated value', function() {
    const timeseriesData = [
      { target: 'min', datapoints: [[0,0], [1,0], [2,0], [3,0], [4,0]] },
      { target: 'max', datapoints: [[1,0], [1,0], [1,0], [1,0], [7,0]] },
      { target: 'avgsum', datapoints: [[2,0], [3,0], [4,0], [5,0], [6,0]] }
    ];

    const topFunction = dataProcessor.aaFunctions.top;
    const minTopData = topFunction(3, 'min', timeseriesData);
    const maxTopData = topFunction(1, 'max', timeseriesData);
    const avgTopData = topFunction(1, 'avg', timeseriesData);
    const sumTopData = topFunction(1, 'sum', timeseriesData);

    const bottomFunction = dataProcessor.aaFunctions.bottom;
    const minBottomData = bottomFunction(3, 'min', timeseriesData);
    const maxBottomData = bottomFunction(1, 'max', timeseriesData);
    const avgBottomData = bottomFunction(1, 'avg', timeseriesData);
    const sumBottomData = bottomFunction(1, 'sum', timeseriesData);

    expect(minTopData).to.have.length(3);
    expect(minTopData[0].target).to.equal('avgsum');
    expect(minTopData[1].target).to.equal('max');
    expect(minTopData[2].target).to.equal('min');

    expect(maxTopData).to.have.length(1);
    expect(maxTopData[0].target).to.equal('max');

    expect(avgTopData).to.have.length(1);
    expect(avgTopData[0].target).to.equal('avgsum');

    expect(sumTopData).to.have.length(1);
    expect(sumTopData[0].target).to.equal('avgsum');

    expect(minBottomData).to.have.length(3);
    expect(minBottomData[0].target).to.equal('min');
    expect(minBottomData[1].target).to.equal('max');
    expect(minBottomData[2].target).to.equal('avgsum');

    expect(maxBottomData).to.have.length(1);
    expect(maxBottomData[0].target).to.equal('min');

    expect(avgBottomData).to.have.length(1);
    expect(avgBottomData[0].target).to.equal('min');

    expect(sumBottomData).to.have.length(1);
    expect(sumBottomData[0].target).to.equal('min');
  });

});
