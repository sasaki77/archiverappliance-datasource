import { DataSource } from './DataSource';
import { DataSourcePlugin } from '@grafana/data';
import { ConfigEditor, QueryEditor } from './components';
import { AAQuery, AADataSourceOptions } from './types';
import './aa-add-func';
import './func_editor';

export const plugin = new DataSourcePlugin<DataSource, AAQuery, AADataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
