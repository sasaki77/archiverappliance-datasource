import defaults from 'lodash/defaults';
import React, { ChangeEvent, PureComponent, KeyboardEvent } from 'react';
import { InlineFormLabel, InlineSwitch } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';
import Autosuggest from 'react-autosuggest';
import { DataSource } from '../DataSource';
import { AADataSourceOptions, AAQuery, defaultQuery, operatorList, FunctionDescriptor } from '../types';

import { Functions } from './Functions';

type Props = QueryEditorProps<DataSource, AAQuery, AADataSourceOptions>;

interface State {
  pvSuggestions: any[];
  oprSuggestions: any[];
}

const colorYellow = '#d69e2e';

const getSuggestionValue = (suggestion: any) => {
  return suggestion;
};

const renderSuggestion = (suggestion: any) => {
  return <span>{suggestion}</span>;
};

export class QueryEditor extends PureComponent<Props, State> {
  state = { pvSuggestions: [], oprSuggestions: [] };

  onPVChange = (event: React.FormEvent<HTMLElement>, { newValue }: any) => {
    const { onChange, query } = this.props;
    onChange({ ...query, target: newValue });
  };

  onRegexChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onChange, onRunQuery, query } = this.props;
    onChange({ ...query, regex: !query.regex });
    onRunQuery();
  };

  onOperatorChange = (event: React.FormEvent<HTMLElement>, { newValue }: any) => {
    const { onChange, query } = this.props;
    onChange({ ...query, operator: newValue });
  };

  onAliasChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, alias: event.target.value });
  };

  onAliaspatternChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, aliasPattern: event.target.value });
  };

  onFuncsChange = (funcs: FunctionDescriptor[]) => {
    const { onChange, query } = this.props;
    onChange({ ...query, functions: funcs });
  };

  onStreamChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onChange, onRunQuery, query } = this.props;
    onChange({ ...query, stream: !query.stream });
    onRunQuery();
  };

  onStrmIntChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, strmInt: event.target.value });
  };

  onStrmCapChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, strmCap: event.target.value });
  };

  onKeydownEnter = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.keyCode === 13) {
      event.currentTarget.blur();
    }
  };

  loadPVSuggestions(value: string) {
    const templateSrv = getTemplateSrv();
    const replacedQuery = templateSrv.replace(value, undefined, 'regex');
    const { regex } = this.props.query;
    const searchQuery = regex ? replacedQuery : `.*${replacedQuery}.*`;
    this.props.datasource.pvNamesFindQuery(searchQuery, 100).then((res: any) => {
      this.setState({
        pvSuggestions: res,
      });
    });
  }

  onPVSuggestionsFetchRequested = ({ value }: { value: any }) => {
    this.loadPVSuggestions(value);
  };

  onPVSuggestionsClearRequested = () => {
    this.setState({
      pvSuggestions: [],
    });
  };

  onOprSuggestionsFetchRequested = ({ value }: { value: any }) => {
    const escapedValue = value.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp('.*' + escapedValue, 'i');

    const suggestions = operatorList.filter((operator) => regex.test(operator));

    this.setState({
      oprSuggestions: suggestions,
    });
  };

  onOprSuggestionsClearRequested = () => {
    this.setState({
      oprSuggestions: [],
    });
  };

  onSuggestionsSelected = (
    event: React.FormEvent<any>,
    {
      suggestion,
      suggestionValue,
      suggestionIndex,
      sectionIndex,
      method,
    }: {
      suggestion: any;
      suggestionValue: string;
      suggestionIndex: number;
      sectionIndex: number | null;
      method: 'click' | 'enter';
    }
  ) => {
    if (method === 'enter') {
      event.currentTarget.blur();
    }
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { pvSuggestions, oprSuggestions } = this.state;
    const { onRunQuery } = this.props;
    const pvInputStyle = query.regex ? { color: colorYellow } : {};
    const aliasInputStyle = query.aliasPattern ? { color: colorYellow } : {};

    return (
      <>
        <div className="gf-form-inline">
          <InlineFormLabel
            width={6}
            className="query-keyword"
            tooltip={
              <p>
                Set PV name to be visualized. It is allowed to set multiple PVs by using Regular Expressoins alternation
                pattern (e.g. <code>(PV:1|PV:2)</code>)
              </p>
            }
          >
            PV
          </InlineFormLabel>
          <div className="max-width-30 gf-form-spacing">
            <Autosuggest
              suggestions={pvSuggestions}
              onSuggestionsFetchRequested={this.onPVSuggestionsFetchRequested}
              onSuggestionsClearRequested={this.onPVSuggestionsClearRequested}
              getSuggestionValue={getSuggestionValue}
              onSuggestionSelected={this.onSuggestionsSelected}
              renderSuggestion={renderSuggestion}
              shouldRenderSuggestions={() => true}
              inputProps={{
                value: query.target,
                className: 'gf-form-input',
                placeholder: 'PV name',
                spellCheck: false,
                style: pvInputStyle,
                onChange: this.onPVChange,
                onBlur: onRunQuery,
                onKeyDown: this.onKeydownEnter,
              }}
            />
          </div>
          <InlineFormLabel
            width={7}
            className="query-keyword"
            tooltip={
              <p>
                Enable/Disable Regex mode. You can select multiple PVs using Regular Expressoins.
              </p>
            }
          >
            Regex
          </InlineFormLabel>
          <InlineSwitch
            value={query.regex}
            onChange={this.onRegexChange}
          />
        </div>
        <div className="gf-form-inline">
          <InlineFormLabel
            width={6}
            className="query-keyword"
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
            Operator
          </InlineFormLabel>
          <div className="max-width-30 gf-form-spacing">
            <Autosuggest
              suggestions={oprSuggestions}
              onSuggestionsFetchRequested={this.onOprSuggestionsFetchRequested}
              onSuggestionsClearRequested={this.onOprSuggestionsClearRequested}
              onSuggestionSelected={this.onSuggestionsSelected}
              getSuggestionValue={getSuggestionValue}
              renderSuggestion={renderSuggestion}
              shouldRenderSuggestions={() => true}
              inputProps={{
                value: query.operator,
                className: 'gf-form-input',
                placeholder: 'mean',
                spellCheck: false,
                onChange: this.onOperatorChange,
                onBlur: onRunQuery,
                onKeyDown: this.onKeydownEnter,
              }}
            />
          </div>
          <InlineFormLabel
            width={7}
            className="query-keyword"
            tooltip={
              <p>
                Stream allows to periodically update the data without refreshing the dashboard. The difference data from the last updated values is only retrieved.
              </p>
            }
          >
            Stream
          </InlineFormLabel>
          <InlineSwitch
            value={query.stream}
            onChange={this.onStreamChange}
            className="gf-form-spacing"
          />
          <InlineFormLabel
            width={6}
            className="query-keyword"
            tooltip={
              <p>
                Streaming interval in milliseconds. You can also use a number with unit. e.g. <code>1s</code>,{' '}
                <code>1m</code>, <code>1h</code>. The default is determined by width of panel and time range.
              </p>
            }
          >
            Interval
          </InlineFormLabel>
          <input
            type="text"
            value={query.strmInt}
            className="gf-form-input max-width-7"
            placeholder="auto"
            onChange={this.onStrmIntChange}
            onBlur={onRunQuery}
            onKeyDown={this.onKeydownEnter}
          />
          <InlineFormLabel
            width={6}
            className="query-keyword"
            tooltip={
              <p>
                The stream data is stored in a circular buffer. Capacity determines the buffer size. The default is
                detemined by initial data size.
              </p>
            }
          >
            Capacity
          </InlineFormLabel>
          <input
            type="text"
            value={query.strmCap}
            className="gf-form-input max-width-7"
            placeholder="auto"
            onChange={this.onStrmCapChange}
            onBlur={onRunQuery}
            onKeyDown={this.onKeydownEnter}
          />
        </div>
        <div className="gf-form">
          <InlineFormLabel width={6} className="query-keyword" tooltip={<p>Set alias for the legend.</p>}>
            Alias
          </InlineFormLabel>
          <input
            type="text"
            value={query.alias}
            className="gf-form-input max-width-30"
            placeholder="Alias"
            style={aliasInputStyle}
            onChange={this.onAliasChange}
            onBlur={onRunQuery}
            onKeyDown={this.onKeydownEnter}
          />
          <InlineFormLabel
            width={7}
            className="query-keyword"
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
            Pattern
          </InlineFormLabel>
          <input
            type="text"
            value={query.aliasPattern}
            className="gf-form-input max-width-30"
            placeholder="Alias regex pattern"
            style={{ color: colorYellow }}
            onChange={this.onAliaspatternChange}
            onBlur={onRunQuery}
            onKeyDown={this.onKeydownEnter}
          />
        </div>
        <Functions funcs={query.functions} onChange={this.onFuncsChange} onRunQuery={onRunQuery} />
      </>
    );
  }
}
