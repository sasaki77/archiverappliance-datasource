import React, { PureComponent, ChangeEvent } from 'react';
import { DataSourceHttpSettings, InlineSwitch, InlineField, Select } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, toOption, SelectableValue } from '@grafana/data';
import { AADataSourceOptions, operatorList } from '../types';

const LABEL_WIDTH = 26;
const operatorOptions: Array<SelectableValue<string>> = operatorList.map(toOption);

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

  onOperatorChange = (option: SelectableValue) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      defaultOperator: option.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onUseLiveUpdateChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      useLiveUpdate: !options.jsonData.useLiveUpdate,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onLiveUpdateURIChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      liveUpdateURI: event.target.value,
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
          <div className="gf-form-inline">
            <InlineField
              label="Default Operator"
              labelWidth={LABEL_WIDTH}
              tooltip={
                <p>
                  Controls processing of data during data retrieval. Refer{' '}
                  <a
                    href="https://slacmshankar.github.io/epicsarchiver_docs/userguide.html"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Archiver Appliance User Guide
                  </a>{' '}
                  about processing of data. Special operator <code>raw</code> and <code>last</code> are also available.{' '}
                  <code>raw</code> allows to retrieve the data without processing. <code>last</code> allows to retrieve
                  the last data in the specified time range.
                </p>
              }
            >
              <Select
                value={options.jsonData.defaultOperator}
                options={operatorOptions}
                onChange={this.onOperatorChange}
                placeholder="mean"
              />
            </InlineField>
          </div>
          <div className="gf-form-inline">
            <InlineField
              label="Use live feature (Alpha)"
              labelWidth={LABEL_WIDTH}
              tooltip="(Caution) This is a alpha feature. Live feature provides live updating with PVWS WebSocket server."
            >
              <InlineSwitch
                value={options.jsonData.useLiveUpdate ?? false}
                onChange={this.onUseLiveUpdateChange}
              />
            </InlineField>
          </div>
          <div className="gf-form-inline">
            <InlineField
              label="PVWS URI (Alpha)"
              labelWidth={LABEL_WIDTH}
              tooltip="URI for PVWS WebSocket server."
            >
              <input
                type="text"
                value={options.jsonData.liveUpdateURI}
                className="gf-form-input max-width-15"
                placeholder="ws://localhost:8080/pvws/pv"
                onChange={this.onLiveUpdateURIChange}
              />
            </InlineField>
          </div>
        </div>
      </>
    );
  }
}
