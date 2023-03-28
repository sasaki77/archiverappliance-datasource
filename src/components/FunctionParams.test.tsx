import React from 'react';
import { render } from '@testing-library/react';

import { FunctionParams, FunctionParamsProps } from './FunctionParams';

const setup = (propOverrides?: object) => {
  const props: FunctionParamsProps = {
    func: {
      def: {
        category: 'Filter Series',
        defaultParams: [5, 'avg'],
        name: 'top',
        params: [
          {
            name: 'number',
            type: 'int',
          },
          {
            name: 'value',
            options: ['avg', 'min', 'max', 'absoluteMin', 'absoluteMax', 'sum'],
            type: 'string',
          },
        ],
      },
      params: ['5', 'avg'],
    },
    index: 1,
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return render(<FunctionParams {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(wrapper).toMatchSnapshot();
  });
});
