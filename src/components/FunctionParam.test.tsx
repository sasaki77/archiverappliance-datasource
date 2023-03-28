import React from 'react';
import { render } from '@testing-library/react';

import { FunctionParam, FunctionParamProps } from './FunctionParam';

const setup = (propOverrides?: object) => {
  const props: FunctionParamProps = {
    param: 'avg',
    paramDef: {
      name: 'value',
      options: ['avg', 'min', 'max', 'absoluteMin', 'absoluteMax', 'sum'],
      type: 'string',
    },
    index: 1,
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return render(<FunctionParam {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(wrapper).toMatchSnapshot();
  });
});
