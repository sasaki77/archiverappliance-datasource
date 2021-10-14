import React from 'react';
import { reduce, each } from 'lodash';
import { ButtonCascader, CascaderOption } from '@grafana/ui';
import { getCategories, getFuncDef } from '../aafunc';
import { FuncDef } from '../types';

export interface FunctionAddProps {
  addFunc: (func: FuncDef) => void;
}

const getAllFunctionNames = (categories: { [key: string]: FuncDef[] }) => {
  const allNames: CascaderOption[] = reduce(
    categories,
    (list, category, key) => {
      const nlist: CascaderOption[] = [];
      each(category, (func) => nlist.push({ label: func.name, value: func.name }));
      list.push({ label: key, value: key, children: nlist });
      return list;
    },
    [] as CascaderOption[]
  );

  return allNames;
};

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
      <ButtonCascader options={allFunctions} value={undefined} onChange={this.onChange}>
        Add
      </ButtonCascader>
    );
  }
}

export { FunctionAdd };
