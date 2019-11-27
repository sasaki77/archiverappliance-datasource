import {Datasource} from "../module";
import Q from "q";
import * as aafunc from '../aafunc';

describe('ArchiverapplianceFunc', function() {
    var ctx = {};

    beforeEach( function() {
        ctx.instanceSettings = {
            url: "url_header:",
            jsonData: { prefix: "prefix", noparams: 1, param_names: ["test"], enbSearch: true}
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
                data: [
                    {
                        "meta": { "name": "PV" , "PREC": "0"},
                        "data": [
                            { "secs": 1262304001, "val": 1, "nanos": 456000000, "severity":0, "status":0 },
                            { "secs": 1262304002, "val": 2, "nanos": 789000000, "severity":0, "status":0 },
                        ]
                    }
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
          return data;
        }

        const query = {
            targets: [{
                target: "PV",
                refId: "A",
                functions: [
                    aafunc.createFuncInstance(aafunc.getFuncDef("scale"), [100])
                ]
            }],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
            maxDataPoints: 1000,
        };

        ctx.ds.query(query).then( result => {
            expect(result.data).to.have.length(1);
            const series = result.data[0];
            expect(series.target).to.equal("PV");
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
                data: [
                    {
                        "meta": { "name": "PV" , "PREC": "0"},
                        "data": [
                            { "secs": 1262304001, "val": 1, "nanos": 456000000, "severity":0, "status":0 },
                            { "secs": 1262304002, "val": 2, "nanos": 789000000, "severity":0, "status":0 },
                        ]
                    }
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
          return data;
        }

        const query = {
            targets: [{
                target: "PV",
                refId: "A",
                functions: [
                    aafunc.createFuncInstance(aafunc.getFuncDef("offset"), [100])
                ]
            }],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
            maxDataPoints: 1000,
        };

        ctx.ds.query(query).then( result => {
            expect(result.data).to.have.length(1);
            const series = result.data[0];
            expect(series.target).to.equal("PV");
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
                data: [
                    {
                        "meta": { "name": "PV" , "PREC": "0"},
                        "data": [
                            { "secs": 1262304001, "val": 1, "nanos": 456000000, "severity":0, "status":0 },
                            { "secs": 1262304002, "val": 2, "nanos": 789000000, "severity":0, "status":0 },
                        ]
                    }
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
          return data;
        }

        const query = {
            targets: [{
                target: "PV",
                refId: "A",
                functions: [
                    aafunc.createFuncInstance(aafunc.getFuncDef("delta"), [])
                ]
            }],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
            maxDataPoints: 1000,
        };

        ctx.ds.query(query).then( result => {
            expect(result.data).to.have.length(1);
            const series = result.data[0];
            expect(series.target).to.equal("PV");
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
                data: [
                    {
                        "meta": { "name": "PV" , "PREC": "0"},
                        "data": [
                            { "secs": 1262304001, "val": 100, "nanos": 456000000, "severity":0, "status":0 },
                            { "secs": 1262304002, "val": 200, "nanos": 789000000, "severity":0, "status":0 },
                            { "secs": 1262304002, "val": 300, "nanos": 789000000, "severity":0, "status":0 },
                        ]
                    }
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
          return data;
        }

        const query = {
            targets: [{
                target: "PV",
                refId: "A",
                functions: [
                    aafunc.createFuncInstance(aafunc.getFuncDef("fluctuation"), [])
                ]
            }],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
            maxDataPoints: 1000,
        };

        ctx.ds.query(query).then( result => {
            expect(result.data).to.have.length(1);
            const series = result.data[0];
            expect(series.target).to.equal("PV");
            expect(series.datapoints).to.have.length(3);
            expect(series.datapoints[0][0]).to.equal(0);
            expect(series.datapoints[1][0]).to.equal(100);
            expect(series.datapoints[2][0]).to.equal(200);
            done();
        });
    });

});
