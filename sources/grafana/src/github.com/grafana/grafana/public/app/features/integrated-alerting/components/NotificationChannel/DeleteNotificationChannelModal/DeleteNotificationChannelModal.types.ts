import { NotificationChannel } from '../NotificationChannel.types';

export interface DeleteNotificationChannelModalProps {
  isVisible: boolean;
  notificationChannel?: NotificationChannel | null;
  setVisible: (value: boolean) => void;
}
