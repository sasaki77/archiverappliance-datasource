import React from 'react';
import { Tooltip } from '@grafana/ui';
import { FunctionEditorControls, FunctionEditorControlsProps } from './FunctionEditorControls';
import { FunctionDescriptor } from '../types';

interface FunctionEditorProps extends FunctionEditorControlsProps {
  func: FunctionDescriptor;
  index: number;
}

interface FunctionEditorState {
  showingDescription: boolean;
}

class FunctionEditor extends React.PureComponent<FunctionEditorProps, FunctionEditorState> {
  constructor(props: FunctionEditorProps) {
    super(props);

    this.state = {
      showingDescription: false,
    };
  }

  renderContent = ({ updatePopperPosition }: any) => {
    const {
      onMoveLeft,
      onMoveRight,
      func: {
        def: { name, description },
      },
    } = this.props;
    const { showingDescription } = this.state;

    if (showingDescription) {
      return (
        <div style={{ overflow: 'auto', maxHeight: '30rem', textAlign: 'left', fontWeight: 'normal' }}>
          <h4 style={{ color: 'white' }}> {name} </h4>
          <div>{description}</div>
        </div>
      );
    }

    return (
      <FunctionEditorControls
        {...this.props}
        onMoveLeft={() => {
          onMoveLeft(this.props.func, this.props.index);
          updatePopperPosition();
        }}
        onMoveRight={() => {
          onMoveRight(this.props.func, this.props.index);
          updatePopperPosition();
        }}
        onDescriptionShow={() => {
          this.setState({ showingDescription: true }, () => {
            updatePopperPosition();
          });
        }}
      />
    );
  };

  render() {
    return (
      <>
        <Tooltip content={this.renderContent} placement="top" interactive>
          <span style={{ cursor: 'pointer' }}>{this.props.func.def.name}</span>
        </Tooltip>
      </>
    );
  }
}

export { FunctionEditor };
