import React from 'react';
import { shallow } from 'enzyme';
import toJson from 'enzyme-to-json';

import { ConfigEditor, Props } from './ConfigEditor';

const setup = (propOverrides?: object) => {
  const props: Props = {
    options: {
      id: 1,
      orgId: 1,
      typeLogoUrl: '',
      name: 'CloudWatch',
      access: 'proxy',
      url: '',
      database: '',
      type: 'cloudwatch',
      user: '',
      password: '',
      basicAuth: false,
      basicAuthPassword: '',
      basicAuthUser: '',
      isDefault: true,
      readOnly: false,
      withCredentials: false,
      secureJsonFields: {
        accessKey: false,
        secretKey: false,
      },
      jsonData: {},
      secureJsonData: {
        secretKey: '',
        accessKey: '',
      },
    },
    onOptionsChange: jest.fn(),
  };

  Object.assign(props, propOverrides);

  return shallow(<ConfigEditor {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(toJson(wrapper)).toMatchSnapshot();
  });
});
