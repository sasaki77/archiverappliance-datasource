import _ from 'lodash';

let functions = {
  multiply: multiply
};

function multiply(n, data) {
  return _.map( data, d => {
    const datapoints = _.map(d.datapoints, point => {
        return [ point[0] * n, point[1] ];
    });
    return {"target": d.target, "datapoints": datapoints};
  });
}

export default {
  get aaFunctions() {
    return functions;
  }
};
