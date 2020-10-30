import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface AAQuery extends DataQuery {
  target: string;
  alias: string;
  operator: string;
  regex: boolean;
  aliasPattern: string;
  functions: FunctionDescriptor[];
}

export const defaultQuery: Partial<AAQuery> = {
  target: '',
  alias: '',
  operator: '',
  regex: false,
  aliasPattern: '',
  functions: [],
};

export interface TargetQuery {
  target: string;
  refId: string;
  hide: boolean | undefined;
  alias: string;
  operator: string;
  functions: FunctionDescriptor[];
  regex: boolean;
  aliasPattern: string;
  options: { [key: string]: string };
  from: Date;
  to: Date;
  interval: string;
}

export interface AADataQueryData {
  meta: { name: string; waveform: boolean; PREC: string };
  data: [{ millis: number; val: number | number[] | string | string[] }];
}

export interface AADataQueryDataScalar {
  meta: { name: string; waveform: boolean; PREC: string };
  data: [{ millis: number; val: number | string }];
}

export interface AADataQueryDataNumberArray {
  meta: { name: string; waveform: boolean; PREC: string };
  data: [{ millis: number; val: number[] }];
}

export interface AADataQueryResponse {
  data: {
    data: AADataQueryData;
  };
  status: number;
  statusText: string;
  ok: boolean;
  redirected: boolean;
  type: string;
  url: string;
}

export interface FuncDef {
  defaultParams?: any;
  shortName?: any;
  version?: string;
  category: string;
  description?: string;
  fake?: boolean;
  name: string;
  params: FuncDefParam[];
}

export interface FunctionDescriptor {
  params: string[];
  def: FuncDef;
}

export interface FuncDefParam {
  name: string;
  options?: string[];
  type: string;
}

export const operatorList: string[] = [
  'firstSample',
  'lastSample',
  'firstFill',
  'lastFill',
  'mean',
  'min',
  'max',
  'count',
  'ncount',
  'nth',
  'median',
  'std',
  'jitter',
  'ignoreflyers',
  'flyers',
  'variance',
  'popvariance',
  'kurtosis',
  'skewness',
  'raw',
  'last',
];

export function isNumberArray(response: AADataQueryData): response is AADataQueryDataNumberArray {
  if (!response.meta.waveform) {
    return false;
  }

  if (Array.isArray(response.data[0].val)) {
    return typeof response.data[0].val[0] === 'number';
  }

  return false;
}

/**
 * These are options configured for each DataSource instance
 */
export interface AADataSourceOptions extends DataSourceJsonData {
  path?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface AASecureJsonData {
  apiKey?: string;
}
