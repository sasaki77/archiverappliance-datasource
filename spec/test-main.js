import prunk from 'prunk';
import chai from 'chai';

// Mock Grafana modules that are not available outside of the core project
// Required for loading module.js
prunk.mock('./css/query-editor.css!', 'no css, dude.');
prunk.mock('app/plugins/sdk', {
  QueryCtrl: null,
});

prunk.mock('app/core/core_module', {
  directive: () => {},
});

// Setup jsdom
// Required for loading angularjs
const jsdom = require('jsdom');

const { JSDOM } = jsdom;

global.dom = new JSDOM('<!doctype html><html><body></body></html>');
global.window = dom.window;
global.document = dom.window.document;
global.navigator = global.window.navigator;
global.Element = () => {};

// Setup Chai
chai.should();
global.assert = chai.assert;
global.expect = chai.expect;
