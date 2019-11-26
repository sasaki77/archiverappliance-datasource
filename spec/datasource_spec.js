import {Datasource} from "../module";
import Q from "q";
import * as aafunc from '../aafunc';

describe('ArchiverapplianceDatasource', function() {
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

    it('should return an empty array when no targets are set', function(done) {
        ctx.ds.query({targets: []}).then( function(result) {
            expect(result.data).to.have.length(0);
            done();
        });
    });

    it('should return an valid url', function(done) {
        const target = {
            target: "PV1",
            interval: "9",
            from: new Date("2010-01-01T00:00:00.000Z"),
            to: new Date("2010-01-01T00:00:30.000Z")
        };

        ctx.ds.buildUrl(target).then( url => {
            expect(url).to.equal("url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
            done();
        });
    });

    it('should return an Error when invalid data processing is required', function(done) {
        const target = {
            target: "PV1",
            operator: "invalid",
            interval: "9",
            from: new Date("2010-01-01T00:00:00.000Z"),
            to: new Date("2010-01-01T00:00:30.000Z")
        };

        ctx.ds.buildUrl(target).then( url => {
        }).catch( error => {
          done();
        });
    });

    it('should return an data processing url', function(done) {
        const targets = [
                {target: "PV1", operator: "mean", interval: "9",
                    from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
                {target: "PV2", operator: "raw", interval: "9",
                    from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
                {target: "PV3", operator: "", interval: "9",
                    from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
                {target: "PV4", interval: "9",
                    from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
            ];

        const urls = targets.map( target => {
            return ctx.ds.buildUrl(target);
        });

        ctx.$q.all(urls).then( urls => {
            expect(urls).to.have.length(4);
            expect(urls[0]).to.equal("url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
            expect(urls[1]).to.equal("url_header:/data/getData.json?pv=PV2&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
            expect(urls[2]).to.equal("url_header:/data/getData.json?pv=mean_9(PV3)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
            expect(urls[3]).to.equal("url_header:/data/getData.json?pv=mean_9(PV4)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
            done();
        });
    });

    it('should return valid interval time in integer', function(done) {
        ctx.templateSrv.replace = function(data) {
          return data;
        };

        const options = {
            targets: [ {target: "PV1", refId: "A"} ],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T01:00:01.000Z") },
            maxDataPoints: 1800
        };

        const query = ctx.ds.buildQueryParameters(options);

        expect(query.targets).to.have.length(1);
        expect(query.targets[0].interval).to.equal("2");
        done();
    });

    it('should return no interval data when interval time is less than 1 second', function(done) {
        ctx.templateSrv.replace = function(data) {
          return data;
        };

        const options = {
            targets: [ {target: "PV1", refId: "A"} ],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z") },
            maxDataPoints: 1000
        };

        const query = ctx.ds.buildQueryParameters(options);

        expect(query.targets).to.have.length(1);
        expect(query.targets[0].interval).to.equal("");
        done();
    });

    it('should return filtered array when target is empty or undefined', function(done) {
        ctx.templateSrv.replace = function(data) {
          return data;
        }

        const options = {
            targets: [
                {target: "PV",      refId: "A"},
                {target: "",        refId: "B"},
                {target: undefined, refId: "C"}
            ],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z") },
            maxDataPoints: 1000
        };

        const query = ctx.ds.buildQueryParameters(options);

        expect(query.targets).to.have.length(1);
        done();
    });

    it('should return the server results when a target is set', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            return ctx.$q.when({
                _request: request,
                data: [
                    {
                        "meta": { "name": "PV" , "PREC": "0"},
                        "data": [
                            { "secs": 1262304000, "val": 0, "nanos": 123000000, "severity":0, "status":0 },
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
            targets: [ {target: "PV", refId: "A"} ],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z")},
            maxDataPoints: 1000
        };

        ctx.ds.query(query).then( result => {
            expect(result.data).to.have.length(1);
            const series = result.data[0];
            expect(series.target).to.equal("PV");
            expect(series.datapoints).to.have.length(3);
            expect(series.datapoints[0][0]).to.equal(0);
            expect(series.datapoints[0][1]).to.equal(1262304000123);
            done();
        });
    });

    it('should return the server results with alias', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            const pv = request.url.slice(33, 36);
            return ctx.$q.when({
                _request: request,
                data: [ { "meta": { "name": pv , "PREC": "0"}, "data": [] } ]
            });
        };

        ctx.templateSrv.replace = function(data) {
          return data;
        }

        let query = {
            targets: [
                {target: "PV1", refId: "A", alias: "alias"},
                {target: "PV2", refId: "B", alias: ""},
                {target: "PV3", refId: "C", alias: undefined},
                {target: "PV4", refId: "D"}
            ],
            range: { from: new Date("2010-01-01T00:00:00.000Z"), to: new Date("2010-01-01T00:00:30.000Z") },
            maxDataPoints: 1000
        };

        ctx.ds.query(query).then( result => {
            expect(result.data).to.have.length(4);
            expect(result.data[0].target).to.equal("alias");
            expect(result.data[1].target).to.equal("PV2");
            expect(result.data[2].target).to.equal("PV3");
            expect(result.data[3].target).to.equal("PV4");
            done();
        });
    });

    it ('should return the pv name results when a target is null', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            return ctx.$q.when({
                _request: request,
                data: [
                    "metric_0",
                    "metric_1",
                    "metric_2",
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
            return data;
        }

        ctx.ds.PVNamesFindQuery(null).then( result => {
            expect(result).to.have.length(0);
            done();
        });
    });

    it ('should return the pv name results when a target is undefined', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            return ctx.$q.when({
                _request: request,
                data: [
                    "metric_0",
                    "metric_1",
                    "metric_2",
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
            return data;
        }

        ctx.ds.PVNamesFindQuery(undefined).then( result => {
            expect(result).to.have.length(0);
            done();
        });
    });

    it ('should return the pv name results when a target is empty', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            return ctx.$q.when({
                _request: request,
                data: [
                    "metric_0",
                    "metric_1",
                    "metric_2",
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
            return data;
        }

        ctx.ds.PVNamesFindQuery("").then( result => {
            expect(result).to.have.length(0);
            done();
        });
    });

    it ('should return the pv name results when a target is set', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            return ctx.$q.when({
                _request: request,
                data: [
                    "metric_0",
                    "metric_1",
                    "metric_2",
                ]
            });
        };

        ctx.templateSrv.replace = function(data) {
            return data;
        }

        ctx.ds.PVNamesFindQuery("metric").then( result => {
            expect(result).to.have.length(3);
            expect(result[0]).to.equal('metric_0');
            expect(result[1]).to.equal('metric_1');
            expect(result[2]).to.equal('metric_2');
            done();
        });
    });
});
