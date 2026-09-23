import { screen } from '@testing-library/react';

interface TextScope {
  getAllByText: typeof screen.getAllByText;
}

export function detailGroupIds(): (string | null)[] {
  return screen.getAllByTestId(/^detail-group-/).map((element) => element.getAttribute('data-testid'));
}

export function fieldLabelsIn(scope: TextScope, labels: RegExp): (string | null)[] {
  return scope.getAllByText(labels, { selector: 'label' }).map((element) => element.textContent);
}
