interface Action {
  title: string;
  action: () => void;
  disabled?: boolean;
}

export interface MultipleActionsProps {
  actions: Action[];
  disabled?: boolean;
  dataQa?: string;
}
