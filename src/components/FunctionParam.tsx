import React, { KeyboardEvent, FormEvent, FocusEvent } from 'react';
import Autosuggest from 'react-autosuggest';

export interface FunctionParamProps {
  param: string;
  paramDef: { name: string; options?: string[]; type: string };
  index: number;
  onChange: (paramIndex: number, newValue: string) => void;
  onRunQuery: () => void;
}

interface State {
  suggestions: any[];
  focused: boolean;
}

const getSuggestionValue = (suggestion: any) => {
  return suggestion;
};

const renderSuggestion = (suggestion: any) => {
  return <span>{suggestion}</span>;
};

class FunctionParam extends React.PureComponent<FunctionParamProps, State> {
  state = { suggestions: [], focused: false };

  constructor(props: FunctionParamProps) {
    super(props);
  }

  onChange = (paramIndex: number, event: FormEvent<any>, { newValue }: any) => {
    this.props.onChange(paramIndex, String(newValue));
  };

  onKeydownEnter = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.keyCode === 13) {
      event.currentTarget.blur();
    }
  };

  onPVSuggestionsFetchRequested = ({ value }: { value: any }, options: string[] | undefined) => {
    const suggestions = options === undefined ? [] : options;
    this.setState({
      suggestions: suggestions,
    });
  };

  onPVSuggestionsClearRequested = () => {
    this.setState({
      suggestions: [],
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

  onFocus = (event: FocusEvent<any>) => {
    this.setState({
      focused: true,
    });
  };

  onBlur = (event: FocusEvent<any>) => {
    const { onRunQuery } = this.props;
    this.setState({
      focused: false,
    });
    onRunQuery();
  };

  calcInputWidth = (focused: boolean, param: string, placeholder: string) => {
    if (focused) {
      return '90px';
    }

    const paramLength = param.length;
    return paramLength < 1 ? `${placeholder.length}ch` : `${paramLength}ch`;
  };

  render() {
    const { suggestions, focused } = this.state;
    const { param, paramDef, index } = this.props;
    const inputWidth = this.calcInputWidth(focused, param, paramDef.name);
    return (
      <div className="aa-func-param">
        <Autosuggest
          suggestions={suggestions}
          onSuggestionsFetchRequested={e => this.onPVSuggestionsFetchRequested(e, paramDef.options)}
          onSuggestionsClearRequested={this.onPVSuggestionsClearRequested}
          onSuggestionSelected={this.onSuggestionsSelected}
          getSuggestionValue={getSuggestionValue}
          renderSuggestion={renderSuggestion}
          inputProps={{
            value: param,
            className: 'tight-form-func-param',
            placeholder: paramDef ? paramDef.name : '',
            spellCheck: false,
            onChange: (e, params) => this.onChange(index, e, params),
            onBlur: this.onBlur,
            onKeyDown: this.onKeydownEnter,
            onFocus: this.onFocus,
            style: { width: inputWidth, fontFamily: 'Consolas, "Courier New", Courier, Monaco, monospace' },
          }}
        />
      </div>
    );
  }
}

export { FunctionParam };
