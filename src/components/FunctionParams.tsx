import React from 'react';
import zip from 'lodash/zip';
import { FunctionParam } from './FunctionParam';

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

  onParamChange = (paramIndex: number, newValue: string) => {
    const { func, index } = this.props;
    const { params } = func;
    const newParams = [...params];
    newParams.splice(paramIndex, 1, newValue);
    const newFunc = { ...func, params: newParams };
    this.props.onChange(newFunc, index);
  };

  render() {
    const { func, onRunQuery } = this.props;
    const { params, def } = func;
    const paramArray = zip(params, def.params);
    return (
      <>
        {paramArray &&
          paramArray.map(([param, paramDef], paramIndex) => {
            if (param === undefined || paramDef === undefined) {
              return;
            }

            return (
              <>
                {paramIndex > 0 ? <span className="comma">,&nbsp;</span> : ''}
                <div>
                  <FunctionParam
                    param={param}
                    paramDef={paramDef}
                    index={paramIndex}
                    onRunQuery={onRunQuery}
                    onChange={this.onParamChange}
                  />
                </div>
              </>
            );
          })}
      </>
    );
  }
}

export { FunctionParams };