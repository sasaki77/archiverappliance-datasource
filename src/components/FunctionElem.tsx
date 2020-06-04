import React from 'react';
import { FunctionDescriptor } from '../types';

import { FunctionEditor } from './FunctionEditor';
import { FunctionEditorControlsProps } from './FunctionEditorControls';
import { FunctionParams } from './FunctionParams';

export interface FunctionElemProps extends FunctionEditorControlsProps {
  func: FunctionDescriptor;
  index: number;
  onChange: (func: FunctionDescriptor, index: number) => void;
  onRunQuery: () => void;
}

class FunctionElem extends React.PureComponent<FunctionElemProps> {
  constructor(props: FunctionElemProps) {
    super(props);
  }

  render() {
    const { onMoveLeft, onMoveRight, onRemove, func, index, onChange, onRunQuery } = this.props;
    return (
      <span className="gf-form-label query-part">
        <FunctionEditor
          func={func}
          index={index}
          onMoveLeft={onMoveLeft}
          onMoveRight={onMoveRight}
          onRemove={onRemove}
        />
        <span>(</span>
        <FunctionParams func={func} index={index} onChange={onChange} onRunQuery={onRunQuery} />
        <span>)</span>
      </span>
    );
  }
}

export { FunctionElem };
