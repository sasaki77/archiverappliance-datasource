import React, { PureComponent } from 'react';
import { DataSourceHttpSettings } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { AADataSourceOptions } from '../types';

export type Props = DataSourcePluginOptionsEditorProps<AADataSourceOptions>;

export class ConfigEditor extends PureComponent<Props> {
  constructor(props: Props) {
    super(props);
  }

  render() {
    const { options, onOptionsChange } = this.props;

    return (
      <>
        <DataSourceHttpSettings
          defaultUrl="http://localhost:17668/retrieval"
          dataSourceConfig={options}
          onChange={onOptionsChange}
        />
      </>
    );
  }
}
