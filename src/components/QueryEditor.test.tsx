import defaults from 'lodash/defaults';
import React from 'react';
import { shallow } from 'enzyme';
import toJson from 'enzyme-to-json';
import { QueryEditor } from './QueryEditor';
import { DataSource } from '../DataSource';
import { AAQuery, defaultQuery } from '../types';

const setup = (propOverrides?: object) => {
  const datasourceMock: unknown = {
    createQuery: jest.fn(q => q),
    pvNamesFindQuery: jest.fn((q, num) => ['PV']),
  };
  const datasource: DataSource = datasourceMock as DataSource;
  const onRunQuery = jest.fn();
  const onChange = jest.fn();
  const query: AAQuery = {
    target: 'pvname',
    alias: 'alias',
    operator: 'mean',
    regex: false,
    aliasPattern: '',
    functions: [],
    refId: 'A',
  };

  const props: any = {
    datasource,
    onChange,
    onRunQuery,
    query,
  };

  Object.assign(props, propOverrides);

  const wrapper = shallow(<QueryEditor {...props} />);
  const instance = wrapper.instance() as QueryEditor;

  return {
    instance,
    wrapper,
  };
};

describe('Render Editor with basic options', () => {
  it('should render normally', () => {
    const { wrapper } = setup();
    expect(toJson(wrapper)).toMatchSnapshot();
  });

  it('should render regex mode', () => {
    const props = defaults({ regex: true }, defaultQuery);
    const { wrapper } = setup({ query: props });
    expect(toJson(wrapper)).toMatchSnapshot();
  });

  it('should render with alias pattern', () => {
    const props = defaults({ aliasPattern: '.*' }, defaultQuery);
    const { wrapper } = setup({ query: props });
    expect(toJson(wrapper)).toMatchSnapshot();
  });

  it('should render with top function', () => {
    const func = [
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
    ];
    const props = defaults({ functions: func }, defaultQuery);
    const { wrapper } = setup({ query: props });
    expect(toJson(wrapper)).toMatchSnapshot();
  });
});
