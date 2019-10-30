import prunk from 'prunk';
import chai from 'chai';

// Mock Grafana modules that are not available outside of the core project
// Required for loading module.js
prunk.mock('./css/query-editor.css!', 'no css, dude.');
prunk.mock('app/plugins/sdk', {
    QueryCtrl: null
});

// Setup jsdom
// Required for loading angularjs
const jsdom = require("jsdom");
const { JSDOM } = jsdom;
global.document = new JSDOM('<html><head><script></script></head><body></body></html>');
global.window = global.document.parentWindow;

// Setup Chai
chai.should();
global.assert = chai.assert;
global.expect = chai.expect;
