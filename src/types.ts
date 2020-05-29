import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface AAQuery extends DataQuery {
  target: string;
  alias: string;
  operator: string;
  regex: boolean;
  aliasPattern: string;
  functions: any[];
}

export const defaultQuery: Partial<AAQuery> = {
  target: '',
  alias: '',
  operator: '',
  regex: false,
  aliasPattern: '',
  functions: [],
};

export interface FuncDef {
  defaultParams?: any;
  shortName?: any;
  version?: string;
  category: string;
  description?: string;
  fake?: boolean;
  name: string;
  params: Array<{ name: string; options?: string[]; type: string }>;
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
];

export interface FunctionDescriptor {
  text: string;
  params: string[];
  def: FuncDef;
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
