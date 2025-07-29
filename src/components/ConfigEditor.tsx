import React, { PureComponent, ChangeEvent } from 'react';
import { DataSourceHttpSettings, InlineSwitch, InlineField, Combobox, ComboboxOption } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { AADataSourceOptions, operatorList } from '../types';
import { toComboboxOption } from './utils';

const LABEL_WIDTH = 26;
const operatorOptions: Array<ComboboxOption<string>> = operatorList.map(toComboboxOption);

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

  onOperatorChange = (option: ComboboxOption) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      defaultOperator: option.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onHideInvalidChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      hideInvalid: !options.jsonData.hideInvalid,
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
        <InlineField
          label="Use Backend"
          labelWidth={LABEL_WIDTH}
          tooltip="Checking this option will enable the data retrieval with backend. The archived data is retrieved and processed on Grafana server, then the data is sent to Grafana client."
          interactive={true}
        >
          <InlineSwitch value={options.jsonData.useBackend ?? false} onChange={this.onUseBEChange} />
        </InlineField>
        <InlineField
          label="Default Operator"
          labelWidth={LABEL_WIDTH}
          interactive={true}
          tooltip={
            <p>
              Controls processing of data during data retrieval. Refer{' '}
              <a
                href="https://epicsarchiver.readthedocs.io/en/latest/user/userguide.html#processing-of-data"
                target="_blank"
                rel="noopener noreferrer"
              >
                Archiver Appliance User Guide
              </a>{' '}
              about processing of data. Special operator <code>raw</code> and <code>last</code> are also available.{' '}
              <code>raw</code> allows to retrieve the data without processing. <code>last</code> allows to retrieve the
              last data in the specified time range.
            </p>
          }
        >
          <Combobox
            value={options.jsonData.defaultOperator}
            options={operatorOptions}
            onChange={this.onOperatorChange}
            placeholder="mean"
          />
        </InlineField>
        <InlineField
          label="Hide Invalid"
          labelWidth={LABEL_WIDTH}
          interactive={true}
          tooltip={<p>Hide sample data whose severity is invalid with a null value.</p>}
        >
          <InlineSwitch value={options.jsonData.hideInvalid ?? false} onChange={this.onHideInvalidChange} />
        </InlineField>
        <InlineField
          label="Use live feature (Alpha)"
          labelWidth={LABEL_WIDTH}
          interactive={true}
          tooltip="(Caution) This is a alpha feature. Live feature provides live updating with PVWS WebSocket server."
        >
          <InlineSwitch value={options.jsonData.useLiveUpdate ?? false} onChange={this.onUseLiveUpdateChange} />
        </InlineField>
        <InlineField
          label="PVWS URI (Alpha)"
          labelWidth={LABEL_WIDTH}
          interactive={true}
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
      </>
    );
  }
}
