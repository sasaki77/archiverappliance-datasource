import React, { ChangeEvent } from 'react';
import zip from 'lodash/zip';

import { FunctionDescriptor } from '../types';

interface FunctionParamsProps {
  func: FunctionDescriptor;
  index: number;
  onChange: (func: FunctionDescriptor, index: number) => void;
  onRunQuery: () => void;
}

class FunctionParams extends React.PureComponent<FunctionParamsProps> {
  constructor(props: FunctionParamsProps) {
    super(props);
  }

  onChange(paramIndex: number, event: ChangeEvent<HTMLInputElement>) {
    const { func, index } = this.props;
    const { params } = func;
    const newParams = [...params];
    newParams.splice(paramIndex, 1, event.target.value);
    const newFunc = { ...func, params: newParams };
    this.props.onChange(newFunc, index);
  }

  render() {
    const { func, onRunQuery } = this.props;
    const { params, def } = func;
    const paramArray = zip(params, def.params);
    return (
      <>
        {paramArray &&
          paramArray.map(([param, paramDef], paramIndex) => (
            <>
              {paramIndex > 0 ? <span className="comma">,</span> : ''}
              <input
                value={param}
                key={paramIndex}
                className="input-small tight-form-func-param"
                placeholder={paramDef ? paramDef.name : ''}
                onChange={e => this.onChange(paramIndex, e)}
                onBlur={onRunQuery}
              ></input>
            </>
          ))}
      </>
    );
  }
}

export { FunctionParams };
