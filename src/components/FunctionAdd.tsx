import React from 'react';
import { ButtonCascader, CascaderOption } from '@grafana/ui';
import { getCategories, getFuncDef } from '../aafunc';
import { FuncDef } from '../types';

export interface FunctionAddProps {
  addFunc: (func: FuncDef) => void;
}

const getAllFunctionNames = (categories: { [key: string]: FuncDef[] }): CascaderOption[] =>
  Object.entries(categories).map(([key, category]) => ({
    label: key,
    value: key,
    children: category.map((func) => ({ label: func.name, value: func.name })),
  }));

class FunctionAdd extends React.PureComponent<FunctionAddProps> {
  constructor(props: FunctionAddProps) {
    super(props);
  }

  onChange = (value: any[], selectedOptions: CascaderOption[]) => {
    if (value.length < 2) {
      return;
    }
    const funcName = value[1];
    this.props.addFunc(getFuncDef(funcName));
  };

  render() {
    const categories = getCategories();
    const allFunctions = getAllFunctionNames(categories);
    return (
      <ButtonCascader options={allFunctions} value={[]} onChange={this.onChange} variant="primary" icon="plus">
        Add
      </ButtonCascader>
    );
  }
}

export { FunctionAdd };
