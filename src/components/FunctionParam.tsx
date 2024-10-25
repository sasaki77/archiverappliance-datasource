import React from 'react';
import { GrafanaTheme2 } from '@grafana/data';
import { useStyles2 } from '@grafana/ui';
import { css } from '@emotion/css';
import { SegmentInput, Segment } from '@grafana/ui';
import { toSelectableValue } from './toSelectableValue';

export interface FunctionParamProps {
  param: string;
  paramDef: { name: string; options?: string[]; type: string };
  index: number;
  onChange: (paramIndex: number, newValue: string) => void;
  onRunQuery: () => void;
}

export const FunctionParam = ({ param, paramDef, index, onChange, onRunQuery }: FunctionParamProps): JSX.Element => {
  const onInputChange = (paramIndex: number, value: any) => {
    onChange(paramIndex, String(value));
    onRunQuery();
  };

  const styles = useStyles2(getStyles);

  if (paramDef.options === undefined) {
    return (
      <SegmentInput
        value={param}
        placeholder={paramDef.name}
        className={styles.input}
        onChange={(text) => {
          onInputChange(index, text);
        }}
        // input style
        style={{ height: '25px', paddingTop: '2px', marginTop: '2px', paddingLeft: '4px', minWidth: '100px' }}
      />
    );
  }

  const options = paramDef.options.map(toSelectableValue);

  return (
    <Segment
      value={param}
      className={styles.segment}
      options={options}
      placeholder={paramDef.name}
      inputMinWidth={150}
      onChange={(item) => {
        onInputChange(index, item.value);
      }}
      allowCustomValue
      width={param.length}
    />
  );
};

const getStyles = (theme: GrafanaTheme2) => ({
  segment: css({
    margin: 0,
    padding: 0,
  }),
  input: css`
    margin: 0;
    padding: 0;
    input {
      height: 25px;
    },
  `,
});
