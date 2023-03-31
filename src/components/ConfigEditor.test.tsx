import React from 'react';
import { render } from '@testing-library/react';

import { ConfigEditor, Props } from './ConfigEditor';

const setup = (propOverrides?: object) => {
  const props: Props = {
    options: {
      id: 1,
      orgId: 1,
      uid: '',
      typeLogoUrl: '',
      name: 'ArchiverAppliance',
      access: 'proxy',
      url: '',
      database: '',
      type: 'ArchiverAppliance',
      typeName: 'ArchiverAppliance',
      user: '',
      basicAuth: false,
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

  return render(<ConfigEditor {...props} />);
};

describe('Render', () => {
  it('should render component', () => {
    const wrapper = setup();
    expect(wrapper).toMatchSnapshot();
  });
});
