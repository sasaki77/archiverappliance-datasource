import defaults from 'lodash/defaults';
import debounce from 'debounce-promise';
import React, { ChangeEvent, KeyboardEvent, useState } from 'react';
import { InlineSwitch, Input, InlineField, Combobox, ComboboxOption } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { InlineFieldRow } from '@grafana/ui';
import { getTemplateSrv } from '@grafana/runtime';
import { DataSource } from '../DataSource';
import { AADataSourceOptions, AAQuery, defaultQuery, operatorList, FunctionDescriptor } from '../types';

import { Functions } from './Functions';
import { toComboboxOption } from './utils';

type Props = QueryEditorProps<DataSource, AAQuery, AADataSourceOptions>;

const colorYellow = '#d69e2e';
const operatorOptions: Array<ComboboxOption<string>> = operatorList.map(toComboboxOption);

export const QueryEditor = ({ query, onChange, onRunQuery, datasource }: Props): React.JSX.Element => {
  const defaultPvOption = query.target ? toComboboxOption(query.target) : undefined;
  const defaultOperatorOption = query.operator && query.operator != '' ? toComboboxOption(query.operator) : undefined;

  const [pvOptionValue, setPVOptionValue] = useState(defaultPvOption);
  const [operatorOptionValue, setOperatorOptionValue] = useState(defaultOperatorOption);

  const onPVChange = (option: ComboboxOption | null) => {
    if (option === null) {
      onChange({ ...query, target: '' });
      setPVOptionValue(undefined);
    } else {
      onChange({ ...query, target: option.value });
      setPVOptionValue(option);
    }

    onRunQuery();
  };

  const onRegexChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    onChange({ ...query, regex: !query.regex });
    onRunQuery();
  };

  const onLiveChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    onChange({ ...query, live: !query.live });
    onRunQuery();
  };

  const onOperatorChange = (option: ComboboxOption | null) => {
    if (option === null) {
      onChange({ ...query, operator: defaultOperator });
      setOperatorOptionValue(undefined);
    } else {
      onChange({ ...query, operator: option.value });
      setOperatorOptionValue(option);
    }

    onRunQuery();
  };

  const onAliasChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, alias: event.target.value });
  };

  const onAliaspatternChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, aliasPattern: event.target.value });
  };

  const onFuncsChange = (funcs: FunctionDescriptor[]) => {
    onChange({ ...query, functions: funcs });
  };

  const onStreamChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    onChange({ ...query, stream: !query.stream });
    onRunQuery();
  };

  const onStrmIntChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, strmInt: event.target.value });
  };

  const onStrmCapChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, strmCap: event.target.value });
  };

  const onKeydownEnter = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') {
      event.currentTarget.blur();
    }
  };

  const loadPVSuggestions = (value: string) => {
    const templateSrv = getTemplateSrv();
    const replacedQuery = templateSrv.replace(value, undefined, 'regex');
    const { regex } = query;
    const searchQuery = regex ? replacedQuery : `.*${replacedQuery}.*`;
    return datasource.pvNamesFindQuery(searchQuery, 100).then((res: any) => {
      const suggestions: Array<ComboboxOption<string>> = res.map(toComboboxOption);
      return suggestions;
    });
  };
  const debounceLoadSuggestions = debounce((query: string) => loadPVSuggestions(query), 200);

  const query_ = defaults(query, defaultQuery);
  const defaultOperator = datasource.defaultOperator || 'mean';
  const useLiveUpdate = datasource.useLiveUpdate || false;
  const aliasInputStyle = query_.aliasPattern ? { color: colorYellow } : {};

  return (
    <>
      <InlineFieldRow label="PV select">
        <InlineField
          labelWidth={12}
          label={'PV'}
          interactive={true}
          tooltip={
            <p>
              Set PV name to be visualized. It is allowed to set multiple PVs by using Regular Expressoins alternation
              pattern (e.g. <code>(PV:1|PV:2)</code>)
            </p>
          }
        >
          <Combobox
            width={56}
            value={pvOptionValue}
            options={debounceLoadSuggestions}
            createCustomValue
            isClearable
            onChange={onPVChange}
            placeholder="PV name"
          />
        </InlineField>
        <InlineField
          labelWidth={12}
          label={'Regex'}
          interactive={true}
          tooltip={<p>Enable/Disable Regex mode. You can select multiple PVs using Regular Expressoins.</p>}
        >
          <InlineSwitch value={query.regex} onChange={onRegexChange} />
        </InlineField>
        {useLiveUpdate === true && (
          <InlineField labelWidth={12} label={'Live'} interactive={true} tooltip={<p>Enable/Disable Live mode.</p>}>
            <InlineSwitch value={query.live} onChange={onLiveChange} />
          </InlineField>
        )}
      </InlineFieldRow>
      <InlineFieldRow label="Retrieval settings">
        <InlineField
          labelWidth={12}
          label={'Operator'}
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
            width={56}
            value={operatorOptionValue}
            options={operatorOptions}
            onChange={onOperatorChange}
            createCustomValue
            isClearable
            placeholder={defaultOperator}
          />
        </InlineField>
        <InlineField
          labelWidth={12}
          label={'Stream'}
          interactive={true}
          tooltip={
            <p>
              Stream allows to periodically update the data without refreshing the dashboard. The difference data from
              the last updated values is only retrieved.
            </p>
          }
        >
          <InlineSwitch value={query.stream} onChange={onStreamChange} />
        </InlineField>
        <InlineField
          labelWidth={12}
          label={'Interval'}
          interactive={true}
          tooltip={
            <p>
              Streaming interval in milliseconds. You can also use a number with unit. e.g. <code>1s</code>,{' '}
              <code>1m</code>, <code>1h</code>. The default is determined by width of panel and time range.
            </p>
          }
        >
          <Input
            width={10}
            value={query.strmInt}
            placeholder="auto"
            onChange={onStrmIntChange}
            onBlur={onRunQuery}
            onKeyDown={onKeydownEnter}
          />
        </InlineField>
        <InlineField
          labelWidth={12}
          label={'Capacity'}
          interactive={true}
          tooltip={
            <p>
              The stream data is stored in a circular buffer. Capacity determines the buffer size. The default is
              detemined by initial data size.
            </p>
          }
        >
          <Input
            width={10}
            value={query.strmCap}
            placeholder="auto"
            onChange={onStrmCapChange}
            onBlur={onRunQuery}
            onKeyDown={onKeydownEnter}
          />
        </InlineField>
      </InlineFieldRow>
      <InlineFieldRow label="Alias">
        <InlineField labelWidth={12} label={'Alias'} interactive={true} tooltip={<p>Set alias for the legend.</p>}>
          <Input
            value={query.alias}
            width={56}
            placeholder="Alias"
            style={aliasInputStyle}
            onChange={onAliasChange}
            onBlur={onRunQuery}
            onKeyDown={onKeydownEnter}
          />
        </InlineField>
        <InlineField
          labelWidth={12}
          label={'Pattern'}
          interactive={true}
          tooltip={
            <p>
              Set regular expressoin pattern to use PV name for legend alias. Alias pattern is used to match PV name.
              Matched characters within parentheses can be used in <code>Alias</code> text input like <code>$1</code>,{' '}
              <code>$2</code>, …, <code>$n</code>. Refer the{' '}
              <a
                href="https://sasaki77.github.io/archiverappliance-datasource/query.html#legend-alias-with-regex-pattern"
                target="_blank"
                rel="noopener noreferrer"
              >
                documentation
              </a>{' '}
              for more detail.
            </p>
          }
        >
          <Input
            value={query.aliasPattern}
            width={52}
            placeholder="Alias regex pattern"
            style={{ color: colorYellow }}
            onChange={onAliaspatternChange}
            onBlur={onRunQuery}
            onKeyDown={onKeydownEnter}
          />
        </InlineField>
      </InlineFieldRow>
      <Functions funcs={query.functions} onChange={onFuncsChange} onRunQuery={onRunQuery} />
    </>
  );
};
