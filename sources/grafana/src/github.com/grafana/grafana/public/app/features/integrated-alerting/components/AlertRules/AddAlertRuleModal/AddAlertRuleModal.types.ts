import { SelectableValue } from '@grafana/data';
import { AlertRule } from '../AlertRules.types';

export interface AddAlertRuleModalProps {
  isVisible: boolean;
  setVisible: (value: boolean) => void;
  alertRule?: AlertRule | null;
}

export enum Severity {
  SEVERITY_CRITICAL = 'SEVERITY_CRITICAL',
  SEVERITY_ERROR = 'SEVERITY_ERROR',
  SEVERITY_WARNING = 'SEVERITY_WARNING',
  SEVERITY_NOTICE = 'SEVERITY_NOTICE',
}

export interface AddAlertRuleFormValues {
  template: SelectableValue<string>;
  name: string;
  threshold: string | null;
  duration: number;
  filters: string;
  notificationChannels: Array<SelectableValue<string>>;
  severity: SelectableValue<Severity>;
  enabled: boolean;
}
