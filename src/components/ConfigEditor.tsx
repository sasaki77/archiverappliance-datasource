import React, { PureComponent } from 'react';
import { DataSourceHttpSettings } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { LegacyForms } from '@grafana/ui';
import { AADataSourceOptions } from '../types';


export type Props = DataSourcePluginOptionsEditorProps<AADataSourceOptions>;

export class ConfigEditor extends PureComponent<Props> {
  constructor(props: Props) {
    super(props);
  }

  onUseBEChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useBackend: !options.jsonData.useBackend
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
          <div className="gf-form">
            <LegacyForms.Switch
              checked={options.jsonData.useBackend ?? false}
              label="Use Backend"
              labelClass={'width-13'}
              tooltip="Enable/Disable Regex mode. You can select multiple PVs using Regular Expressoins."
              onChange={this.onUseBEChange}
            />
          </div>
        </div>
      </>
    );
  }
}
