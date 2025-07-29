import React, { PureComponent, ChangeEvent } from 'react';
import { Input, Field, Label, Icon, Tooltip, Combobox, ComboboxOption, Divider, Switch, Stack } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { AADataSourceOptions, operatorList } from '../types';
import { toComboboxOption } from './utils';
import {
  ConfigSection,
  ConnectionSettings,
  Auth,
  AdvancedHttpSettings,
  convertLegacyAuthProps,
  EditorStack,
  ConfigSubSection,
} from '@grafana/plugin-ui';

//const LABEL_WIDTH = 26;
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
        <ConnectionSettings config={options} onChange={onOptionsChange} />

        <Divider spacing={4} />

        <Auth
          {...convertLegacyAuthProps({
            config: options,
            onChange: onOptionsChange,
          })}
        />

        <Divider spacing={4} />

        <ConfigSection title="Advanced settings" isCollapsible>
          <Stack gap={3} direction="column">
            <AdvancedHttpSettings config={options} onChange={onOptionsChange} />
            <ConfigSubSection title="Data Retrieval Options">
              <Field
                label={
                  <Label>
                    <EditorStack gap={0.5}>
                      <span>Use Backend</span>
                      <Tooltip
                        content={
                          <span>
                            Checking this option will enable the data retrieval with backend. The archived data is
                            retrieved and processed on Grafana server, then the data is sent to Grafana client.
                          </span>
                        }
                      >
                        <Icon name="info-circle" size="sm" />
                      </Tooltip>
                    </EditorStack>
                  </Label>
                }
              >
                <Switch value={options.jsonData.useBackend ?? false} onChange={this.onUseBEChange} />
              </Field>

              <Field
                label={
                  <Label>
                    <EditorStack gap={0.5}>
                      <span>Default Operator</span>
                      <Tooltip
                        content={
                          <p>
                            Controls processing of data during data retrieval. Refer{' '}
                            <a
                              href="https://epicsarchiver.readthedocs.io/en/latest/user/userguide.html#processing-of-data"
                              target="_blank"
                              rel="noopener noreferrer"
                            >
                              Archiver Appliance User Guide
                            </a>{' '}
                            <code>raw</code> allows to retrieve the data without processing. <code>last</code> allows to
                            retrieve the last data in the specified time range.
                          </p>
                        }
                      >
                        <Icon name="info-circle" size="sm" />
                      </Tooltip>
                    </EditorStack>
                  </Label>
                }
              >
                <Combobox
                  value={options.jsonData.defaultOperator}
                  options={operatorOptions}
                  width={40}
                  onChange={this.onOperatorChange}
                  placeholder="mean"
                />
              </Field>

              <Field
                label={
                  <Label>
                    <EditorStack gap={0.5}>
                      <span>Hide Invalid</span>
                      <Tooltip content={<span>Hide sample data whose severity is invalid with a null value.</span>}>
                        <Icon name="info-circle" size="sm" />
                      </Tooltip>
                    </EditorStack>
                  </Label>
                }
              >
                <Switch value={options.jsonData.hideInvalid ?? false} onChange={this.onHideInvalidChange} />
              </Field>
            </ConfigSubSection>

            <ConfigSubSection title="Live Feature Options">
              <Field
                label={
                  <Label>
                    <EditorStack gap={0.5}>
                      <span>Use live feature (Alpha)</span>
                      <Tooltip
                        content={
                          <span>
                            (Caution) This is an alpha feature. Live feature provides live updating with PVWS WebSocket
                            server.
                          </span>
                        }
                      >
                        <Icon name="info-circle" size="sm" />
                      </Tooltip>
                    </EditorStack>
                  </Label>
                }
              >
                <Switch value={options.jsonData.useLiveUpdate ?? false} onChange={this.onUseLiveUpdateChange} />
              </Field>

              <Field
                label={
                  <Label>
                    <EditorStack gap={0.5}>
                      <span>PVWS URI (Alpha)</span>
                      <Tooltip content={<span>URI for PVWS WebSocket server.</span>}>
                        <Icon name="info-circle" size="sm" />
                      </Tooltip>
                    </EditorStack>
                  </Label>
                }
              >
                <Input
                  value={options.jsonData.liveUpdateURI}
                  placeholder="ws://localhost:8080/pvws/pv"
                  width={40}
                  onChange={this.onLiveUpdateURIChange}
                />
              </Field>
            </ConfigSubSection>
          </Stack>
        </ConfigSection>
      </>
    );
  }
}
