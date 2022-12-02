import React, { PureComponent } from 'react';
import { DataSourceHttpSettings, InlineSwitch, InlineField } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { AADataSourceOptions } from '../types';

const LABEL_WIDTH = 26;

export type Props = DataSourcePluginOptionsEditorProps<AADataSourceOptions>;

export class ConfigEditor extends PureComponent<Props> {
  constructor(props: Props) {
    super(props);
  }

  onUseBEChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useBackend: !options.jsonData.useBackend,
    };
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options, onOptionsChange } = this.props;

    return (
      <>
        <DataSourceHttpSettings
          defaultUrl="http://localhost:17668/retrieval"
          dataSourceConfig={options}
          onChange={onOptionsChange}
        />
        <h3 className="page-heading">Misc</h3>
        <div className="gf-form-group">
          <div className="gf-form-inline">
            <InlineField
              label="Use Backend"
              labelWidth={LABEL_WIDTH}
              tooltip="Checking this option will enable the data retrieval with backend. The archived data is retrieved and processed on Grafana server, then the data is sent to Grafana client."
            >
              <InlineSwitch
                value={options.jsonData.useBackend ?? false}
                onChange={this.onUseBEChange}
              />
            </InlineField>
          </div>
        </div>
      </>
    );
  }
}
