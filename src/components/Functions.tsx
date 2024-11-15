import React from 'react';
import { InlineFieldRow, InlineLabel } from '@grafana/ui';
import { createFuncDescriptor } from '../aafunc';
import { FuncDef, FunctionDescriptor } from '../types';

import { FunctionElem } from './FunctionElem';
import { FunctionAdd } from './FunctionAdd';

export interface FunctionsProps {
  funcs: FunctionDescriptor[];
  onChange: (func: FunctionDescriptor[]) => void;
  onRunQuery: () => void;
}

const arrayMove = (array: any[], from: number, to: number): any[] => {
  const newArray = [...array];
  const moveElem = newArray.splice(from, 1)[0];
  newArray.splice(to, 0, moveElem);

  return newArray;
};

class Functions extends React.PureComponent<FunctionsProps> {
  constructor(props: FunctionsProps) {
    super(props);
  }

  addFunction = (funcDef: FuncDef) => {
    const { onChange, onRunQuery } = this.props;
    const newFunc = createFuncDescriptor(funcDef, funcDef.defaultParams);
    const newFuncs = [...this.props.funcs];
    newFuncs.push(newFunc);
    onChange(newFuncs);

    if (newFunc.params.length || newFunc.def.params.length === 0) {
      onRunQuery();
    }
  };

  handleRemoveFunction = (func: FunctionDescriptor, index: number) => {
    const { onChange, onRunQuery } = this.props;
    const newFuncs = [...this.props.funcs];
    newFuncs.splice(index, 1);
    onChange(newFuncs);
    onRunQuery();
  };

  handleMoveLeft = (func: FunctionDescriptor, index: number) => {
    if (index < 1) {
      return;
    }

    const { onChange, onRunQuery } = this.props;
    const newFuncs = arrayMove(this.props.funcs, index, index - 1);
    onChange(newFuncs);
    onRunQuery();
  };

  handleMoveRight = (func: FunctionDescriptor, index: number) => {
    if (index >= this.props.funcs.length - 1) {
      return;
    }

    const { onChange, onRunQuery } = this.props;
    const newFuncs = arrayMove(this.props.funcs, index, index + 1);
    onChange(newFuncs);
    onRunQuery();
  };

  onFuncChange = (func: FunctionDescriptor, index: number) => {
    const { onChange } = this.props;
    const newFuncs = [...this.props.funcs];
    newFuncs.splice(index, 1, func);
    onChange(newFuncs);
  };

  render() {
    const { funcs, onRunQuery } = this.props;
    return (
      <InlineFieldRow>
        <InlineLabel
          width={12}
          interactive={true}
          tooltip={
            <p>
              Functions are used to apply post processing to the data. You can add, move and remove functions. Select
              preffered functions from categoires. Some functions require parameters. You can edit parameters after
              adding the function. Functions are applied from left to right.
            </p>
          }
        >
          Functions
        </InlineLabel>
        {funcs &&
          funcs.map((func, index) => (
            <FunctionElem
              key={index}
              index={index}
              func={func}
              onMoveLeft={this.handleMoveLeft}
              onMoveRight={this.handleMoveRight}
              onRemove={this.handleRemoveFunction}
              onChange={this.onFuncChange}
              onRunQuery={onRunQuery}
            />
          ))}
        <FunctionAdd addFunc={this.addFunction} />
        <div className="gf-form-label gf-form-label--grow"></div>
      </InlineFieldRow>
    );
  }
}

export { Functions };
