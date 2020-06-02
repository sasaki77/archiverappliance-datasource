import React from 'react';
import { shallow } from 'enzyme';
import toJson from 'enzyme-to-json';

import { FunctionAdd, FunctionAddProps } from './FunctionAdd';

const setup = (propOverrides?: object) => {
  const props: FunctionAddProps = {
    addFunc: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return shallow(<FunctionAdd {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(toJson(wrapper)).toMatchSnapshot();
  });
});
