import React from 'react';
import { render } from '@testing-library/react';

import { FunctionAdd, FunctionAddProps } from './FunctionAdd';

const setup = (propOverrides?: object) => {
  const props: FunctionAddProps = {
    addFunc: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return render(<FunctionAdd {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(wrapper).toMatchSnapshot();
  });
});
