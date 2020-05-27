import React, { ChangeEvent, PureComponent, KeyboardEvent } from 'react';
import { InlineFormLabel, LegacyForms } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../DataSource';
import { AADataSourceOptions, AAQuery } from '../types';
import { Functions } from './Functions';
import { FunctionDescriptor } from './FunctionEditorControls';

type Props = QueryEditorProps<DataSource, AAQuery, AADataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  onPVChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, target: event.target.value });
  };

  onRegexChange = (event: React.SyntheticEvent<HTMLInputElement>) => {
    const { onChange, onRunQuery, query } = this.props;
    onChange({ ...query, regex: !query.regex });
    onRunQuery();
  };

  onOperatorChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, operator: event.target.value });
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

  render() {
    const { query, onRunQuery } = this.props;

    return (
      <>
        <div className="gf-form-inline">
          <InlineFormLabel width={6} className="query-keyword">
            PV
          </InlineFormLabel>
          <input
            type="text"
            value={query.target}
            className="gf-form-input max-width-30"
            placeholder="PV name"
            onChange={this.onPVChange}
            onBlur={onRunQuery}
            onKeyDown={this.onKeydownEnter}
          />
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
          <input
            type="text"
            value={query.operator}
            className="gf-form-input max-width-30"
            placeholder="mean"
            onChange={this.onOperatorChange}
            onBlur={onRunQuery}
            onKeyDown={this.onKeydownEnter}
          />
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
