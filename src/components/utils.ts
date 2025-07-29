import { ComboboxOption } from '@grafana/ui';

export const toComboboxOption = (value: string): ComboboxOption => ({
  label: value,
  value,
});
