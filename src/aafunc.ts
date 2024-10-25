import _ from 'lodash';
import { FuncDef, FunctionDescriptor } from './types';
import { MutableDataFrame } from '@grafana/data';
import { arrayFunctions, seriesFunctions } from './dataProcessor';

const funcIndex: { [key: string]: FuncDef } = {};
const categories: { [key: string]: FuncDef[] } = {
  Transform: [],
  'Array to Scalar': [],
  'Filter Series': [],
  Sort: [],
  Options: [],
};

function addFuncDef(newFuncDef: FuncDef) {
  const funcDef = { ...newFuncDef };

  funcDef.params = funcDef.params || [];
  funcDef.defaultParams = funcDef.defaultParams || [];

  if (funcDef.category) {
    categories[funcDef.category].push(funcDef);
  }
  funcIndex[funcDef.name] = funcDef;
  funcIndex[funcDef.shortName || funcDef.name] = funcDef;
}

function pickFuncDefsFromCategories(functionDefs: FunctionDescriptor[], requireCatecories: string[]) {
  const requiredCategoryFuncNames = _.reduce(
    requireCatecories,
    (funcNames: string[], category: string) => _.concat(funcNames, _.map(categories[category], 'name')),
    []
  );

  const pickedFuncDefs = _.filter(functionDefs, (func) => _.includes(requiredCategoryFuncNames, func.def.name));

  return pickedFuncDefs;
}

function bindFunction(
  metricFunctions: { [key: string]: (...args: any[]) => MutableDataFrame[] },
  funcDef: FunctionDescriptor
) {
  const func = metricFunctions[funcDef.def.name];

  if (!func) {
    throw new Error(`Method not found ${funcDef.def.name}`);
  }

  // Bind function arguments
  let bindedFunc = func;
  let param;
  for (let i = 0; i < funcDef.params.length; i += 1) {
    param = funcDef.params[i];

    // Convert numeric params
    if (funcDef.def.params[i].type === 'int' || funcDef.def.params[i].type === 'float') {
      param = Number(param);
    }
    bindedFunc = _.partial(bindedFunc, param);
  }
  return bindedFunc;
}

// Transform

addFuncDef({
  name: 'scale',
  category: 'Transform',
  params: [{ name: 'factor', type: 'float' }],
  defaultParams: ['100'],
});

addFuncDef({
  name: 'offset',
  category: 'Transform',
  params: [{ name: 'delta', type: 'float' }],
  defaultParams: ['100'],
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

addFuncDef({
  name: 'movingAverage',
  category: 'Transform',
  params: [{ name: 'windowSize', type: 'int' }],
  defaultParams: ['10'],
});

// Array to Scalar

addFuncDef({
  name: 'toScalarByAvg',
  category: 'Array to Scalar',
  params: [],
  defaultParams: [],
});

addFuncDef({
  name: 'toScalarByMax',
  category: 'Array to Scalar',
  params: [],
  defaultParams: [],
});

addFuncDef({
  name: 'toScalarByMin',
  category: 'Array to Scalar',
  params: [],
  defaultParams: [],
});

addFuncDef({
  name: 'toScalarBySum',
  category: 'Array to Scalar',
  params: [],
  defaultParams: [],
});

addFuncDef({
  name: 'toScalarByMed',
  category: 'Array to Scalar',
  params: [],
  defaultParams: [],
});

addFuncDef({
  name: 'toScalarByStd',
  category: 'Array to Scalar',
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
  defaultParams: ['5', 'avg'],
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
  defaultParams: ['5', 'avg'],
});

addFuncDef({
  name: 'exclude',
  category: 'Filter Series',
  params: [{ name: 'pattern', type: 'string' }],
  defaultParams: [''],
});

// Sort
addFuncDef({
  name: 'sortByAvg',
  category: 'Sort',
  params: [{ name: 'order', type: 'string', options: ['desc', 'asc'] }],
  defaultParams: ['desc'],
});

addFuncDef({
  name: 'sortByMax',
  category: 'Sort',
  params: [{ name: 'order', type: 'string', options: ['desc', 'asc'] }],
  defaultParams: ['desc'],
});

addFuncDef({
  name: 'sortByMin',
  category: 'Sort',
  params: [{ name: 'order', type: 'string', options: ['desc', 'asc'] }],
  defaultParams: ['desc'],
});

addFuncDef({
  name: 'sortBySum',
  category: 'Sort',
  params: [{ name: 'order', type: 'string', options: ['desc', 'asc'] }],
  defaultParams: ['desc'],
});

addFuncDef({
  name: 'sortByAbsMax',
  category: 'Sort',
  params: [{ name: 'order', type: 'string', options: ['desc', 'asc'] }],
  defaultParams: ['desc'],
});

addFuncDef({
  name: 'sortByAbsMin',
  category: 'Sort',
  params: [{ name: 'order', type: 'string', options: ['desc', 'asc'] }],
  defaultParams: ['desc'],
});

// Options

addFuncDef({
  name: 'maxNumPVs',
  category: 'Options',
  params: [{ name: 'number', type: 'int' }],
  defaultParams: ['100'],
});

addFuncDef({
  name: 'binInterval',
  category: 'Options',
  params: [{ name: 'interval', type: 'int' }],
  defaultParams: ['900'],
});

addFuncDef({
  name: 'disableAutoRaw',
  category: 'Options',
  params: [{ name: 'boolean', type: 'string', options: ['true', 'false'] }],
  defaultParams: ['true'],
});

addFuncDef({
  name: 'disableExtrapol',
  category: 'Options',
  params: [{ name: 'boolean', type: 'string', options: ['true', 'false'] }],
  defaultParams: ['true'],
});

addFuncDef({
  name: 'arrayFormat',
  category: 'Options',
  params: [{ name: 'format', type: 'string', options: ['timeseries', 'index', 'dt-space'] }],
  defaultParams: ['timeseries'],
});

addFuncDef({
  name: 'ignoreEmptyErr',
  category: 'Options',
  params: [{ name: 'boolean', type: 'string', options: ['true', 'false'] }],
  defaultParams: ['true'],
});

addFuncDef({
  name: 'liveOnly',
  category: 'Options',
  params: [{ name: 'boolean', type: 'string', options: ['true', 'false'] }],
  defaultParams: ['true'],
});

export function getFuncDef(name: string) {
  return funcIndex[name];
}

export function getCategories() {
  return categories;
}

export function createFuncDescriptor(funcDef: FuncDef, params?: string[]): FunctionDescriptor {
  if (!params) {
    // Create with default params
    const defaultParams = funcDef.defaultParams.slice(0);
    return { def: funcDef, params: defaultParams };
  }

  return { def: funcDef, params: params };
}

export function applyFunctionDefs(functionDefs: FunctionDescriptor[], dataFrames: MutableDataFrame[]) {
  const applyFuncDefs = pickFuncDefsFromCategories(functionDefs, ['Transform', 'Filter Series', 'Sort']);

  const promises = _.reduce(
    applyFuncDefs,
    (prevPromise, func) =>
      prevPromise.then((res) => {
        const bindedFunc = bindFunction(seriesFunctions, func);

        return Promise.resolve(bindedFunc(res));
      }),
    Promise.resolve(dataFrames)
  );

  return promises;
}

export function getToScalarFuncs(functionDefs: FunctionDescriptor[]): any[] {
  const appliedOptionFuncs = pickFuncDefsFromCategories(functionDefs, ['Array to Scalar']);

  const funcs = _.map(appliedOptionFuncs, (func) => {
    return arrayFunctions[func.def.name];
  });

  return funcs;
}

export function getOptions(functionDefs: FunctionDescriptor[]) {
  const appliedOptionFuncs = pickFuncDefsFromCategories(functionDefs, ['Options']);

  const options = _.reduce(
    appliedOptionFuncs,
    (optionMap: { [key: string]: string }, func) => {
      [optionMap[func.def.name]] = func.params;
      return optionMap;
    },
    {}
  );

  return options;
}
