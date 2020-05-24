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
