// Wholesale workflow states. MUST mirror backend/pkg/constants/wholesalestate (gaps leave room for new states).
export const WholesaleStates = {
  OnRequest: 10,
  Connected: 20,
  Blocked: 30,
  Declined: 40,
} as const;

export type WholesaleState = (typeof WholesaleStates)[keyof typeof WholesaleStates];

type WholesaleStateInfo = {
  value: WholesaleState;
  label: string;
  color: string;
};

export const WHOLESALE_STATE_INFO: readonly WholesaleStateInfo[] = [
  { value: WholesaleStates.OnRequest, label: 'On request', color: 'gold' },
  { value: WholesaleStates.Connected, label: 'Connected', color: 'green' },
  { value: WholesaleStates.Blocked, label: 'Blocked', color: 'red' },
  { value: WholesaleStates.Declined, label: 'Declined', color: 'default' },
];

// Unknown values (a state added on the backend first) still render.
export function wholesaleStateInfo(state: number): { label: string; color: string } {
  return WHOLESALE_STATE_INFO.find((s) => s.value === state) ?? { label: String(state), color: 'default' };
}
