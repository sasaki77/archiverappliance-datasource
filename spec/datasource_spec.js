import {Datasource} from "../module";
import Q from "q";

describe('ArchiverapplianceDatasource', function() {
    var ctx = {};

    beforeEach(function() {
        ctx.instanceSettings = {url: "url_header:", jsonData: { prefix: "prefix", noparams: 1, param_names: ["test"], enbSearch: true}};
        ctx.$q = Q;
        ctx.backendSrv = {};
        ctx.templateSrv = {};
        ctx.ds = new Datasource(ctx.instanceSettings, ctx.$q, ctx.backendSrv, ctx.templateSrv);
    });

    it('should return an empty array when no targets are set', function(done) {
        ctx.ds.query({targets: []}).then(function(result) {
            expect(result.data).to.have.length(0);
            done();
        });
    });

    it('should return an valid url', function(done) {
        const target = {target: "PV1"}
        const options = {
            intervalMs: 10000,
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        ctx.ds.buildUrl(target, options).then(function(url) {
          expect(url).to.equal("url_header:/data/getData.json?pv=mean_9(PV1)&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
          done();
        });
    });

    it('should return an Error when invalid data processing is required', function(done) {
        const target = {target: "PV1", operator: "invalid"}
        const options = {
            intervalMs: 10000,
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        ctx.ds.buildUrl(target, options).then(url => {
        }).catch( error => {
          done();
        });
    });

    it('should return an data processing url', function(done) {
        const targets = [
                {target: "PV1", operator: "mean"},
                {target: "PV2", operator: "raw"},
                {target: "PV3", operator: ""},
                {target: "PV4", },
            ]
        const options = {
            intervalMs: 10000,
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        const urls = targets.map( target => {
            return ctx.ds.buildUrl(target, options);
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

    it('should return raw data processing url when operator is less than 1 second', function(done) {
        const target = {target: "PV1", operator: "mean"};
        const options = {
            intervalMs: 1000,
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        ctx.ds.buildUrl(target, options).then( url => {
          expect(url).to.equal("url_header:/data/getData.json?pv=PV1&from=2010-01-01T00:00:00.000Z&to=2010-01-01T00:00:30.000Z");
          done();
        });
    });

    it('should return filtered array when target is empty or undefined', function(done) {
        ctx.templateSrv.replace = function(data) {
          return data;
        }

        let options = {
            targets: [
                {target: "PV",      refId: "A"},
                {target: "",        refId: "B"},
                {target: undefined, refId: "C"}
            ],
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        let query = ctx.ds.buildQueryParameters(options);

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

        let query = {
            targets: [{target: "PV", refId: "A"}],
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        ctx.ds.query(query).then(function(result) {
            expect(result.data).to.have.length(1);
            var series = result.data[0];
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
                data: [
                    { "meta": { "name": pv , "PREC": "0"}, "data": [] },
                ]
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
            range: { "from": "2010-01-01T00:00:00.000Z", "to": "2010-01-01T00:00:30.000Z"}
        };

        ctx.ds.query(query).then(function(result) {
            expect(result.data).to.have.length(4);
            expect(result.data[0].target).to.equal("alias");
            expect(result.data[1].target).to.equal("PV2");
            expect(result.data[2].target).to.equal("PV3");
            expect(result.data[3].target).to.equal("PV4");
            done();
        });
    });

    it ('should return the metric results when a target is null', function(done) {
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

        ctx.ds.metricFindQuery({target: null}).then(function(result) {
            expect(result).to.have.length(3);
            expect(result[0].text).to.equal('metric_0');
            expect(result[0].value).to.equal('metric_0');
            expect(result[1].text).to.equal('metric_1');
            expect(result[1].value).to.equal('metric_1');
            expect(result[2].text).to.equal('metric_2');
            expect(result[2].value).to.equal('metric_2');
            done();
        });
    });

    it ('should return the metric target results when a target is set', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            var target = request.data.target;
            var result = [target + "_0", target + "_1", target + "_2"];

            return ctx.$q.when({
                _request: request,
                data: result
            });
        };

        ctx.templateSrv.replace = function(data) {
            return data;
        }

        ctx.ds.metricFindQuery('entity=search').then(function(result) {
            expect(result).to.have.length(3);
            expect(result[0].text).to.equal('search_0');
            expect(result[0].value).to.equal('search_0');
            expect(result[1].text).to.equal('search_1');
            expect(result[1].value).to.equal('search_1');
            expect(result[2].text).to.equal('search_2');
            expect(result[2].value).to.equal('search_2');
            done();
        });
    });

    it ('should return the metric results when the target is an empty string', function(done) {
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

        ctx.ds.metricFindQuery('').then(function(result) {
            expect(result).to.have.length(3);
            expect(result[0].text).to.equal('metric_0');
            expect(result[0].value).to.equal('metric_0');
            expect(result[1].text).to.equal('metric_1');
            expect(result[1].value).to.equal('metric_1');
            expect(result[2].text).to.equal('metric_2');
            expect(result[2].value).to.equal('metric_2');
            done();
        });
    });

    it ('should return the metric results when the args are an empty object', function(done) {
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

        ctx.ds.metricFindQuery().then(function(result) {
            expect(result).to.have.length(3);
            expect(result[0].text).to.equal('metric_0');
            expect(result[0].value).to.equal('metric_0');
            expect(result[1].text).to.equal('metric_1');
            expect(result[1].value).to.equal('metric_1');
            expect(result[2].text).to.equal('metric_2');
            expect(result[2].value).to.equal('metric_2');
            done();
        });
    });

    it ('should return the metric target results when the args are a string', function(done) {
        ctx.backendSrv.datasourceRequest = function(request) {
            var target = request.data.target;
            var result = [target + "_0", target + "_1", target + "_2"];

            return ctx.$q.when({
                _request: request,
                data: result
            });
        };

        ctx.templateSrv.replace = function(data) {
            return data;
        }

        ctx.ds.metricFindQuery('entity=search').then(function(result) {
            expect(result).to.have.length(3);
            expect(result[0].text).to.equal('search_0');
            expect(result[0].value).to.equal('search_0');
            expect(result[1].text).to.equal('search_1');
            expect(result[1].value).to.equal('search_1');
            expect(result[2].text).to.equal('search_2');
            expect(result[2].value).to.equal('search_2');
            done();
        });
    });

    it ('should return data as text and as value', function(done) {
        var result = ctx.ds.mapToTextValue({data: ["zero", "one", "two"]});

        expect(result).to.have.length(3);
        expect(result[0].text).to.equal('zero');
        expect(result[0].value).to.equal('zero');
        expect(result[1].text).to.equal('one');
        expect(result[1].value).to.equal('one');
        expect(result[2].text).to.equal('two');
        expect(result[2].value).to.equal('two');
        done();
    });

    it ('should return text as text and value as value', function(done) {
        var data = [
            {text: "zero", value: "value_0"},
            {text: "one", value: "value_1"},
            {text: "two", value: "value_2"},
        ];

        var result = ctx.ds.mapToTextValue({data: data});

        expect(result).to.have.length(3);
        expect(result[0].text).to.equal('zero');
        expect(result[0].value).to.equal('value_0');
        expect(result[1].text).to.equal('one');
        expect(result[1].value).to.equal('value_1');
        expect(result[2].text).to.equal('two');
        expect(result[2].value).to.equal('value_2');
        done();
    });

    it ('should return data as text and index as value', function(done) {
        var data = [
            {a: "zero", b: "value_0"},
            {a: "one", b: "value_1"},
            {a: "two", b: "value_2"},
        ];

        var result = ctx.ds.mapToTextValue({data: data});

        expect(result).to.have.length(3);
        expect(result[0].text).to.equal(data[0]);
        expect(result[0].value).to.equal(0);
        expect(result[1].text).to.equal(data[1]);
        expect(result[1].value).to.equal(1);
        expect(result[2].text).to.equal(data[2]);
        expect(result[2].value).to.equal(2);
        done();
    });
});
