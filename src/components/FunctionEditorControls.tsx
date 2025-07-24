import React from 'react';
import { Icon } from '@grafana/ui';
import { FunctionDescriptor } from '../types';

declare const __VERSION__: string;

export interface FunctionEditorControlsProps {
  onMoveLeft: (func: FunctionDescriptor, index: number) => void;
  onMoveRight: (func: FunctionDescriptor, index: number) => void;
  onRemove: (func: FunctionDescriptor, index: number) => void;
}

const FunctionHelpButton = (props: {
  description: string | undefined;
  name: string;
  onDescriptionShow: () => void;
}) => {
  if (props.description) {
    return <span className="pointer fa fa-question-circle" onClick={props.onDescriptionShow} />;
  }

  return (
    <Icon
      className="pointer"
      name="question-circle"
      onClick={() => {
        window.open(
          `https://sasaki77.github.io/archiverappliance-datasource/${__VERSION__}/functions.html#${props.name}`,
          '_blank'
        );
      }}
    />
  );
};

export const FunctionEditorControls = (
  props: FunctionEditorControlsProps & {
    func: FunctionDescriptor;
    onDescriptionShow: () => void;
    index: number;
  }
) => {
  const { func, index, onMoveLeft, onMoveRight, onRemove, onDescriptionShow } = props;
  return (
    <div
      style={{
        display: 'flex',
        width: '60px',
        justifyContent: 'space-between',
      }}
    >
      <Icon name="arrow-left" onClick={() => onMoveLeft(func, index)} />
      <FunctionHelpButton
        name={func.def.name}
        description={func.def.description}
        onDescriptionShow={onDescriptionShow}
      />
      <Icon name="times" onClick={() => onRemove(func, index)} />
      <Icon name="arrow-right" onClick={() => onMoveRight(func, index)} />
    </div>
  );
};
