import React from 'react';
import { shallow } from 'enzyme';
import toJson from 'enzyme-to-json';

import { FunctionElem, FunctionElemProps } from './FunctionElem';

const setup = (propOverrides?: object) => {
  const props: FunctionElemProps = {
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
    onMoveLeft: jest.fn(),
    onMoveRight: jest.fn(),
    onRemove: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return shallow(<FunctionElem {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(toJson(wrapper)).toMatchSnapshot();
  });
});
