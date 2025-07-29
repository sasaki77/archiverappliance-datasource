import { SelectableValue } from '@grafana/data';
import { ComboboxOption } from '@grafana/ui';

export const toComboboxOption = (value: string): ComboboxOption => ({
  label: value,
  value,
});

export function toSelectableValue<T extends string>(t: T): SelectableValue<T> {
  return { label: t, value: t };
}
