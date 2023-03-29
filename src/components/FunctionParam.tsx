import React from 'react';
import { SegmentInput, Segment } from '@grafana/ui';
import { toSelectableValue } from './toSelectableValue';

export interface FunctionParamProps {
  param: string;
  paramDef: { name: string; options?: string[]; type: string };
  index: number;
  onChange: (paramIndex: number, newValue: string) => void;
  onRunQuery: () => void;
}

class FunctionParam extends React.PureComponent<FunctionParamProps> {

  constructor(props: FunctionParamProps) {
    super(props);
  }

  onChange = (paramIndex: number, value: any) => {
    console.log("test");
    const { onRunQuery } = this.props;
    this.props.onChange(paramIndex, String(value));
    onRunQuery();
  };

  render() {
    const { param, paramDef, index } = this.props;

    if (paramDef.options === undefined) {
      return (
        <SegmentInput
          value={param}
          placeholder={paramDef.name}
          onChange={(text) => {
            this.onChange(index, text);
          }}
        />
      );
    }

    const options = paramDef.options.map(toSelectableValue);

    return (
      <Segment
        value={param}
        options={options}
        placeholder="Enter a value"
        onChange={(item) => {
          this.onChange(index, item.value);
        }}
        width={param.length}
      />
    );
  }
}

export { FunctionParam };
