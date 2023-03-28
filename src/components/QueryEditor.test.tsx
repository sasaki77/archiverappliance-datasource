import defaults from 'lodash/defaults';
import React from 'react';
import { render } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';
import { DataSource } from '../DataSource';
import { AAQuery, defaultQuery } from '../types';

const setup = (propOverrides?: object) => {
  const datasourceMock: unknown = {
    createQuery: jest.fn((q) => q),
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
    stream: false,
    strmInt: '',
    strmCap: '',
  };

  const props: any = {
    datasource,
    onChange,
    onRunQuery,
    query,
  };

  Object.assign(props, propOverrides);

  return render(<QueryEditor {...props} />);
};

describe('Render Editor with basic options', () => {
  it('should render normally', () => {
    const wrapper = setup();
    expect(wrapper).toMatchSnapshot();
  });

  it('should render regex mode', () => {
    const props = defaults({ regex: true }, defaultQuery);
    const wrapper = setup({ query: props });
    expect(wrapper).toMatchSnapshot();
  });

  it('should render with alias pattern', () => {
    const props = defaults({ aliasPattern: '.*' }, defaultQuery);
    const wrapper = setup({ query: props });
    expect(wrapper).toMatchSnapshot();
  });

  it('should render with stream', () => {
    const props = defaults({ stream: true }, defaultQuery);
    const wrapper = setup({ query: props });
    expect(wrapper).toMatchSnapshot();
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
    const wrapper = setup({ query: props });
    expect(wrapper).toMatchSnapshot();
  });
});
