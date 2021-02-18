import React, { FC, useContext } from 'react';
import { withTypes, Field } from 'react-final-form';
import { HorizontalGroup, Select, Button, useStyles } from '@grafana/ui';
import { AppEvents } from '@grafana/data';
import { Modal, LoaderButton, TextInputField, validators, logger } from '@percona/platform-core';
import { appEvents } from 'app/core/core';
import { NotificationChannelProvider } from '../NotificationChannel.provider';
import {
  NotificationChannelRenderProps,
  NotificationChannelType,
  PagerDutyKeyType,
} from '../NotificationChannel.types';
import { AddNotificationChannelModalProps } from './AddNotificationChannelModal.types';
import { getStyles } from './AddNotificationChannelModal.styles';
import { Messages } from './AddNotificationChannelModal.messages';
import { TYPE_OPTIONS } from './AddNotificationChannel.constants';
import { NotificationChannelService } from '../NotificationChannel.service';
import { getInitialValues } from './AddNotificationChannelModal.utils';
import { EmailFields } from './EmailFields/EmailFields';
import { SlackFields } from './SlackFields/SlackFields';
import { PagerDutyFields } from './PagerDutyFields/PagerDutyFields';

const { required } = validators;
// Our "values" typings won't be right without using this
const { Form } = withTypes<NotificationChannelRenderProps>();

const TypeField: FC<{ values: NotificationChannelRenderProps }> = ({ values }) => {
  const { type } = values;

  switch (type?.value) {
    case NotificationChannelType.email:
      return <EmailFields />;
    case NotificationChannelType.pagerDuty:
      return <PagerDutyFields values={values} />;
    case NotificationChannelType.slack:
      return <SlackFields />;
    default:
      return null;
  }
};

export const AddNotificationChannelModal: FC<AddNotificationChannelModalProps> = ({
  isVisible,
  notificationChannel,
  setVisible,
}) => {
  const styles = useStyles(getStyles);
  const initialValues = getInitialValues(notificationChannel);
  const { getNotificationChannels } = useContext(NotificationChannelProvider);
  const onSubmit = async (values: NotificationChannelRenderProps) => {
    const submittedValues = { ...values };

    if (submittedValues.keyType === PagerDutyKeyType.routing) {
      submittedValues.service = '';
    } else {
      submittedValues.routing = '';
    }

    try {
      if (notificationChannel?.channelId) {
        await NotificationChannelService.change(notificationChannel.channelId, submittedValues);
      } else {
        await NotificationChannelService.add(submittedValues);
      }
      setVisible(false);
      appEvents.emit(AppEvents.alertSuccess, [notificationChannel ? Messages.editSuccess : Messages.addSuccess]);
      getNotificationChannels();
    } catch (e) {
      logger.error(e);
    }
  };

  return (
    <Modal title={Messages.title} isVisible={isVisible} onClose={() => setVisible(false)}>
      <Form
        initialValues={initialValues}
        onSubmit={onSubmit}
        render={({ handleSubmit, valid, pristine, submitting, values }) => (
          <form onSubmit={handleSubmit}>
            <>
              <TextInputField name="name" label={Messages.fields.name} validators={[required]} />
              <Field name="type">
                {({ input }) => (
                  <>
                    <label className={styles.label} data-qa="type-field-label">
                      {Messages.fields.type}
                    </label>
                    <Select className={styles.select} options={TYPE_OPTIONS} {...input} />
                  </>
                )}
              </Field>
              <TypeField values={values} />
              <HorizontalGroup justify="center" spacing="md">
                <LoaderButton
                  data-qa="notification-channel-add-button"
                  // TODO: fix LoaderButton types
                  // @ts-ignore
                  size="md"
                  variant="primary"
                  disabled={!valid || pristine}
                  loading={submitting}
                >
                  {notificationChannel ? Messages.editAction : Messages.addAction}
                </LoaderButton>
                <Button
                  data-qa="notification-channel-cancel-button"
                  variant="secondary"
                  onClick={() => setVisible(false)}
                >
                  {Messages.cancelAction}
                </Button>
              </HorizontalGroup>
            </>
          </form>
        )}
      />
    </Modal>
  );
};
