import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './DataSource';
import { ConfigEditor, QueryEditor } from './components';
import { AAQuery, AADataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, AAQuery, AADataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
