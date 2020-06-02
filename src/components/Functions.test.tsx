import React from 'react';
import { shallow } from 'enzyme';
import toJson from 'enzyme-to-json';

import { Functions, FunctionsProps } from './Functions';

const setup = (propOverrides?: object) => {
  const props: FunctionsProps = {
    funcs: [],
    onChange: jest.fn(),
    onRunQuery: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return shallow(<Functions {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(toJson(wrapper)).toMatchSnapshot();
  });

  it('should render with functions', () => {
    const funcs = [
      {
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
        text: 'top(5, avg)',
      },
      {
        def: {
          category: 'Transform',
          defaultParams: [100],
          name: 'scale',
          params: [
            {
              name: 'factor',
              type: 'float',
            },
          ],
        },
        params: ['0.001'],
        text: 'scale(0.001)',
      },
    ];
    const wrapper = setup({ funcs: funcs });
    expect(toJson(wrapper)).toMatchSnapshot();
  });
});
