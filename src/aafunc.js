import _ from 'lodash';
import $ from 'jquery';

const funcIndex = [];
const categories = {
  Transform: [],
  'Filter Series': [],
};

function addFuncDef(newFuncDef) {
  const funcDef = { ...newFuncDef };

  funcDef.params = funcDef.params || [];
  funcDef.defaultParams = funcDef.defaultParams || [];

  if (funcDef.category) {
    categories[funcDef.category].push(funcDef);
  }
  funcIndex[funcDef.name] = funcDef;
  funcIndex[funcDef.shortName || funcDef.name] = funcDef;
}

// Transform

addFuncDef({
  name: 'scale',
  category: 'Transform',
  params: [
    { name: 'factor', type: 'float', options: [100, 0.01, 10, -1] },
  ],
  defaultParams: [100],
});

addFuncDef({
  name: 'offset',
  category: 'Transform',
  params: [
    { name: 'delta', type: 'float', options: [-100, 100] },
  ],
  defaultParams: [100],
});

addFuncDef({
  name: 'delta',
  category: 'Transform',
  params: [],
  defaultParams: [],
});

addFuncDef({
  name: 'fluctuation',
  category: 'Transform',
  params: [],
  defaultParams: [],
});

// Filter Series

addFuncDef({
  name: 'top',
  category: 'Filter Series',
  params: [
    { name: 'number', type: 'int' },
    {
      name: 'value',
      type: 'string',
      options: ['avg', 'min', 'max', 'absoluteMin', 'absoluteMax', 'sum'],
    },
  ],
  defaultParams: [5, 'avg'],
});

addFuncDef({
  name: 'bottom',
  category: 'Filter Series',
  params: [
    { name: 'number', type: 'int' },
    {
      name: 'value',
      type: 'string',
      options: ['avg', 'min', 'max', 'absoluteMin', 'absoluteMax', 'sum'],
    },
  ],
  defaultParams: [5, 'avg'],
});

class FuncInstance {
  constructor(funcDef, params) {
    this.def = funcDef;

    if (params) {
      this.params = params;
    } else {
      // Create with default params
      this.params = [];
      this.params = funcDef.defaultParams.slice(0);
    }

    this.updateText();
  }

  bindFunction(metricFunctions) {
    const func = metricFunctions[this.def.name];

    if (!func) {
      throw new Error({ message: `Method not found ${this.def.name}` });
    }

    // Bind function arguments
    let bindedFunc = func;
    let param;
    for (let i = 0; i < this.params.length; i += 1) {
      param = this.params[i];

      // Convert numeric params
      if (
        this.def.params[i].type === 'int'
        || this.def.params[i].type === 'float'
      ) {
        param = Number(param);
      }
      bindedFunc = _.partial(bindedFunc, param);
    }
    return bindedFunc;
  }

  render(metricExp) {
    const str = `${this.def.name}(`;
    const parameters = _.map(this.params, (value, index) => {
      const paramType = this.def.params[index].type;
      if (
        paramType === 'int'
        || paramType === 'float'
        || paramType === 'value_or_series'
        || paramType === 'boolean'
      ) {
        return value;
      }

      if (paramType === 'int_or_interval' && $.isNumeric(value)) {
        return value;
      }

      return `'${value}'`;
    }, this);

    if (metricExp) {
      parameters.unshift(metricExp);
    }

    return `${str}${parameters.join(', ')})`;
  }

  _hasMultipleParamsInString(strValue, index) {
    if (strValue.indexOf(',') === -1) {
      return false;
    }

    return this.def.params[index + 1] && this.def.params[index + 1].optional;
  }

  updateParam(strValue, index) {
    // handle optional parameters
    // if string contains ',' and next param is optional, split and update both
    if (this._hasMultipleParamsInString(strValue, index)) {
      _.each(strValue.split(','), (partVal, idx) => {
        this.updateParam(partVal.trim(), idx);
      }, this);
      return;
    }

    if (strValue === '' && this.def.params[index].optional) {
      this.params.splice(index, 1);
    } else {
      this.params[index] = strValue;
    }

    this.updateText();
  }

  updateText() {
    if (this.params.length === 0) {
      this.text = `${this.def.name}()`;
      return;
    }

    const text = `${this.def.name}(${this.params.join(', ')})`;
    this.text = text;
  }
}

export function createFuncInstance(funcDef, params) {
  if (_.isString(funcDef)) {
    if (!funcIndex[funcDef]) {
      throw new Error({ message: `Method not found ${funcDef.name}` });
    }
    return new FuncInstance(funcIndex[funcDef], params);
  }

  return new FuncInstance(funcDef, params);
}

export function getFuncDef(name) {
  return funcIndex[name];
}

export function getCategories() {
  return categories;
}
