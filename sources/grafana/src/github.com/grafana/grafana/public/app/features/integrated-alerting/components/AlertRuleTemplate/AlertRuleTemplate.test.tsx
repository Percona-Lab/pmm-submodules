import React from 'react';
import { mount, ReactWrapper } from 'enzyme';
import { dataQa } from '@percona/platform-core';
import { act } from 'react-dom/test-utils';
import { AlertRuleTemplate } from './AlertRuleTemplate';
import { AlertRuleTemplateService } from './AlertRuleTemplate.service';
import { templateStubs } from './__mocks__/alertRuleTemplateStubs';

jest.mock('./AlertRuleTemplate.service', () => ({
  AlertRuleTemplateService: {
    list: () => ({
      templates: templateStubs,
    }),
  },
}));
jest.mock('@percona/platform-core', () => {
  const originalModule = jest.requireActual('@percona/platform-core');
  return {
    ...originalModule,
    logger: {
      error: jest.fn(),
    },
  };
});

describe('AlertRuleTemplate', () => {
  afterEach(() => {
    jest.clearAllMocks();
  });

  it('should render add modal', async () => {
    let wrapper: ReactWrapper;

    await act(async () => {
      wrapper = await mount(<AlertRuleTemplate />);
    });

    expect(wrapper.find('textarea')).toBeTruthy();
    expect(wrapper.contains(dataQa('modal-wrapper'))).toBeFalsy();

    wrapper
      .find(dataQa('alert-rule-template-add-modal-button'))
      .find('button')
      .simulate('click');

    expect(wrapper.find(dataQa('modal-wrapper'))).toBeTruthy();
  });

  it('should render table content', async () => {
    let wrapper: ReactWrapper;

    await act(async () => {
      wrapper = await mount(<AlertRuleTemplate />);
    });

    wrapper.update();

    expect(wrapper.find(dataQa('table-thead')).find('tr')).toHaveLength(1);
    expect(wrapper.find(dataQa('table-tbody')).find('tr')).toHaveLength(3);
    expect(wrapper.find(dataQa('table-no-data'))).toHaveLength(0);
  });

  it('should render correctly without data', async () => {
    jest.spyOn(AlertRuleTemplateService, 'list').mockImplementation(() => {
      throw Error('test error');
    });

    let wrapper: ReactWrapper;

    await act(async () => {
      wrapper = await mount(<AlertRuleTemplate />);
    });

    wrapper.update();

    expect(wrapper.find(dataQa('table-thead')).find('tr')).toHaveLength(0);
    expect(wrapper.find(dataQa('table-tbody')).find('tr')).toHaveLength(0);
    expect(wrapper.find(dataQa('table-no-data'))).toHaveLength(1);
  });

  it('should have table initially loading', async () => {
    let wrapper: ReactWrapper;

    await act(async () => {
      wrapper = await mount(<AlertRuleTemplate />);
    });

    expect(wrapper.find(dataQa('table-loading'))).toHaveLength(1);
    expect(wrapper.find(dataQa('table-thead')).find('tr')).toHaveLength(0);
    expect(wrapper.find(dataQa('table-no-data'))).toHaveLength(0);
  });
});
