import defaults from 'lodash/defaults';
import React, { ChangeEvent, PureComponent, KeyboardEvent } from 'react';
import { InlineFormLabel, LegacyForms } from '@grafana/ui';
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

  onPVChange = (event: ChangeEvent<HTMLInputElement>, { newValue }: any) => {
    const { onChange, query } = this.props;
    onChange({ ...query, target: newValue });
  };

  onRegexChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onChange, onRunQuery, query } = this.props;
    onChange({ ...query, regex: !query.regex });
    onRunQuery();
  };

  onOperatorChange = (event: ChangeEvent<HTMLInputElement>, { newValue }: any) => {
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

    const suggestions = operatorList.filter(operator => regex.test(operator));

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
          <InlineFormLabel width={6} className="query-keyword">
            PV
          </InlineFormLabel>
          <div className="max-width-30" style={{ marginRight: '4px' }}>
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
          <LegacyForms.Switch
            checked={query.regex}
            label="Regex"
            labelClass={'width-6  query-keyword'}
            onChange={this.onRegexChange}
          />
        </div>
        <div className="gf-form">
          <InlineFormLabel width={6} className="query-keyword">
            Operator
          </InlineFormLabel>
          <div className="max-width-30">
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
        </div>
        <div className="gf-form">
          <InlineFormLabel width={6} className="query-keyword">
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
          <InlineFormLabel width={6} className="query-keyword">
            Alias pattern
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
